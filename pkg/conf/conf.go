package conf

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"os"

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

func NewDIFConf(configFilePath string) (*DIFConf, error) {

	glog.Infof("Read TurboDIF probe configuration from %s", configFilePath)
	config, err := readConfig(configFilePath)

	if err != nil {
		return nil, err
	}

	if config.Communicator == nil {
		return nil, fmt.Errorf("unable to read the turbo communication config from %s", configFilePath)
	}

	if config.TargetConf == nil {
		return nil, fmt.Errorf("unable to read the turbo target config from %s", configFilePath)
	}

	if len(config.TargetConf.Address) > 0 && len(config.TargetConf.Name) == 0 {
		glog.Errorf("unspecified name for target with addressr: %s", config.TargetConf.Address)
		os.Exit(1)
	}

	if len(config.TargetConf.Name) > 0 && len(config.TargetConf.Address) == 0 {
		glog.Errorf("unspecified address for target with name: %s", config.TargetConf.Name)
		os.Exit(1)
	}

	return config, nil
}

func readConfig(path string) (*DIFConf, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorf("File error: %v", err)
		return nil, err
	}
	glog.V(4).Info(string(file))

	var config DIFConf
	err = json.Unmarshal(file, &config)

	if err != nil {
		glog.Errorf("Unmarshal error :%v", err)
		return nil, err
	}
	glog.V(4).Infof("Results: %+v", spew.Sdump(config))

	return &config, nil
}
