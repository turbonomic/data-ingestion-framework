package main

import (
	"flag"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.ibm.com/turbonomic/data-ingestion-framework/version"

	"github.com/spf13/viper"
)

func main() {

	// Ignore errors
	_ = flag.Set("logtostderr", "false")
	_ = flag.Set("alsologtostderr", "true")
	_ = flag.Set("log_dir", "/var/log")
	defer glog.Flush()

	// Config pretty print for debugging
	spew.Config = spew.ConfigState{
		Indent:                "  ",
		MaxDepth:              0,
		DisableMethods:        true,
		DisablePointerMethods: true,
		ContinueOnMethod:      false,
		SortKeys:              true,
		SpewKeys:              false,
	}

	// Set up arguments specific to DIF
	args := conf.NewDIFProbeArgs()
	// Parse the flags
	flag.Parse()

	// Watch the configmap and detect the change on it
	go WatchConfigMap()

	// Print out critical information
	glog.Infof("Running turbodif VERSION: %s, GIT_COMMIT: %s, BUILD_TIME: %s",
		version.Version, version.GitCommit, version.BuildTime)
	glog.Infof("IgnoreIfPresent is set to %v for all commodities.", *args.IgnoreCommodityIfPresent)
	glog.Infof("The discovery interval in seconds is set to %d sec.", *args.DiscoveryIntervalSec)

	s, err := pkg.NewDIFTAPService(args)
	if err != nil {
		glog.Fatalf("Failed to run turbodif: %v", err)
	}

	s.Start()

	return
}

func WatchConfigMap() {
	//Check if the /etc/turbodif/turbo-autoreload.config exists
	autoReloadConfigFilePath := "/etc/turbodif"
	autoReloadConfigFileName := "turbo-autoreload.config"

	viper.AddConfigPath(autoReloadConfigFilePath)
	viper.SetConfigType("json")
	viper.SetConfigName(autoReloadConfigFileName)
	for {
		verr := viper.ReadInConfig()
		if verr == nil {
			break
		} else {
			glog.V(4).Infof("Can't read the autoreload config file %s/%s due to the error: %v, will retry in 3 seconds", autoReloadConfigFilePath, autoReloadConfigFileName, verr)
			time.Sleep(30 * time.Second)
		}
	}

	glog.V(1).Infof("Start watching the autoreload config file %s/%s", autoReloadConfigFilePath, autoReloadConfigFileName)
	updateConfig := func() {
		newLoggingLevel := viper.GetString("logging.level")
		currentLoggingLevel := flag.Lookup("v").Value.String()
		if newLoggingLevel != currentLoggingLevel {
			if newLogVInt, err := strconv.Atoi(newLoggingLevel); err != nil || newLogVInt < 0 {
				glog.Errorf("Invalid log verbosity %v in the autoreload config file", newLoggingLevel)
			} else {
				err := flag.Lookup("v").Value.Set(newLoggingLevel)
				if err != nil {
					glog.Errorf("Can't apply the new logging level setting due to the error:%v", err)
				} else {
					glog.V(1).Infof("Logging level is changed from %v to %v", currentLoggingLevel, newLoggingLevel)
				}
			}
		}
	}
	updateConfig() //update the logging level during startup
	viper.OnConfigChange(func(in fsnotify.Event) {
		updateConfig()
	})

	viper.WatchConfig()
}
