package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.ibm.com/turbonomic/data-ingestion-framework/version"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/service"
)

const (
	// Config file paths as volume mounts for the service pod
	DefaultConfPath            = "/etc/turbodif/turbodif-config.json"
	DefaultSupplyChainConfPath = "/opt/turbonomic/conf/app-supply-chain-config.yaml"

	// Debug file paths
	LocalDebugConfPath            = "configs/turbodif-config.json"
	LocalDebugSupplyChainConfPath = "configs/app-supply-chain-config.yaml"

	DefaultProbeCategory   = "Custom"
	DefaultProbeUICategory = "Custom"
	DefaultTargetType      = "DataIngestionFramework"

	credentialsDirPath   = "/etc/turbonomic-credentials"
	usernameFilePath     = "/etc/turbonomic-credentials/username"
	passwordFilePath     = "/etc/turbonomic-credentials/password"
	clientIdFilePath     = "/etc/turbonomic-credentials/clientid"
	clientSecretFilePath = "/etc/turbonomic-credentials/clientsecret"
)

// Configuration for the TurboDIF probe
type DIFConf struct {
	// configuration for connecting to the Turbo server
	Communicator *service.TurboCommunicationConfig `json:"communicationConfig,omitempty"`
	// configuration for the DIF Probe target
	TargetConf *DIFTargetConf `json:"targetConfig,omitempty"`
}

// Configuration for the TurboDIF target
type DIFTargetConf struct {
	Address string `json:"targetAddress,omitempty"`
	Name    string `json:"targetName,omitempty"`
}

func NewDIFConf(configFilePath string) (config *DIFConf, err error) {
	glog.Infof("Read TurboDIF probe configuration from %s", configFilePath)
	config, err = readConfig(configFilePath)
	if err != nil {
		return
	}
	if config.Communicator == nil {
		err = fmt.Errorf("unable to read the turbo communication config from %s", configFilePath)
		return
	}
	if config.TargetConf == nil {
		// Create an empty target config
		config.TargetConf = &DIFTargetConf{}
		return
	}
	if len(config.TargetConf.Address) > 0 && len(config.TargetConf.Name) == 0 {
		err = fmt.Errorf("unspecified name for target with address: %s", config.TargetConf.Address)
		return
	}
	if len(config.TargetConf.Name) > 0 && len(config.TargetConf.Address) == 0 {
		err = fmt.Errorf("unspecified address for target with name: %s", config.TargetConf.Name)
		return
	}

	// Read secure credentials
	if _, err := os.Stat(credentialsDirPath); os.IsNotExist(err) {
		glog.V(2).Infof("credentials mount path %s does not exist", credentialsDirPath)
	} else {
		if err := loadOpsMgrCredentialsFromSecret(config); err != nil {
			return nil, err
		}
		if err := loadClientIdSecretFromSecret(config); err != nil {
			return nil, err
		}
	}
	setTurboDifBuildVersion(config)
	return
}

func readConfig(path string) (*DIFConf, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("file error: %v", err)
	}
	glog.V(4).Info(string(file))
	var config DIFConf
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error :%v", err)
	}
	glog.V(4).Infof("Results: %+v", spew.Sdump(config))
	return &config, nil
}

func loadOpsMgrCredentialsFromSecret(config *DIFConf) error {
	// Return unchanged if the mounted file isn't present
	// for backward compatibility.
	if _, err := os.Stat(usernameFilePath); os.IsNotExist(err) {
		glog.V(2).Infof("server api credentials from secret unavailable. Checked path: %s", usernameFilePath)
		return nil
	}
	if _, err := os.Stat(passwordFilePath); os.IsNotExist(err) {
		glog.V(2).Infof("server api credentials from secret unavailable. Checked path: %s", passwordFilePath)
		return nil
	}

	username, err := os.ReadFile(usernameFilePath)
	if err != nil {
		return fmt.Errorf("error reading server api credentials from secret: username: %v", err)
	}
	password, err := os.ReadFile(passwordFilePath)
	if err != nil {
		return fmt.Errorf("error reading server api credentials from secret: password: %v", err)
	}

	config.Communicator.OpsManagerUsername = strings.TrimSpace(string(username))
	config.Communicator.OpsManagerPassword = strings.TrimSpace(string(password))

	return nil
}

func loadClientIdSecretFromSecret(config *DIFConf) error {
	// Return unchanged if the mounted file isn't present
	// for backward compatibility.
	if _, err := os.Stat(clientIdFilePath); os.IsNotExist(err) {
		glog.V(2).Infof("secure server credentials from secret unavailable. Checked path: %s", clientIdFilePath)
		return nil
	}
	if _, err := os.Stat(clientSecretFilePath); os.IsNotExist(err) {
		glog.V(2).Infof("secure server credentials from secret unavailable. Checked path: %s", clientSecretFilePath)
		return nil
	}

	clientId, err := os.ReadFile(clientIdFilePath)
	if err != nil {
		return fmt.Errorf("error reading secure server credentials from secret: clientId: %v", err)
	}
	clientSecret, err := os.ReadFile(clientSecretFilePath)
	if err != nil {
		return fmt.Errorf("error reading secure server credentials from secret: clientSecret: %v", err)
	}

	config.Communicator.ClientId = strings.TrimSpace(string(clientId))
	config.Communicator.ClientSecret = strings.TrimSpace(string(clientSecret))
	glog.V(4).Infof("Obtained credentials to set up secure probe communication")
	return nil
}

func setTurboDifBuildVersion(config *DIFConf) {
	if config.Communicator != nil && config.Communicator.Version == "" {
		config.Communicator.Version = version.Version
	}
}
