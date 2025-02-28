package mediationcontainer

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/golang/glog"
	"github.ibm.com/turbonomic/turbo-api/pkg/client"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/version"
)

var (
	defaultRemoteMediationServerEndpoints = []WebSocketEndpoint{
		{client.API, "/vmturbo/remoteMediation"},
		{client.TopologyProcessor, "/remoteMediation"},
	}

	defaultRemoteMediationServerUser   = "vmtRemoteMediation"
	defaultRemoteMediationServerPwd    = "vmtRemoteMediation"
	defaultRemoteMediationLocalAddress = "http://127.0.0.1"
)

const (
	DefaultRegistrationTimeOut          = 300
	DefaultRegistrationTimeoutThreshold = 60
	DefaultChunkSendDelayMillis         = 0
	MaxChunkSendDelayMillis             = 60 * 1000 // 1 min
	DefaultNumObjectsPerChunk           = 5000
	MaxNumObjectsPerChunk               = 100000
)

type ServerMeta struct {
	TurboServer  string `json:"turboServer,omitempty"`
	Version      string `json:"version,omitempty"`
	Proxy        string `json:"proxy,omitempty"`
	ClientId     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

type WebSocketEndpoint struct {
	Service  string
	Endpoint string
}

func (meta *ServerMeta) ValidateServerMeta() error {
	if meta.TurboServer == "" {
		return errors.New("Turbo Server URL is missing")
	}
	if _, err := url.ParseRequestURI(meta.TurboServer); err != nil {
		return fmt.Errorf("Invalid turbo address url: %v", meta)
	}
	if meta.Version == "" {
		meta.Version = string(version.PROTOBUF_VERSION)
	}
	// User should either specify neither OAuth2 clientid nor clientsecret, or
	// specify both clientid and clientsecret
	if (meta.ClientId == "" && meta.ClientSecret != "") ||
		(meta.ClientId != "" && meta.ClientSecret == "") {
		return errors.New("both OAuth2 clientid and clientsecret must be provided")
	}
	return nil
}

type WebSocketConfig struct {
	LocalAddress       string `json:"localAddress,omitempty"`
	WebSocketUsername  string `json:"websocketUsername,omitempty"`
	WebSocketPassword  string `json:"websocketPassword,omitempty"`
	ConnectionRetry    int16  `json:"connectionRetry,omitempty"`
	WebSocketEndpoints []WebSocketEndpoint
}

func (wsc *WebSocketConfig) ValidateWebSocketConfig() error {
	if wsc.LocalAddress == "" {
		wsc.LocalAddress = defaultRemoteMediationLocalAddress
	}
	// Make sure the local address string provided is a valid URL
	if _, err := url.ParseRequestURI(wsc.LocalAddress); err != nil {
		return fmt.Errorf("invalid local address url found in WebSocket config: %v", wsc)
	}

	wsc.WebSocketEndpoints = defaultRemoteMediationServerEndpoints

	if wsc.WebSocketUsername == "" {
		wsc.WebSocketUsername = defaultRemoteMediationServerUser
	}
	if wsc.WebSocketPassword == "" {
		wsc.WebSocketPassword = defaultRemoteMediationServerPwd
	}
	return nil
}

// Configuration options used when establishing sdk protocol connection with the server
type SdkProtocolConfig struct {
	// Probe registration response timeout
	RegistrationTimeoutSec int `json:"registrationTimeoutSec,omitempty"`
	// If the probe container should exit if there is timeout during probe registration
	RestartOnRegistrationTimeout bool `json:"restartOnRegistrationTimeout,omitempty"`
}

func (sdkProtocolConfig *SdkProtocolConfig) ValidateSdkProtocolConfig() error {
	glog.Infof("SdkProtocolConfig from config file [%++v]", sdkProtocolConfig)

	// Default to 300 seconds if the timeout is less than 60 seconds
	if sdkProtocolConfig.RegistrationTimeoutSec < DefaultRegistrationTimeoutThreshold {
		glog.Warningf("Changing invalid 'RegistrationTimeoutSec' config [%v] to default %v",
			sdkProtocolConfig.RegistrationTimeoutSec, DefaultRegistrationTimeOut)
		sdkProtocolConfig.RegistrationTimeoutSec = DefaultRegistrationTimeOut
	}

	glog.Infof("Validated SdkProtocolConfig [%++v]: ", sdkProtocolConfig)

	return nil
}

type MediationContainerConfig struct {
	ServerMeta
	WebSocketConfig
	CommunicationBindingChannel string
	ChunkSendDelayMillis        int
	NumObjectsPerChunk          int
	SdkProtocolConfig
}

// Validate the mediation container config and set default value if necessary.
func (containerConfig *MediationContainerConfig) ValidateMediationContainerConfig() error {
	if err := containerConfig.ValidateServerMeta(); err != nil {
		return err
	}
	if err := containerConfig.ValidateWebSocketConfig(); err != nil {
		return err
	}
	if err := containerConfig.ValidateSdkProtocolConfig(); err != nil {
		return err
	}
	// DIF and Prometurbo probes aren't currently sending this value, so replace with the default
	if containerConfig.NumObjectsPerChunk == 0 {
		containerConfig.NumObjectsPerChunk = DefaultNumObjectsPerChunk
	}
	glog.V(4).Infof("The mediation container config is %v", containerConfig)
	return nil
}
