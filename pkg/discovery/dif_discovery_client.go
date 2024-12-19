package discovery

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/dtofactory"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.ibm.com/turbonomic/data-ingestion-framework/version"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// DIFDiscoveryClient implements the TurboDiscoveryClient interface
type DIFDiscoveryClient struct {
	discoveryTargetParams *TargetParams
	keepStandalone        bool
	metricEndpoint        []string
	supplyChain           *registration.SupplyChain
	bindingChannel        string
}

type TargetParams struct {
	OptionalTargetAddress *string
	OptionalTargetName    *string
	TargetType            string
	ProbeCategory         string
}

type DifSecureProbeTargetProvider struct {
	//discoveryClient DIFDiscoveryClient
	targetInfo     *probe.TurboTargetInfo
	bindingChannel string
}

func NewDiscoveryClient(targetParams *TargetParams, keepStandalone bool,
	supplyChain *registration.SupplyChain, bindingChannel string) (probe.TurboDiscoveryClient, probe.ISecureProbeTargetProvider) {
	difDiscoveryClient := &DIFDiscoveryClient{
		discoveryTargetParams: targetParams,
		keepStandalone:        keepStandalone,
		supplyChain:           supplyChain,
		bindingChannel:        bindingChannel,
	}
	if registration.IsPrometurboProbe(supplyChain.GetTargetType()) {
		prometurboDiscoveryClient := &PrometurboDiscoveryClient{
			difDiscoveryClient,
		}
		return prometurboDiscoveryClient, DifSecureProbeTargetProvider{
			targetInfo:     prometurboDiscoveryClient.GetAccountValues(),
			bindingChannel: bindingChannel,
		}
	}
	return difDiscoveryClient, DifSecureProbeTargetProvider{
		targetInfo:     difDiscoveryClient.GetAccountValues(),
		bindingChannel: bindingChannel,
	}
}

func (targetProvider DifSecureProbeTargetProvider) GetTargetIdentifier() string {
	if targetProvider.targetInfo.TargetIdentifierField() == "" {
		glog.Warning("Cannot build default secure probe target, target identifier is not provided")
	}
	return targetProvider.targetInfo.TargetIdentifierField()
}

func (targetProvider DifSecureProbeTargetProvider) GetSecureProbeTarget() *proto.ProbeTargetInfo {
	// Do not register the following account definitions if no target has been defined
	// in kubeturbo configuration. The target will be added manually.
	if targetProvider.targetInfo.TargetIdentifierField() == "" {
		return &proto.ProbeTargetInfo{}
	}
	glog.V(2).Infof("Begin to build default secure probe target")
	return &proto.ProbeTargetInfo{
		InputValues:                 targetProvider.targetInfo.AccountValues(),
		CommunicationBindingChannel: &targetProvider.bindingChannel,
	}
}

// GetAccountValues gets the Account Values to create VMTTarget in the turbo server corresponding to this client
func (d *DIFDiscoveryClient) GetAccountValues() *probe.TurboTargetInfo {
	targetParams := d.discoveryTargetParams

	targetAddr := ""
	if targetParams.OptionalTargetAddress != nil {
		targetAddr = *targetParams.OptionalTargetAddress
	}

	targetName := ""
	if targetParams.OptionalTargetName != nil {
		targetName = *targetParams.OptionalTargetName
	}

	// this field is used to reach the target
	targetIdField := registration.TargetIdField
	targetIdVal := &proto.AccountValue{
		Key:         &targetIdField,
		StringValue: &targetName,
	}

	// this field is used as name of the target for displaying in the UI
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
		targetIdVal,
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
func (d *DIFDiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	glog.V(2).Infof("Validating target %s", accountValues)
	targetAddr, found := matchingAccountValue(accountValues, registration.TargetAddressField)
	if !found {
		description := fmt.Sprintf("No target address (%s) in account values %v",
			registration.TargetIdField, accountValueKeyNames(accountValues))
		return failValidation(description), nil
	}
	return d.validateTarget(targetAddr)
}

func (d *DIFDiscoveryClient) validateTarget(targetAddr string) (*proto.ValidationResponse, error) {
	// validate metric endpoint address
	validationResponse := &proto.ValidationResponse{}
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
		validationResponse = failValidationWithError(targetAddr, err)
	}
	return validationResponse, nil
}

// Discover the Target Topology
func (d *DIFDiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	targetAddr, found := matchingAccountValue(accountValues, registration.TargetAddressField)
	if !found {
		description := fmt.Sprintf("No target address (%s) in account values %v",
			registration.TargetIdField, accountValueKeyNames(accountValues))
		return failDiscovery(description), nil
	}
	return d.discoverTarget(targetAddr)
}

func (d *DIFDiscoveryClient) discoverTarget(targetAddr string) (*proto.DiscoveryResponse, error) {
	glog.V(2).Infof("Discovering target %s", targetAddr)
	// Create data sources to get the metrics or topology
	endpoints := strings.Split(targetAddr, ",")
	glog.Infof("Metric endpoints %++v", endpoints)
	var dataSources []MetricDataSource

	for _, ep := range endpoints {
		datasource := CreateMetricDataSource(ep)
		if datasource != nil {
			dataSources = append(dataSources, datasource)
		}
	}

	if len(dataSources) == 0 {
		return nil, fmt.Errorf("no valid data source")
	}

	// Get the data from all the sources
	difResult, err := GetDIFData(dataSources)
	if err != nil {
		glog.Errorf("err : %++v", err)
		return failDiscoveryWithError(targetAddr, err), nil
	}

	glog.V(2).Infof("Number of parsed entities %d", len(difResult.ParsedEntities))

	// Create the repository and entities
	repository := data.NewDIFRepository()
	repository.InitRepository(difResult.ParsedEntities)

	for entityType, eMap := range repository.EntityMap {
		for entityId, difEntity := range eMap {
			glog.V(4).Infof("Entity %s::%s	-----> ", entityType, entityId)
			for pType, pIds := range difEntity.GetProviders() {
				glog.V(4).Infof("		provider %s [%v] ---> ", pType, pIds)
			}
			for pType, pMap := range difEntity.HostsByIP {
				for pId := range pMap {
					glog.V(4).Infof("		hostedBy %s [%v] ---> ", pType, pId)
				}
			}
		}
	}

	// Build Entity DTOs
	DTOs, err := d.buildEntities(repository, difResult.Scope)
	if err != nil {
		return failDiscoveryWithError(targetAddr, err), nil
	}

	return &proto.DiscoveryResponse{EntityDTO: DTOs}, nil
}

func accountValueKeyNames(accountValues []*proto.AccountValue) []*string {
	names := make([]*string, len(accountValues))
	for i := range accountValues {
		names[i] = accountValues[i].Key
	}
	return names
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

	idToEntityMap := make(map[string]*data.BasicDIFEntity)
	supplyChainNodeMap := d.supplyChain.GetSupplyChainNodes()
	// Build entity DTOs using the corresponding supply chain template
	for difEntityType, eMap := range repository.EntityMap {
		for difEntityId, difEntity := range eMap {
			// entity type
			eType := dtofactory.EntityType(difEntityType)
			if eType == nil {
				glog.Errorf("Invalid entity type %v", difEntity)
				continue
			}
			entityType := *eType

			supplyChainNode, validType := supplyChainNodeMap[entityType]
			if !validType {
				glog.Errorf("Supply chain does not support entity type %v", difEntity)
				continue
			}
			if existingEntity, seen := idToEntityMap[difEntityId]; seen {
				// Entities with the same ID and same type will be merged. Entities with the same ID but different
				// types will be rejected.
				glog.Errorf("Duplicated entity ID detected with different entity types. Entity ID: %v,"+
					" Entity types: %v, %v", difEntityId, existingEntity.EntityType, difEntityType)
				continue
			}
			ab := dtofactory.NewGenericEntityBuilder(entityType, difEntity, scope,
				d.keepStandalone, supplyChainNode)
			dto, err := ab.BuildEntity()
			if err != nil {
				glog.Errorf("Error building entity %s::%s %++v", difEntityType, difEntityId, err)
				continue
			}
			idToEntityMap[difEntityId] = difEntity
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

func failDiscoveryWithError(targetAddr string, err error) *proto.DiscoveryResponse {
	return failDiscovery(fmt.Sprintf("Discovery of %s failed due to error: %v", targetAddr, err))
}

func failDiscovery(description string) *proto.DiscoveryResponse {
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

func failValidationWithError(targetAddr string, err error) *proto.ValidationResponse {
	return failValidation(fmt.Sprintf("Validation of %s failed due to error: %v", targetAddr, err))
}

func failValidation(description string) *proto.ValidationResponse {
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
