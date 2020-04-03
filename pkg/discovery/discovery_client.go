package discovery

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/data-ingestion-framework/pkg/dtofactory"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"strings"
)

// Implements the TurboDiscoveryClient interface
type DIFDiscoveryClient struct {
	discoveryTargetParams *DiscoveryTargetParams
	keepStandalone        bool
	metricEndpoint        []string
	supplyChainConfig     *conf.SupplyChainConfig
}

type EntityBuilderParams struct {
	keepStandalone    bool
	supplyChainConfig *conf.SupplyChainConfig
}

type DiscoveryTargetParams struct {
	OptionalTargetAddress *string
	TargetType            string
	TargetName            string
	ProbeCategory         string
}

func NewDiscoveryClient(discoveryTargetParams *DiscoveryTargetParams, keepStandalone bool,
	supplyChainConfig *conf.SupplyChainConfig) *DIFDiscoveryClient {
	return &DIFDiscoveryClient{
		discoveryTargetParams: discoveryTargetParams,
		keepStandalone:        keepStandalone,
		supplyChainConfig:     supplyChainConfig,
	}
}

// Get the Account Values to create VMTTarget in the turbo server corresponding to this client
func (d *DIFDiscoveryClient) GetAccountValues() *probe.TurboTargetInfo {
	targetParams := d.discoveryTargetParams

	targetAddr := ""
	if targetParams.OptionalTargetAddress != nil {
		targetAddr = *targetParams.OptionalTargetAddress
	}

	targetId := registration.TargetIdField
	targetIdVal := &proto.AccountValue{
		Key:         &targetId,
		StringValue: &targetAddr,
	}

	targetName := registration.TargetNameField
	targetNameVal := &proto.AccountValue{
		Key:         &targetName,
		StringValue: &targetParams.TargetName,
	}

	accountValues := []*proto.AccountValue{
		targetIdVal,
		targetNameVal,
	}

	targetInfo := probe.NewTurboTargetInfoBuilder(targetParams.ProbeCategory, targetParams.TargetType,
		registration.TargetIdField, accountValues).Create()

	return targetInfo
}

// Validate the Target
func (d *DIFDiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	targetAddr, found := targetAddress(accountValues)
	if !found {
		description := fmt.Sprintf("No target address (%s) in account values %v",
			registration.TargetIdField, accountValueKeyNames(accountValues))
		return d.failValidation(description), nil
	}
	validationResponse := &proto.ValidationResponse{}

	// validate metric endpoint address
	var errors []error
	endpoints := strings.Split(targetAddr, ",")
	for _, ep := range endpoints {
		err := ValidateMetricDataSource(ep)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		err := fmt.Errorf("%v", errors)
		validationResponse = d.failValidationWithError(targetAddr, err)
	}

	glog.V(2).Infof("Validating to validate target %s", targetAddr)

	return validationResponse, nil
}

// Discover the Target Topology
func (d *DIFDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {

	glog.V(2).Infof("Discovering the target %s", accountValues)
	targetAddr, found := targetAddress(accountValues)
	if !found {
		description := fmt.Sprintf("No target address (%s) in account values %v",
			registration.TargetIdField, accountValueKeyNames(accountValues))
		return d.failDiscovery(description), nil
	}

	// Create datasources to get the metrics or topology
	endpoints := strings.Split(targetAddr, ",")
	glog.Infof("Metric endpoints %++v", endpoints)
	var datasources []MetricDataSource

	for _, ep := range endpoints {
		datasource := CreateMetricDataSource(ep)
		datasources = append(datasources, datasource)
	}

	// Get the data from all the sources
	difResult, err := GetDIFData(datasources)
	if err != nil {
		glog.Errorf("err : %++v", err)
		return d.failDiscoveryWithError(targetAddr, err), nil
	}

	glog.V(2).Infof("Number of parsed entities %d", len(difResult.ParsedEntities))

	// Create the repository and entities
	repository := data.NewDIFRepository()
	repository.InitRepository(difResult.ParsedEntities)

	for entityType, eMap := range repository.EntityMap {
		for entityId, cdpEntity := range eMap {
			glog.V(4).Infof("Entity %s::%s	-----> ", entityType, entityId)
			for pType, pIds := range cdpEntity.GetProviders() {
				glog.V(4).Infof("		provider %s [%v] ---> ", pType, pIds)
			}
			for pType, pMap := range cdpEntity.HostsByIP {
				for pId, _ := range pMap {
					glog.V(4).Infof("		hostedBy %s [%v] ---> ", pType, pId)
				}
			}
		}
	}

	// Build Entity DTOs
	dtos, err := d.buildEntities(repository, difResult.Scope)
	if err != nil {
		return d.failDiscoveryWithError(targetAddr, err), nil
	}

	return &proto.DiscoveryResponse{EntityDTO: dtos}, nil
}

func accountValueKeyNames(accountValues []*proto.AccountValue) []*string {
	names := make([]*string, len(accountValues))
	for i := range accountValues {
		names[i] = accountValues[i].Key
	}
	return names
}

// targetAddress reads the target address from the array of account values.
// The first value returned is the address, if found.
// The second value returned is a bool indicating whether or not the address was successfully found.
func targetAddress(accountValues []*proto.AccountValue) (string, bool) {
	return matchingAccountValue(accountValues, registration.TargetIdField)
}

func matchingAccountValue(accountValues []*proto.AccountValue, matchKey string) (string, bool) {
	for _, value := range accountValues {
		if *value.Key == matchKey {
			return *value.StringValue, true
		}
	}

	return "", false
}

func (d *DIFDiscoveryClient) buildEntities(repository *data.DIFRepository, scope string) ([]*proto.EntityDTO, error) {

	var entities []*proto.EntityDTO

	// Create the supply chain nodes using the config
	supplyChain, err := registration.NewSupplyChain(d.supplyChainConfig)
	if err != nil {
		return entities, err
	}

	supplyChainNodeMap := supplyChain.GetSupplyChainNodes()
	// Build entity DTOs using the corresponding supply chain template
	for cdpEntityType, eMap := range repository.EntityMap {
		for cdpEntityId, cdpEntity := range eMap {
			// entity type
			eType := dtofactory.EntityType(cdpEntityType)
			if eType == nil {
				glog.Errorf("Invalid entity type %v", cdpEntity)
				continue
			}
			entityType := *eType

			supplyChainNode, validType := supplyChainNodeMap[entityType]
			if !validType {
				glog.Errorf("Supply chain does not support entity type %v", cdpEntity)
				continue
			}

			ab := dtofactory.NewGenericEntityBuilder(entityType, cdpEntity, scope,
				d.keepStandalone, supplyChainNode)
			dto, err := ab.BuildEntity()
			if err != nil {
				glog.Errorf("Error building entity %s::%s %++v", cdpEntityType, cdpEntityId, err)
				continue
			}
			entities = append(entities, dto)
		}
	}

	// create proxy providers
	for providerType, pIds := range repository.ExternalProxyProvidersByIP {
		eType := dtofactory.EntityType(providerType)
		if eType == nil {
			glog.Errorf("Invalid entity type %v", providerType)
			continue
		}
		entityType := *eType

		supplyChainNode, validType := supplyChainNodeMap[entityType]
		if !validType {
			glog.Errorf("Supply chain does not support external provider entity type %v", entityType)
			continue
		}

		for _, pId := range pIds {
			pp := dtofactory.NewProxyProviderEntityBuilder(entityType, pId, scope, d.keepStandalone, supplyChainNode)
			dto, err := pp.BuildEntity()
			if err != nil {
				glog.Errorf("Error building entity %s::%s %++v", providerType, pId, err)
				continue
			}
			entities = append(entities, dto)
		}
	}

	return entities, nil
}

func (d *DIFDiscoveryClient) failDiscoveryWithError(targetAddr string, err error) *proto.DiscoveryResponse {
	return d.failDiscovery(fmt.Sprintf("Discovery of %s failed due to error: %v", targetAddr, err))
}

func (d *DIFDiscoveryClient) failDiscovery(description string) *proto.DiscoveryResponse {
	glog.Errorf(description)
	severity := proto.ErrorDTO_CRITICAL
	errorDTO := &proto.ErrorDTO{
		Severity:    &severity,
		Description: &description,
	}
	discoveryResponse := &proto.DiscoveryResponse{
		ErrorDTO: []*proto.ErrorDTO{errorDTO},
	}
	return discoveryResponse
}

func (d *DIFDiscoveryClient) failValidationWithError(targetAddr string, err error) *proto.ValidationResponse {
	return d.failValidation(fmt.Sprintf("Validation of %s failed due to error: %v", targetAddr, err))
}

func (d *DIFDiscoveryClient) failValidation(description string) *proto.ValidationResponse {
	glog.Errorf(description)
	severity := proto.ErrorDTO_CRITICAL
	errorDto := &proto.ErrorDTO{
		Severity:    &severity,
		Description: &description,
	}

	validationResponse := &proto.ValidationResponse{
		ErrorDTO: []*proto.ErrorDTO{errorDto},
	}
	return validationResponse
}
