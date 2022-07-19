package conf

import (
	"flag"
)

const (
	defaultDiscoveryIntervalSec     = 600
	defaultKeepStandalone           = true
	defaultIgnoreCommodityIfPresent = false
)

// DIFProbeArgs specifies arguments for the TurboDIF Probe
type DIFProbeArgs struct {
	// The discovery interval in seconds for running the probe
	DiscoveryIntervalSec *int
	// Boolean to indicate if the discovered entities that are not stitched should be deleted or kept in the server
	KeepStandalone *bool
	// Boolean to indicate if a discovered commodity should be merged when the commodity of the same type already
	// exists in the external entity
	IgnoreCommodityIfPresent *bool
}

func NewDIFProbeArgs() *DIFProbeArgs {
	p := &DIFProbeArgs{}

	p.DiscoveryIntervalSec = flag.Int("discovery-interval-sec", defaultDiscoveryIntervalSec, "The discovery interval in seconds")
	p.KeepStandalone = flag.Bool("keepStandalone", defaultKeepStandalone, "Do we keep non-stitched entities")
	p.IgnoreCommodityIfPresent = flag.Bool("ignoreCommodityIfPresent", defaultIgnoreCommodityIfPresent,
		"Do we ignore a commodity that already exists")

	return p
}
