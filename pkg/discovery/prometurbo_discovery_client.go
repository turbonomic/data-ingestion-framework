package discovery

import (
	"fmt"

	"github.com/golang/glog"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.ibm.com/turbonomic/data-ingestion-framework/version"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type PrometurboDiscoveryClient struct {
	*DIFDiscoveryClient
}

// GetAccountValues gets the Account Values to create VMTTarget in the turbo server corresponding to this client
func (pdc *PrometurboDiscoveryClient) GetAccountValues() *probe.TurboTargetInfo {
	targetParams := pdc.discoveryTargetParams

	targetAddr := ""
	if targetParams.OptionalTargetAddress != nil {
		targetAddr = *targetParams.OptionalTargetAddress
	}

	targetName := ""
	if targetParams.OptionalTargetName != nil {
		targetName = *targetParams.OptionalTargetName
	}

	// this field is used to uniquely identify the target
	targetIDField := registration.TargetIdField
	targetIDVal := &proto.AccountValue{
		Key:         &targetIDField,
		StringValue: &targetName,
	}

	// this field is used specify the address of the target
	targetAddressField := registration.TargetAddressField
	targetAddressVal := &proto.AccountValue{
		Key:         &targetAddressField,
		StringValue: &targetAddr,
	}

	//this field is used as probe version of the target for displaying in the UI
	probeVersionField := registration.ProbeVersion
	probeVersionVal := &proto.AccountValue{
		Key:         &probeVersionField,
		StringValue: &version.Version,
	}

	accountValues := []*proto.AccountValue{
		targetIDVal,
		targetAddressVal,
		probeVersionVal,
	}

	targetInfo := probe.NewTurboTargetInfoBuilder(targetParams.ProbeCategory,
		targetParams.TargetType,
		registration.TargetIdField,
		accountValues).
		Create()

	glog.V(2).Infof("Created target info - id field: '%s', address:%s, name:%s",
		targetInfo.TargetIdentifierField(), targetAddr, targetName)

	return targetInfo
}

// Validate the Target
func (pdc *PrometurboDiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	targetAddr, found := matchingAccountValue(accountValues, registration.TargetAddressField)
	if !found {
		description := fmt.Sprintf("No target address (%s) in account values %v",
			registration.TargetAddressField, accountValueKeyNames(accountValues))
		return failValidation(description), nil
	}
	return pdc.validateTarget(targetAddr)
}

// Discover the Target Topology
func (pdc *PrometurboDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	targetAddr, found := matchingAccountValue(accountValues, registration.TargetAddressField)
	if !found {
		description := fmt.Sprintf("No target address (%s) in account values %v",
			registration.TargetAddressField, accountValueKeyNames(accountValues))
		return failDiscovery(description), nil
	}
	return pdc.discoverTarget(targetAddr)
}
