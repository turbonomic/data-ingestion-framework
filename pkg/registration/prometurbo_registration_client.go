package registration

import (
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type PrometurboRegistrationClient struct {
	*DIFRegistrationClient
}

func (p *PrometurboRegistrationClient) GetAccountDefinition() []*proto.AccountDefEntry {
	// This field is used to uniquely identify the target
	// Note that this is different from the default DIF probe target where the 'URL' is the target identifier
	targetID := builder.NewAccountDefEntryBuilder(TargetIdField,
		"Name",
		"Name for the prometurbo metric server endpoint",
		".*",
		true,
		false).
		Create()

	// This field is used specify the address of the target
	targetAddress := builder.NewAccountDefEntryBuilder(TargetAddressField,
		"URL",
		"HTTP URL for the prometurbo metric server",
		".*",
		true,
		false).
		Create()

	// This field is used as probe version of the target for displaying in UI
	probeVersion := builder.NewAccountDefEntryBuilder(ProbeVersion,
		"Prometurbo Version",
		"Release Version of Prometurbo Probe",
		".*",
		false,
		false).
		Create()

	return []*proto.AccountDefEntry{
		targetID,
		targetAddress,
		probeVersion,
	}
}

func (p *PrometurboRegistrationClient) ProbeCategory() string {
	return PrometurboProbeCategory
}

func (p *PrometurboRegistrationClient) ProbeUICategory() string {
	return PrometurboProbeUICategory
}

func (p *PrometurboRegistrationClient) TargetType() string {
	return PrometurboTargetType
}
