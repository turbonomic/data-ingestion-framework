package mediationcontainer

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/golang/glog"
	"github.com/turbonomic/turbo-api/pkg/client"
	"github.com/turbonomic/turbo-go-sdk/pkg/version"
)

var (
	defaultRemoteMediationServerEndpoints = map[string]string{
		client.API:               "/vmturbo/remoteMediation",
		client.TopologyProcessor: "/remoteMediation",
	}
	defaultRemoteMediationServerUser   = "vmtRemoteMediation"
	defaultRemoteMediationServerPwd    = "vmtRemoteMediation"
	defaultRemoteMediationLocalAddress = "http://127.0.0.1"
)

type ServerMeta struct {
	TurboServer string `json:"turboServer,omitempty"`
	Version     string `json:"version,omitempty"`
	Proxy       string `json:"proxy,omitempty"`
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
	return nil
}

type WebSocketConfig struct {
	LocalAddress       string `json:"localAddress,omitempty"`
	WebSocketUsername  string `json:"websocketUsername,omitempty"`
	WebSocketPassword  string `json:"websocketPassword,omitempty"`
	ConnectionRetry    int16  `json:"connectionRetry,omitempty"`
	WebSocketEndpoints map[string]string
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

type MediationContainerConfig struct {
	ServerMeta
	WebSocketConfig
}

// Validate the mediation container config and set default value if necessary.
func (containerConfig *MediationContainerConfig) ValidateMediationContainerConfig() error {
	if err := containerConfig.ValidateServerMeta(); err != nil {
		return err
	}
	if err := containerConfig.ValidateWebSocketConfig(); err != nil {
		return err
	}
	glog.V(4).Infof("The mediation container config is %v", containerConfig)
	return nil
}
