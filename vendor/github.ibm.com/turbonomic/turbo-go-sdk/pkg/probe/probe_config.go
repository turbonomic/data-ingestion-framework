package probe

import (
	"fmt"

	"github.com/golang/glog"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg"
)

// ProbeConfig hosts the configuration for Probe
// probeCategory - information about probe category, used for categorizing the probe in UI
// probeType - information about probe type
// discoveryMetadata - information about the discovery intervals for the probe
type ProbeConfig struct {
	ProbeType         string
	ProbeSubType      string
	ProbeCategory     string
	ProbeUICategory   string
	discoveryMetadata *DiscoveryMetadata
	Version           string
	DisplayName       string
}

// NewProbeConfig creates a new instance of ProbeConfig.
// Sets default discovery intervals for the full, incremental and performance discoveries.
// Returns error if the probe type and category fields cannot be validated.
func NewProbeConfig(probeType, probeSubType, probeCategory, probeUICategory string) (*ProbeConfig, error) {
	if probeType == "" {
		return nil, ErrorInvalidProbeType()
	}

	if probeCategory == "" {
		return nil, ErrorInvalidProbeCategory()
	}

	if probeUICategory == "" {
		return nil, ErrorInvalidProbeUICategory()
	}

	probeConf := &ProbeConfig{
		ProbeCategory:     probeCategory,
		ProbeType:         probeType,
		ProbeSubType:      probeSubType,
		ProbeUICategory:   probeUICategory,
		discoveryMetadata: NewDiscoveryMetadata(),
	}

	return probeConf, nil
}

// WithVersion sets the given probe version in the probe config.
func (probeConfig *ProbeConfig) WithVersion(version string) *ProbeConfig {
	probeConfig.Version = version
	return probeConfig
}

// WithDisplayName sets a display name for the probe in the probe config.
func (probeConfig *ProbeConfig) WithDisplayName(displayName string) *ProbeConfig {
	probeConfig.DisplayName = displayName
	return probeConfig
}

// Validate the probe config instance
// Returns error if the probe type and category fields cannot be validated.
// Sets default discovery intervals for the full, incremental and performance discoveries.
func (probeConfig *ProbeConfig) Validate() error {
	if probeConfig.ProbeType == "" {
		return ErrorInvalidProbeType()
	}

	if probeConfig.ProbeCategory == "" {
		return ErrorInvalidProbeCategory()
	}

	if probeConfig.ProbeUICategory == "" {
		return ErrorInvalidProbeUICategory()
	}

	if probeConfig.discoveryMetadata == nil {
		probeConfig.discoveryMetadata = NewDiscoveryMetadata()
	}

	return nil
}

// Sets the discovery metadata with intervals for the full, incremental and performance discoveries.
func (probeConfig *ProbeConfig) SetDiscoveryMetadata(discoveryMetadata *DiscoveryMetadata) {
	// validate the discovery intervals
	checkRediscoveryIntervalValidity(discoveryMetadata.fullDiscovery,
		discoveryMetadata.incrementalDiscovery,
		discoveryMetadata.performanceDiscovery)
	probeConfig.discoveryMetadata = discoveryMetadata
}

func checkRediscoveryIntervalValidity(rediscoveryIntervalSec,
	incrementalDiscoverySec,
	performanceDiscoverySec int32) {

	if performanceDiscoverySec >= rediscoveryIntervalSec {
		glog.Warning(discoveryConfigError("performance", "full"))
	}

	if incrementalDiscoverySec >= rediscoveryIntervalSec {
		glog.Warning(discoveryConfigError("incremental", "full"))
	}

	if incrementalDiscoverySec >= performanceDiscoverySec &&
		performanceDiscoverySec != pkg.DISCOVERY_NOT_SUPPORTED {
		glog.Warning(discoveryConfigError("incremental", "performance"))
	}
}

func discoveryConfigError(discoveryType1, discoveryType2 string) string {
	return fmt.Sprintf("%s rediscovery interval is greater than %s rediscovery interval, "+
		"will be skipped!", discoveryType1, discoveryType2)
}
