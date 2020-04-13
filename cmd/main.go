package main

import (
	"flag"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/data-ingestion-framework/pkg"
	"github.com/turbonomic/data-ingestion-framework/pkg/conf"

	"os"
)

func parseFlags() {
	flag.Parse()
}

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

	// Parse command line flags
	//parseFlags()

	glog.Info("Starting DIF Turbo...")
	glog.Infof("GIT_COMMIT: %s", os.Getenv("GIT_COMMIT"))

	args := conf.NewDIFProbeArgs(flag.CommandLine)
	flag.Parse()

	s, err := pkg.NewDIFTAPService(args)

	if err != nil {
		glog.Fatalf("Failed creating DIFTurbo: %v", err)
	}

	s.Start()

	return
}
