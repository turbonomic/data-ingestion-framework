package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/turbonomic/turbo-go-sdk/pkg/mediationcontainer"

	"github.com/golang/glog"
)

type RestAPIConfig struct {
	OpsManagerUsername string `json:"opsManagerUsername,omitempty"`
	OpsManagerPassword string `json:"opsManagerPassword,omitempty"`
	APIPath            string `json:"apiPath,omitempty"`
}

func (rc *RestAPIConfig) ValidRestAPIConfig() error {
	// User should either specify neither username nor password, or
	// specify both username and password
	if (rc.OpsManagerUsername == "" && rc.OpsManagerPassword != "") ||
		(rc.OpsManagerUsername != "" && rc.OpsManagerPassword == "") {
		return errors.New("both username and password must be provided")
	}
	return nil
}

// Configuration parameters for communicating with the Turbo server
type TurboCommunicationConfig struct {
	mediationcontainer.ServerMeta        `json:"serverMeta,omitempty"`
	mediationcontainer.WebSocketConfig   `json:"websocketConfig,omitempty"`
	RestAPIConfig                        `json:"restAPIConfig,omitempty"`
	mediationcontainer.SdkProtocolConfig `json:"sdkProtocolConfig,omitempty"`
}

func (turboCommConfig *TurboCommunicationConfig) ValidateTurboCommunicationConfig() error {
	// validate the config
	if err := turboCommConfig.ValidateServerMeta(); err != nil {
		return err
	}
	if err := turboCommConfig.ValidateWebSocketConfig(); err != nil {
		return err
	}
	if err := turboCommConfig.ValidRestAPIConfig(); err != nil {
		return err
	}
	if err := turboCommConfig.ValidateSdkProtocolConfig(); err != nil {
		return err
	}
	return nil
}

func (turboCommConfig *TurboCommunicationConfig) SecureModeCredentialsProvided() bool {
	if len(turboCommConfig.ClientSecret) > 0 && len(turboCommConfig.ClientId) > 0 {
		return true
	}
	return false
}

func (turboCommConfig *TurboCommunicationConfig) TurboAPICredentialsProvided() bool {
	if len(turboCommConfig.OpsManagerUsername) > 0 && len(turboCommConfig.OpsManagerPassword) > 0 {
		return true
	}
	return false
}

func ParseTurboCommunicationConfig(configFile string) (*TurboCommunicationConfig, error) {
	// load the config
	turboCommConfig, err := readTurboCommunicationConfig(configFile)
	if turboCommConfig == nil {
		return nil, err
	}
	glog.V(3).Infof("TurboCommunicationConfig Config: %v", turboCommConfig)
	glog.Infof("TurboCommunicationConfig Config: %v", turboCommConfig)
	if err := turboCommConfig.ValidateTurboCommunicationConfig(); err != nil {
		return nil, err
	}
	glog.V(3).Info("Turbo communication config validation passed.")
	return turboCommConfig, nil
}

func readTurboCommunicationConfig(path string) (*TurboCommunicationConfig, error) {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		return nil, fmt.Errorf("File error: %v\n" + e.Error())
	}
	var config TurboCommunicationConfig
	err := json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("Unmarshall error :%v", err.Error())
	}
	return &config, nil
}
