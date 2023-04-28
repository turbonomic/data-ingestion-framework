package mediationcontainer

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	Closed           TransportStatus = "closed"
	Ready            TransportStatus = "ready"
	handshakeTimeout                 = 60 * time.Second
	wsReadLimit                      = 33554432 // 32 MB
	writeWaitTimeout                 = 120 * time.Second
	pingPeriod                       = 30 * time.Second
)

type TransportStatus string

type WebSocketConnectionConfig MediationContainerConfig

func CreateWebSocketConnectionConfig(connConfig *MediationContainerConfig) (*WebSocketConnectionConfig, error) {
	_, err := url.ParseRequestURI(connConfig.LocalAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse local URL from %s when create WebSocketConnectionConfig.", connConfig.LocalAddress)
	}
	// Change URL scheme from ws to http or wss to https.
	serverURL, err := url.ParseRequestURI(connConfig.TurboServer)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse turboServer URL from %s when create WebSocketConnectionConfig.", connConfig.TurboServer)
	}
	switch serverURL.Scheme {
	case "http":
		serverURL.Scheme = "ws"
	case "https":
		serverURL.Scheme = "wss"
	}
	wsConfig := WebSocketConnectionConfig(*connConfig)
	wsConfig.TurboServer = serverURL.String()
	return &wsConfig, nil
}

// ============================ ClientWebSocketTransport - Start, Connect, Close ======================================
// Implementation of the ITransport for WebSocket communication to send and receive serialized protobuf message bytes
type ClientWebSocketTransport struct {
	status                   TransportStatus // current status of the transport layer.
	wsMux                    sync.Mutex      // protect ws from concurrent writing (ws.WriteMessage())
	ws                       *websocket.Conn // created during Connect()
	connConfig               *WebSocketConnectionConfig
	inputStreamCh            chan []byte // unbuffered channel
	closeRequested           bool
	stopListenerCh           chan bool // buffered channel
	connClosedNotificationCh chan bool // channel where the transport connection error will be notified
	service                  string
}

// Instantiate a new ClientWebSocketTransport endpoint for the client
func CreateClientWebSocketTransport(connConfig *WebSocketConnectionConfig) *ClientWebSocketTransport {
	transport := &ClientWebSocketTransport{
		connConfig:               connConfig,
		connClosedNotificationCh: make(chan bool),
	}
	return transport
}

// WebSocket connection is established with the server
func (wsTransport *ClientWebSocketTransport) Connect(refreshTokenChannel chan struct{}, jwTokenChannel chan string) error {
	// Close any previous connected WebSocket connection and set current connection to nil.
	wsTransport.closeAndResetWebSocket()

	// loop till server is up or close received
	wsTransport.closeRequested = false
	// TODO: give an optional timeout to wait for server in performWebSocketConnection()
	err := wsTransport.performWebSocketConnection(refreshTokenChannel, jwTokenChannel) // Blocks or till transport is closed
	if err != nil {
		return err
	}

	glog.V(4).Infof("[Connect] Connected to server " + wsTransport.GetConnectionId())

	wsTransport.stopListenerCh = make(chan bool, 1) // Channel to stop the routine that listens for messages
	wsTransport.inputStreamCh = make(chan []byte)   // Message Queue
	// Message handler for received messages
	go wsTransport.ListenForMessages()
	go wsTransport.startPing()
	return nil
}

func (wsTransport *ClientWebSocketTransport) NotifyClosed() chan bool {
	return wsTransport.connClosedNotificationCh
}

func (wsTransport *ClientWebSocketTransport) GetService() string {
	return wsTransport.service
}

func (wsTransport *ClientWebSocketTransport) GetConnectionId() string {
	if wsTransport.status == Closed {
		return ""
	}
	wsTransport.wsMux.Lock()
	defer wsTransport.wsMux.Unlock()
	if wsTransport.ws == nil {
		return ""
	}
	return wsTransport.ws.RemoteAddr().String() + "::" + wsTransport.ws.LocalAddr().String()
}

// Close the WebSocket Transport point: this is called by upper module (remoteMediationClient)
func (wsTransport *ClientWebSocketTransport) CloseTransportPoint() {
	wsTransport.closeRequested = true
}

// Close current WebSocket connection and set it to nil.
func (wsTransport *ClientWebSocketTransport) closeAndResetWebSocket() {
	if wsTransport.status == Closed {
		return
	}

	wsTransport.wsMux.Lock()
	defer wsTransport.wsMux.Unlock()

	// close WebSocket
	if wsTransport.ws != nil {
		glog.V(1).Infof("Begin to send websocket Close frame.")
		wsTransport.ws.SetWriteDeadline(time.Now().Add(writeWaitTimeout))
		wsTransport.ws.WriteMessage(websocket.CloseMessage, []byte{})
		wsTransport.ws.Close()
		wsTransport.ws = nil
	}
	wsTransport.status = Closed
}

func (wsTransport *ClientWebSocketTransport) write(mtype int, payload []byte) error {
	wsTransport.wsMux.Lock()
	defer wsTransport.wsMux.Unlock()
	wsTransport.ws.SetWriteDeadline(time.Now().Add(writeWaitTimeout))
	return wsTransport.ws.WriteMessage(mtype, payload)
}

// keep sending Ping msg to make sure the websocket connection is alive
// If don't send Ping msg, *some times* the ws.ReadMessage() won't be able to
//
//	know that the connection has gone.
func (wsTransport *ClientWebSocketTransport) startPing() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-wsTransport.stopListenerCh:
			return
		case <-ticker.C:
			glog.V(3).Infof("begin to send Ping message")
			if err := wsTransport.write(websocket.PingMessage, []byte{}); err != nil {
				glog.Errorf("Failed to send PingMessage to server:%v", err)
				return
			}
			glog.V(4).Infof("Sent ping message success")
		}
	}
}

// ================================================= Message Listener =============================================
// TODO: avoid close a closed channel
func (wsTransport *ClientWebSocketTransport) stopListenForMessages() {
	if wsTransport.stopListenerCh != nil {
		glog.V(4).Infof("[StopListenForMessages] closing stopListenerCh %+v", wsTransport.stopListenerCh)
		close(wsTransport.stopListenerCh)
		glog.V(4).Infof("[StopListenForMessages] closed stopListenerCh %+v", wsTransport.stopListenerCh)
	}
}

// Routine to listen for messages on the websocket.
// The websocket is continuously checked for messages and queued on the clientTransport.inputStream channel
// Routine exits when a message is sent on clientTransport.stopListenerCh.
func (wsTransport *ClientWebSocketTransport) ListenForMessages() {
	glog.V(3).Infof("[ListenForMessages]: ENTER  ")
	defer close(wsTransport.inputStreamCh) //notify the receiver that websocket stop feeding data

	for {
		glog.V(4).Info("[ListenForMessages] waiting for messages on websocket transport")
		glog.V(4).Infof("[ListenForMessages] waiting for messages on websocket transport : %++v", wsTransport)
		select {
		case <-wsTransport.stopListenerCh:
			glog.V(1).Info("[ListenForMessages] stop listening for message")
			return
		default:
			if wsTransport.status != Ready {
				glog.Errorf("WebSocket transport layer status is %s", wsTransport.status)
				glog.Errorf("WebSocket is not ready.")
				continue
			}
			glog.V(3).Infof("[ListenForMessages]: connected, waiting for server response ...")

			msgType, data, err := wsTransport.ws.ReadMessage()
			glog.V(3).Infof("Received websocket message of type %d and size %d", msgType, len(data))

			if wsTransport.closeRequested {
				glog.V(1).Infof("stop listening for message because of requested")
				return
			}

			// Notify errors and break
			if err != nil {
				// destroy its websocket connection whenever there is an error
				if err == io.EOF {
					glog.Errorf("[ListenForMessages] received EOF on websocket %s", err)
				}

				glog.Errorf("[ListenForMessages] error during receive %v", err)
				// close current WebSocket connection.
				wsTransport.closeAndResetWebSocket()
				wsTransport.stopListenForMessages()

				//notify upper module that this connection is closed
				wsTransport.connClosedNotificationCh <- true // Note: this will block till the message is received

				glog.V(1).Infof("[ListenForMessages] websocket error notified, stop listening for messages.")
				return
			}
			// write the message on the channel
			glog.V(3).Infof("[ListenForMessages] received message on websocket of size %d", len(data))

			wsTransport.queueRawMessage(data) // Note: this will block till the message is read
			glog.V(4).Infof("[ListenForMessages] delivered websocket message, continue listening for server messages...")
		} //end select
	} //end for
}

func (wsTransport *ClientWebSocketTransport) queueRawMessage(data []byte) {
	//TODO: this read should be accumulative - see onMessageReceived in AbstractWebsocketTransport
	wsTransport.inputStreamCh <- data
	// TODO: how to ensure that the channel is open to write
}

func (wsTransport *ClientWebSocketTransport) RawMessageReceiver() chan []byte {
	return wsTransport.inputStreamCh
}

// ==================================================== Message Sender ===============================================
// Send serialized protobuf message bytes
func (wsTransport *ClientWebSocketTransport) Send(messageToSend *TransportMessage) error {
	if wsTransport.closeRequested {
		glog.Errorf("Cannot send message : transport endpoint is closed")
		return errors.New("Cannot send message: transport endpoint is closed")
	}

	if wsTransport.status != Ready {
		glog.Errorf("WebSocket transport layer status is %s", wsTransport.status)
		return errors.New("Cannot send message: web socket is not ready")
	}

	if messageToSend == nil { //.RawMsg == nil {
		glog.Errorf("Cannot send message : marshalled msg is nil")
		return errors.New("Cannot send message: marshalled msg is nil")
	}

	err := wsTransport.write(websocket.BinaryMessage, messageToSend.RawMsg)
	if err != nil {
		glog.Errorf("Error sending message on client transport: %s", err)
		return fmt.Errorf("Error sending message on client transport: %s", err)
	}
	glog.V(4).Infof("Successfully sent message on client transport")
	return nil
}

// ====================================== Websocket Connection =========================================================
// Establish connection to server websocket until connected or until the transport endpoint is closed
func (wsTransport *ClientWebSocketTransport) performWebSocketConnection(refreshTokenChannel chan struct{}, jwTokenChannel chan string) error {
	connRetryInterval := time.Second * 30 // TODO: use ConnectionRetry parameter from the connConfig or default
	connConfig := wsTransport.connConfig

	for !wsTransport.closeRequested { // only set when CloseTransportPoint() is called
		// blocked for jwtToken
		refreshTokenChannel <- struct{}{}
		jwtToken := <-jwTokenChannel
		if len(jwtToken) > 0 {
			glog.Infof("Trying to establish secure websocket connection...")
		}
		ws, service, err := openWebSocketConn(connConfig, jwtToken)
		if err != nil {
			// print at debug level after some time
			glog.V(3).Infof("Failed to open websocket connection: %v. Retry in %v.",
				err, connRetryInterval)
			time.Sleep(connRetryInterval)
		} else {
			setupPingPong(ws)
			wsTransport.wsMux.Lock()
			wsTransport.ws = ws
			wsTransport.wsMux.Unlock()
			wsTransport.status = Ready
			wsTransport.service = service
			glog.V(2).Infof("Connected to server " + wsTransport.GetConnectionId())
			glog.V(2).Infof("WebSocket transport layer is ready.")

			return nil
		}
	}
	glog.V(4).Infof("Exit websocket connection routine, close = %v ", wsTransport.closeRequested)
	return errors.New("abort client websocket connection, transport is closed")
}

// set up websocket Ping-Pong protocol handlers
func setupPingPong(ws *websocket.Conn) {
	h := func(message string) error {
		glog.V(3).Infof("Recevied ping msg")
		err := ws.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(writeWaitTimeout))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}

		if err != nil {
			glog.Errorf("Failed to send PongMessage: %v", err)
		}
		return err
	}

	ws.SetPingHandler(h)

	h2 := func(message string) error {
		glog.V(3).Infof("Received pong msg")
		return nil
	}
	ws.SetPongHandler(h2)
	return
}

func openWebSocketConn(connConfig *WebSocketConnectionConfig, jwtToken string) (*websocket.Conn, string, error) {
	//1. set up dialer
	d := &websocket.Dialer{
		HandshakeTimeout: handshakeTimeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
	}
	proxy := connConfig.Proxy
	if proxy != "" {
		//Check if the proxy server requires authentication or not
		//Authenticated proxy format: http://username:password@ip:port
		//Non-Aunthenticated proxy format: http://ip:port
		if strings.Index(proxy, "@") != -1 {
			//Extract the username password portion, with @
			username_password := proxy[strings.Index(proxy, "//")+2 : strings.LastIndex(proxy, "@")+1]
			username := username_password[:strings.Index(username_password, ":")]
			password := username_password[strings.Index(username_password, ":")+1 : strings.LastIndex(username_password, "@")]
			//Extract Proxy address by remove the username_password
			proxyAddr := strings.ReplaceAll(proxy, username_password, "")
			proxyURL, err := url.Parse(proxyAddr)
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse proxy")
			}
			proxyURL.User = url.UserPassword(username, password)
			d.Proxy = http.ProxyURL(proxyURL)
		} else {
			proxyURL, err := url.Parse(proxy)
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse proxy")
			}
			d.Proxy = http.ProxyURL(proxyURL)
		}
	}
	//2. auth header
	header := getAuthHeader(connConfig.WebSocketUsername, connConfig.WebSocketPassword, jwtToken)

	//3. connect it
	for service, endpoint := range connConfig.WebSocketEndpoints {
		// WebSocket URL
		vmtServerUrl := connConfig.TurboServer + endpoint
		glog.Infof("Trying websocket connection to: %s", vmtServerUrl)
		c, _, err := d.Dial(vmtServerUrl, header)
		if err == nil {
			glog.Infof("Successfully connected to %v service at: %s",
				service, vmtServerUrl)
			c.SetReadLimit(wsReadLimit)
			return c, service, nil
		}
		glog.Warningf("Cannot connect to %s: %v", vmtServerUrl, err)
	}
	return nil, "", fmt.Errorf("failed to connect to all server options")
}

func getAuthHeader(user, password, jwtToken string) http.Header {
	dat := []byte(fmt.Sprintf("%s:%s", user, password))
	header := http.Header{
		"Authorization": {"Basic " + base64.StdEncoding.EncodeToString(dat)},
	}

	if jwtToken != "" {
		header.Add("x-auth-token", base64.StdEncoding.EncodeToString([]byte(jwtToken)))
	}

	return header
}
