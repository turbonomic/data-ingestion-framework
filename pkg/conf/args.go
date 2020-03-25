package conf

import (
	"flag"
)

const (
	defaultDiscoveryIntervalSec = 600
	defaultKeepStandalone       = false
)

// Arguments for the TurboDIF Probe
type DIFProbeArgs struct {
	// The discovery interval in seconds for running the probe
	DiscoveryIntervalSec *int
	// Boolean to indicate if the the discovered entities that are not stitched should be deleted or kept in the server
	KeepStandalone *bool
}

func NewDIFProbeArgs(fs *flag.FlagSet) *DIFProbeArgs {
	p := &DIFProbeArgs{}

	p.DiscoveryIntervalSec = fs.Int("discovery-interval-sec", defaultDiscoveryIntervalSec, "The discovery interval in seconds")
	p.KeepStandalone = fs.Bool("keepStandalone", defaultKeepStandalone, "Do we keep non-stitched entities")

	return p
}
