package main

import (
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/data-ingestion-framework/pkg"
	"github.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.com/turbonomic/data-ingestion-framework/version"
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

	// Print out critical information
	glog.Infof("Running turbodif VERSION: %s, GIT_COMMIT: %s, BUILD_TIME: %s",
		version.Version, version.GitCommit, version.BuildTime)
	glog.Infof("IgnoreIfPresent is set to %v for all commodities.", *args.IgnoreCommodityIfPresent)

	s, err := pkg.NewDIFTAPService(args)
	if err != nil {
		glog.Fatalf("Failed to run turbodif: %v", err)
	}

	s.Start()

	return
}
