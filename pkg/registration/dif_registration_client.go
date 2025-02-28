package registration

import (
	"github.com/golang/glog"
	protobuf "github.com/golang/protobuf/proto"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// DIFRegistrationClient implements the TurboRegistrationClient interface
type DIFRegistrationClient struct {
	supplyChain *SupplyChain
}

func (p *DIFRegistrationClient) GetSupplyChainDefinition() []*proto.TemplateDTO {
	glog.Infoln("Building supply chain for Data Injection Framework Probe ..........")

	var templateDTOs []*proto.TemplateDTO

	templateDtoMap := p.supplyChain.CreateSupplyChainNodeTemplates()

	for _, templateDto := range templateDtoMap {
		glog.V(4).Infof("Template DTO for %s:", templateDto.TemplateClass)
		glog.V(4).Infof("%s", protobuf.MarshalTextString(templateDto))
		templateDTOs = append(templateDTOs, templateDto)
	}

	glog.Infoln("Supply chain for DIFTurbo is created.")
	return templateDTOs
}

func (p *DIFRegistrationClient) GetIdentifyingFields() string {
	return TargetIdField
}

func (p *DIFRegistrationClient) GetAccountDefinition() []*proto.AccountDefEntry {
	// this field is used as name of the target for displaying in the UI
	targetID := builder.NewAccountDefEntryBuilder(TargetIdField,
		"Name",
		"Name for the metric server endpoint",
		".*",
		true,
		false).
		Create()

	// this field is used to reach the target
	targetAddress := builder.NewAccountDefEntryBuilder(TargetAddressField,
		"URL",
		"HTTP URL for the JSON DIF data",
		".*",
		true,
		false).
		Create()

	// this field is used as probe version of the target for displaying in UI
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

func (p *DIFRegistrationClient) ProbeCategory() string {
	return conf.DefaultProbeCategory
}

func (p *DIFRegistrationClient) ProbeUICategory() string {
	return conf.DefaultProbeUICategory
}

// TargetType returns the target type as the default target type appended
// an optional (from configuration) suffix
func (p *DIFRegistrationClient) TargetType() string {
	return conf.DefaultTargetType
}

// GetEntityMetadata identities entity metadata for the probe
func (p *DIFRegistrationClient) GetEntityMetadata() []*proto.EntityIdentityMetadata {
	glog.V(3).Infof("Begin to build EntityIdentityMetadata for DIF Probe ...")

	var result []*proto.EntityIdentityMetadata
	var entities []proto.EntityDTO_EntityType
	nodeMap := p.supplyChain.GetSupplyChainNodes()
	for nodeType := range nodeMap {
		entities = append(entities, nodeType)
	}

	for _, eType := range entities {
		meta := p.newIdMetaData(eType, []string{propertyId})
		result = append(result, meta)
	}

	glog.V(2).Infof("EntityIdentityMetaData: %++v", result)

	return result
}

func (p *DIFRegistrationClient) newIdMetaData(eType proto.EntityDTO_EntityType, names []string) *proto.EntityIdentityMetadata {
	var data []*proto.EntityIdentityMetadata_PropertyMetadata
	for _, name := range names {
		dat := &proto.EntityIdentityMetadata_PropertyMetadata{
			Name: &name,
		}
		data = append(data, dat)
	}

	result := &proto.EntityIdentityMetadata{
		EntityType:            &eType,
		NonVolatileProperties: data,
	}

	return result
}

func (p *DIFRegistrationClient) GetActionPolicy() []*proto.ActionPolicyDTO {
	glog.V(3).Infof("Begin to build Action Policies")
	ab := builder.NewActionPolicyBuilder()
	supported := proto.ActionPolicyDTO_SUPPORTED
	recommend := proto.ActionPolicyDTO_NOT_EXECUTABLE
	notSupported := proto.ActionPolicyDTO_NOT_SUPPORTED

	// 1. containerPod: support move, provision and suspend; not resize;
	pod := proto.EntityDTO_CONTAINER_POD
	podPolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	podPolicy[proto.ActionItemDTO_MOVE] = supported
	podPolicy[proto.ActionItemDTO_PROVISION] = supported
	podPolicy[proto.ActionItemDTO_RIGHT_SIZE] = notSupported
	podPolicy[proto.ActionItemDTO_SUSPEND] = supported

	p.addActionPolicy(ab, pod, podPolicy)

	// 2. container: support resize; recommend provision and suspend; not move;
	container := proto.EntityDTO_CONTAINER
	containerPolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	containerPolicy[proto.ActionItemDTO_RIGHT_SIZE] = supported
	containerPolicy[proto.ActionItemDTO_PROVISION] = recommend
	containerPolicy[proto.ActionItemDTO_MOVE] = notSupported
	containerPolicy[proto.ActionItemDTO_SUSPEND] = recommend

	p.addActionPolicy(ab, container, containerPolicy)

	// 3. application: only recommend provision and suspend; all else are not supported
	app := proto.EntityDTO_APPLICATION_COMPONENT
	appPolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	appPolicy[proto.ActionItemDTO_PROVISION] = recommend
	appPolicy[proto.ActionItemDTO_RIGHT_SIZE] = recommend
	appPolicy[proto.ActionItemDTO_MOVE] = notSupported
	appPolicy[proto.ActionItemDTO_SUSPEND] = recommend

	p.addActionPolicy(ab, app, appPolicy)

	// 4. service: no actions are supported
	service := proto.EntityDTO_SERVICE
	servicePolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	servicePolicy[proto.ActionItemDTO_PROVISION] = notSupported
	servicePolicy[proto.ActionItemDTO_RIGHT_SIZE] = notSupported
	servicePolicy[proto.ActionItemDTO_MOVE] = notSupported
	servicePolicy[proto.ActionItemDTO_SUSPEND] = notSupported

	p.addActionPolicy(ab, service, servicePolicy)

	// 5. node: support provision and suspend; not resize; do not set move
	node := proto.EntityDTO_VIRTUAL_MACHINE
	nodePolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	nodePolicy[proto.ActionItemDTO_RIGHT_SIZE] = notSupported
	nodePolicy[proto.ActionItemDTO_SCALE] = notSupported

	// node provision/suspend default is recommend.
	// During Discovery, if cluster API is enabled (i.e. openshift), this will change to "supported"
	nodePolicy[proto.ActionItemDTO_PROVISION] = recommend
	nodePolicy[proto.ActionItemDTO_SUSPEND] = recommend

	p.addActionPolicy(ab, node, nodePolicy)

	// 6. workload controller: support  resize
	controller := proto.EntityDTO_WORKLOAD_CONTROLLER
	controllerPolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	controllerPolicy[proto.ActionItemDTO_RIGHT_SIZE] = supported
	controllerPolicy[proto.ActionItemDTO_SCALE] = supported

	p.addActionPolicy(ab, controller, controllerPolicy)

	// 7. volumes
	volume := proto.EntityDTO_VIRTUAL_VOLUME
	volumePolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	volumePolicy[proto.ActionItemDTO_PROVISION] = recommend
	volumePolicy[proto.ActionItemDTO_RIGHT_SIZE] = notSupported
	volumePolicy[proto.ActionItemDTO_SCALE] = recommend
	volumePolicy[proto.ActionItemDTO_SUSPEND] = notSupported

	p.addActionPolicy(ab, volume, volumePolicy)

	// 8. business apps
	businessApp := proto.EntityDTO_BUSINESS_APPLICATION
	businessAppPolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	businessAppPolicy[proto.ActionItemDTO_PROVISION] = notSupported
	businessAppPolicy[proto.ActionItemDTO_RIGHT_SIZE] = notSupported
	businessAppPolicy[proto.ActionItemDTO_SCALE] = notSupported
	businessAppPolicy[proto.ActionItemDTO_SUSPEND] = notSupported

	p.addActionPolicy(ab, businessApp, businessAppPolicy)

	return ab.Create()
}

func (p *DIFRegistrationClient) addActionPolicy(ab *builder.ActionPolicyBuilder,
	entity proto.EntityDTO_EntityType,
	policies map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability,
) {
	for action, policy := range policies {
		ab.WithEntityActions(entity, action, policy)
	}
}

func (p *DIFRegistrationClient) GetActionMergePolicy() []*proto.ActionMergePolicyDTO {
	glog.V(2).Infof("Begin to build Action Merge Policies")
	// resize action
	resizeActionMergeTarget := builder.NewActionDeDuplicateAndAggregationTargetBuilder().
		DeDuplicatedBy(builder.NewActionAggregationTargetBuilder(proto.EntityDTO_CONTAINER_SPEC,
			proto.ConnectedEntity_CONTROLLED_BY_CONNECTION)).
		AggregatedBy(builder.NewActionAggregationTargetBuilder(proto.EntityDTO_WORKLOAD_CONTROLLER,
			proto.ConnectedEntity_CONTROLLED_BY_CONNECTION))

	resizeActionMergeTarget2 := builder.NewActionDeDuplicateAndAggregationTargetBuilder().
		DeDuplicatedBy(builder.NewActionAggregationTargetBuilder(proto.EntityDTO_CONTAINER_SPEC,
			proto.ConnectedEntity_CONTROLLED_BY_CONNECTION)).
		AggregatedBy(builder.NewActionAggregationTargetBuilder(proto.EntityDTO_WORKLOAD_CONTROLLER,
			proto.ConnectedEntity_OWNS_CONNECTION))

	containerResizeMerge := builder.NewMergePolicyBuilder().
		ForEntityType(proto.EntityDTO_CONTAINER).
		ForCommodity(proto.CommodityDTO_VCPU).
		ForCommodity(proto.CommodityDTO_VMEM).
		ForCommodity(proto.CommodityDTO_VCPU_REQUEST).
		ForCommodity(proto.CommodityDTO_VMEM_REQUEST).
		DeDuplicateAndAggregateBy(resizeActionMergeTarget).
		DeDuplicateAndAggregateBy(resizeActionMergeTarget2)

	// horizontal scale action
	horizontalScaleActionMergeTarget := builder.NewActionAggregationTargetBuilder(proto.EntityDTO_WORKLOAD_CONTROLLER,
		proto.ConnectedEntity_AGGREGATED_BY_CONNECTION)

	containerPodDataDaemonSet := proto.EntityDTO_ContainerPodData{
		ControllerData: &proto.EntityDTO_WorkloadControllerData{
			ControllerType: &proto.EntityDTO_WorkloadControllerData_DaemonSetData{
				DaemonSetData: &proto.EntityDTO_DaemonSetData{},
			},
		},
	}

	containerPodDataReplicaSet := proto.EntityDTO_ContainerPodData{
		ControllerData: &proto.EntityDTO_WorkloadControllerData{
			ControllerType: &proto.EntityDTO_WorkloadControllerData_ReplicaSetData{
				ReplicaSetData: &proto.EntityDTO_ReplicaSetData{},
			},
		},
	}

	containerHorizontalScaleMerge := builder.NewMergePolicyBuilder().
		ForEntityType(proto.EntityDTO_CONTAINER_POD).
		ForCommodity(proto.CommodityDTO_RESPONSE_TIME).
		ForCommodity(proto.CommodityDTO_SERVICE_TIME).
		ForCommodity(proto.CommodityDTO_QUEUING_TIME).
		ForCommodity(proto.CommodityDTO_TRANSACTION).
		ForCommodity(proto.CommodityDTO_CONCURRENT_QUERIES).
		ForCommodity(proto.CommodityDTO_LLM_CACHE).
		ForContainerPodDataExclusionFilter(&containerPodDataDaemonSet).
		ForContainerPodDataExclusionFilter(&containerPodDataReplicaSet).
		AggregateBy(horizontalScaleActionMergeTarget)

	return builder.NewActionMergePolicyBuilder().
		ForResizeAction(proto.EntityDTO_CONTAINER, containerResizeMerge).
		ForHorizontalScaleAction(proto.EntityDTO_CONTAINER_POD, containerHorizontalScaleMerge).
		Create()
}
