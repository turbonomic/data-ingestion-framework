package registration

import (
	"fmt"
	"github.com/golang/glog"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	TargetIdField   string = "targetIdentifier" // this is used to reach the target
	TargetNameField string = "Name"             // this is used to display the target in the UI

	propertyId string = "id"
)

// Implements the TurboRegistrationClient interface
type DIFRegistrationClient struct {
	TargetTypeSuffix string
	supplyChain      *SupplyChain
}

func NewDIFRegistrationClient(supplyChainConfig *conf.SupplyChainConfig, targetTypeSuffix string) (*DIFRegistrationClient, error) {
	supplyChain, err := NewSupplyChain(supplyChainConfig)
	if err != nil {
		return nil, fmt.Errorf("Error parsing supply chain %v", err)

	}
	return &DIFRegistrationClient{
		TargetTypeSuffix: targetTypeSuffix,
		supplyChain:      supplyChain,
	}, nil

}

func (p *DIFRegistrationClient) GetSupplyChainDefinition() []*proto.TemplateDTO {
	glog.Infoln("Building supply chain for Data Injection Framework Probe ..........")

	var templateDtos []*proto.TemplateDTO

	templateDtoMap := p.supplyChain.CreateSupplyChainNodeTemplates()

	for _, templateDto := range templateDtoMap {
		glog.Infof("Template DTO for %s : \n		%++v\n", templateDto.TemplateClass, protobuf.MarshalTextString(templateDto))
		templateDtos = append(templateDtos, templateDto)
	}

	glog.Infoln("Supply chain for DIFTurbo is created.")
	return templateDtos
}

func (p *DIFRegistrationClient) GetIdentifyingFields() string {
	return TargetIdField
}

func (p *DIFRegistrationClient) GetAccountDefinition() []*proto.AccountDefEntry {
	//this field is used as name of the target for displaying in the UI
	targetDisplayField := true
	nameIDAcctDefEntry := builder.NewAccountDefEntryBuilder(TargetNameField,
		"Name",
		"Name for the metric server endpoint",
		".*",
		true,
		false).
		Create()
	nameIDAcctDefEntry.IsTargetDisplayName = &targetDisplayField

	// this field is used to reach the target
	targetIDAcctDefEntry := builder.NewAccountDefEntryBuilder(TargetIdField,
		"URL",
		"HTTP URL for the JSON DIF data",
		".*",
		true,
		false).
		Create()

	return []*proto.AccountDefEntry{
		nameIDAcctDefEntry,
		targetIDAcctDefEntry,
	}
}

func (p *DIFRegistrationClient) ProbeCategory() string {
	probeCategory := p.supplyChain.GetProbeCategory()
	if len(probeCategory) == 0 {
		probeCategory = conf.DefaultProbeCategory
	}
	return probeCategory
}

func (p *DIFRegistrationClient) ProbeUICategory() string {
	probeUICategory := p.supplyChain.GetProbeUICategory()
	if len(probeUICategory) == 0 {
		probeUICategory = conf.DefaultProbeUICategory
	}
	return probeUICategory
}

// TargetType returns the target type as the default target type appended
// an optional (from configuration) suffix
func (p *DIFRegistrationClient) TargetType() string {
	targetType := p.supplyChain.GetTargetType()
	if len(targetType) == 0 {
		targetType = conf.DefaultTargetType
	}
	if len(p.TargetTypeSuffix) == 0 {
		return targetType
	}
	return targetType + "-" + p.TargetTypeSuffix
}

// Identity metadata for the probe
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
