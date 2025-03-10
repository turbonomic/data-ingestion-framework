package registration

import (
	"os"

	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/probe"
)

type ProbeRegistrationClient interface {
	probe.IEntityMetadataProvider
	probe.IActionMergePolicyProvider
	probe.IActionPolicyProvider
	probe.TurboRegistrationClient
	TargetType() string
	ProbeCategory() string
	ProbeUICategory() string
}

func NewRegistrationClient(supplyChain *SupplyChain) ProbeRegistrationClient {
	difRegistrationClient := &DIFRegistrationClient{
		supplyChain: supplyChain,
	}
	if IsPrometurboProbe(supplyChain.GetTargetType()) {
		return &PrometurboRegistrationClient{
			difRegistrationClient,
		}
	}
	return difRegistrationClient
}

// IsPrometurboProbe determines if the current probe is a Prometurbo probe
// We consider a probe to be a Prometurbo probe if:
//   - TURBODIF_TARGET_TYPE_OVERWRITE environment variable is set to Prometheus, or
//   - In the absence of the TURBODIF_TARGET_TYPE_OVERWRITE environment variable, we have DataIngestionFramework
//     target type. This is in case t8c-operator is not upgraded to the latest version.
func IsPrometurboProbe(targetType string) bool {
	value := os.Getenv(TargetTypeOverwriteEnv)
	return value == PrometurboTargetType
}
