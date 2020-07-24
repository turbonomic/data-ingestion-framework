package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/service"
)

const (
	// Config file paths as volume mounts for the service pod
	DefaultConfPath            = "/etc/turbodif/turbodif-config.json"
	DefaultSupplyChainConfPath = "/opt/turbonomic/conf/app-supply-chain-config.yaml"

	// Debug file paths
	LocalDebugConfPath            = "configs/turbodif-config.json"
	LocalDebugSupplyChainConfPath = "configs/app-supply-chain-config.yaml"

	DefaultProbeCategory   string = "Guest OS Processes"
	DefaultProbeUICategory string = "Applications and Databases"
	DefaultTargetType      string = "DataIngestionFramework"
)

// Configuration for the TurboDIF probe
type DIFConf struct {
	// configuration for connecting to the Turbo server
	Communicator *service.TurboCommunicationConfig `json:"communicationConfig,omitempty"`
	// configuration for the DIF Probe target
	TargetConf *DIFTargetConf `json:"targetConfig,omitempty"`
	// Appended to the end of a probe name when registering with the platform. Useful when you need
	// multiple prometheus probe instances with affinity for discovering specific targets.
	TargetTypeSuffix string `json:"targetTypeSuffix,omitempty"`
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
