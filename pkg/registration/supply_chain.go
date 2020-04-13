package registration

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/supplychain"
)

// SupplyChain for the TurboDIF probe created from the supply chain configuration
type SupplyChain struct {
	config  *conf.SupplyChainConfig
	nodeMap map[proto.EntityDTO_EntityType]*SupplyChainNode
}

type SupplyChainNode struct {
	nodeConfig           *conf.NodeConfig
	NodeType             proto.EntityDTO_EntityType
	SupportedComms       map[proto.CommodityDTO_CommodityType]DefaultValue
	SupportedAccessComms map[proto.CommodityDTO_CommodityType]DefaultValue

	// provider type to allowed commodity map
	SupportedBoughtComms       map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]DefaultValue
	SupportedBoughtAccessComms map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]DefaultValue

	// EXTERNAL HOSTED_BY CONFIG
	HostedByProviderProps map[proto.EntityDTO_EntityType][]string
	HostedByProviderType  map[proto.EntityDTO_EntityType]string
	HostedByBoughtComms   map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]DefaultValue
}

// Create new supply chain consisting of supply chain nodes using the supply chain configuration
func NewSupplyChain(config *conf.SupplyChainConfig) (*SupplyChain, error) {
	// parse the node config
	nodeMap := make(map[proto.EntityDTO_EntityType]*SupplyChainNode)
	for _, nodeConfig := range config.Nodes {

		node, err := parseNodeConfig(nodeConfig)
		if err != nil {
			return nil, err
		}
		nodeMap[node.NodeType] = node
	}
	supplyChain := &SupplyChain{
		config:  config,
		nodeMap: nodeMap,
	}

	return supplyChain, nil
}

func (s *SupplyChain) GetProbeCategory() string {
	if s.config.ProbeCategory == nil {
		return conf.DefaultProbeCategory
	}
	return *s.config.ProbeCategory
}

func (s *SupplyChain) GetTargetType() string {
	if s.config.TargetType == nil {
		return conf.DefaultTargetType
	}
	return *s.config.TargetType
}

// Get the supply chain nodes configured in the supply chain
func (s *SupplyChain) GetSupplyChainNodes() map[proto.EntityDTO_EntityType]*SupplyChainNode {
	return s.nodeMap
}

// Create TemplateDTOs for the configured supply chain nodes
func (s *SupplyChain) CreateSupplyChainNodeTemplates() map[proto.EntityDTO_EntityType]*proto.TemplateDTO {
	templateDtoMap := make(map[proto.EntityDTO_EntityType]*proto.TemplateDTO)
	for nodeType, sn := range s.nodeMap {
		templateDto, err := sn.CreateTemplateDTO()
		if err != nil {
			glog.Errorf("Error creating template DTO : %++v", err)
			continue
		}
		templateDtoMap[nodeType] = templateDto
	}
	return templateDtoMap
}

// Parse the node configuration to create the SupplyChainNode structure
func parseNodeConfig(nodeConfig *conf.NodeConfig) (*SupplyChainNode, error) {
	nodeType, exists := TemplateEntityTypeMap[nodeConfig.TemplateClass]

	if !exists {
		return nil, fmt.Errorf("Unknown supply chain node %s", nodeConfig.TemplateClass)
	}

	node := &SupplyChainNode{
		nodeConfig: nodeConfig,
		NodeType:   nodeType,
	}

	err := parseSoldComms(nodeConfig, node)
	if err != nil {
		return nil, err
	}
	err = parseBoughtComms(nodeConfig, node)
	if err != nil {
		return nil, err
	}
	err = parseExternalLinks(nodeConfig, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func parseSoldComms(nodeConfig *conf.NodeConfig, node *SupplyChainNode) error {
	if nodeConfig.CommoditySoldList == nil {
		glog.Infof("%s: no sold commodities\n", nodeConfig.TemplateClass)
		return nil
	}
	supportedComms := make(map[proto.CommodityDTO_CommodityType]DefaultValue)
	supportedAccessComms := make(map[proto.CommodityDTO_CommodityType]DefaultValue)
	if len(nodeConfig.CommoditySoldList) == 0 {
		return fmt.Errorf("%s: Missing commodities in sold commodities section", nodeConfig.TemplateClass)
	}
	for _, sold := range nodeConfig.CommoditySoldList {
		if sold.CommodityType == nil {
			glog.Warningf("%s: Null sold commodity type", nodeConfig.TemplateClass)
			continue
		}
		soldComm := *sold.CommodityType
		if _, exists := TemplateCommodityTypeMap[soldComm]; exists {
			commType := TemplateCommodityTypeMap[soldComm]
			supportedComms[commType] = DefaultValue{Key: sold.Key}
			glog.V(3).Infof("%s Sold comm %s::%v\n", nodeConfig.TemplateClass, soldComm, commType)
		} else {
			glog.Warningf("%s: Invalid sold commodity type %s", nodeConfig.TemplateClass, soldComm)
		}

		if _, exists := AccessTemplateCommodityTypeMap[soldComm]; exists {
			commType := AccessTemplateCommodityTypeMap[soldComm]
			supportedAccessComms[commType] = DefaultValue{Key: sold.Key}
		}
	}
	node.SupportedComms = supportedComms
	node.SupportedAccessComms = supportedAccessComms

	if len(supportedComms) == 0 {
		return fmt.Errorf("%s: Missing commodities in sold commodities section", nodeConfig.TemplateClass)
	}

	return nil
}

func parseBoughtComms(nodeConfig *conf.NodeConfig, node *SupplyChainNode) error {
	if nodeConfig.CommodityBoughtList == nil {
		glog.Infof("%s: no bought commodities\n", nodeConfig.TemplateClass)
		return nil
	}
	// PROVIDER AND BOUGHT COMM CONFIG
	// provider type to allowed commodity map
	supportedBoughtComms := make(map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]DefaultValue)
	supportedBoughtAccessComms := make(map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]DefaultValue)
	for _, bought := range nodeConfig.CommodityBoughtList {
		provider := bought.Provider
		if provider == nil || provider.TemplateClass == nil {
			return fmt.Errorf("%s: Nill provider in bought commodities section", nodeConfig.TemplateClass)
		}

		providerClass := *provider.TemplateClass
		glog.V(3).Infof("%s : providerClass %v\n", nodeConfig.TemplateClass, providerClass)
		if _, exists := TemplateEntityTypeMap[providerClass]; !exists {
			return fmt.Errorf("%s: Invalid provider in bought commodities section for provider %s",
				nodeConfig.TemplateClass, providerClass)
		}

		providerType := TemplateEntityTypeMap[providerClass]
		glog.V(3).Infof("%s : provider type %v\n", nodeConfig.TemplateClass, providerType)
		if bought.Comms == nil || len(bought.Comms) == 0 {
			return fmt.Errorf("%s: Missing bought commodities for provider %s",
				nodeConfig.TemplateClass, providerClass)
		}

		accessCommMap := make(map[proto.CommodityDTO_CommodityType]DefaultValue)
		commMap := make(map[proto.CommodityDTO_CommodityType]DefaultValue)
		for _, comm := range bought.Comms {
			if comm.CommodityType == nil {
				glog.Warningf("%s: Null bought commodity type", nodeConfig.TemplateClass)
				continue
			}
			boughtComm := *comm.CommodityType
			if _, exists := TemplateCommodityTypeMap[boughtComm]; exists {
				commType := TemplateCommodityTypeMap[boughtComm]
				glog.V(3).Infof("%s --> %s Bought comm %s::%s\n", nodeConfig.TemplateClass, providerClass,
					boughtComm, commType)
				commMap[commType] = DefaultValue{Key: comm.Key}
			} else {
				glog.Warningf("%s: Invalid bought commodity type %s", nodeConfig.TemplateClass, boughtComm)
			}
			if _, exists := AccessTemplateCommodityTypeMap[boughtComm]; exists {
				commType := AccessTemplateCommodityTypeMap[boughtComm]
				accessCommMap[commType] = DefaultValue{Key: comm.Key}
			}
		}
		if len(commMap) == 0 {
			return fmt.Errorf("%s: Missing bought commodities for provider %s\n",
				nodeConfig.TemplateClass, providerClass)
		}
		supportedBoughtComms[providerType] = commMap
		supportedBoughtAccessComms[providerType] = accessCommMap
	}

	node.SupportedBoughtComms = supportedBoughtComms
	node.SupportedBoughtAccessComms = supportedBoughtAccessComms

	return nil
}

func parseExternalLinks(nodeConfig *conf.NodeConfig, node *SupplyChainNode) error {
	if nodeConfig.ExternalLinkList == nil {
		glog.Infof("%s: no external links\n", nodeConfig.TemplateClass)
		return nil
	}
	// EXTERNAL HOSTED_BY CONFIG - saving links to external providers
	hostedByProviderProps := make(map[proto.EntityDTO_EntityType][]string)
	hostedByProviderType := make(map[proto.EntityDTO_EntityType]string)
	hostedByBoughtComms := make(map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]DefaultValue)
	for _, externalLink := range nodeConfig.ExternalLinkList {
		if externalLink.Value == nil {
			glog.Warningf("%s: Null external link value", nodeConfig.TemplateClass)
			continue
		}
		link := externalLink.Value
		if link.BuyerRef == nil || link.SellerRef == nil {
			return fmt.Errorf("%s: Null buyer or seller in external link", nodeConfig.TemplateClass)
		}
		if _, exists := TemplateEntityTypeMap[*link.BuyerRef]; !exists {
			return fmt.Errorf("%s: Invalid buyer %s in external link for provider",
				nodeConfig.TemplateClass, *link.BuyerRef)
		}
		buyerType := TemplateEntityTypeMap[*link.BuyerRef]

		if _, exists := TemplateEntityTypeMap[*link.SellerRef]; !exists {
			return fmt.Errorf("%s: Invalid seller %s in external link for provider",
				nodeConfig.TemplateClass, *link.SellerRef)
		}
		sellerType := TemplateEntityTypeMap[*link.SellerRef]

		if sellerType == node.NodeType && buyerType == node.NodeType {
			return fmt.Errorf("%s: Same seller %s and buyer %s in external link",
				nodeConfig.TemplateClass, *link.SellerRef, *link.BuyerRef)
		}
		if sellerType != node.NodeType && buyerType != node.NodeType {
			return fmt.Errorf("%s: Either seller %s or buyer %s must be same as the node in external link",
				nodeConfig.TemplateClass, *link.SellerRef, *link.BuyerRef)
		}

		if link.CommodityDefList == nil || len(link.CommodityDefList) == 0 {
			return fmt.Errorf("%s: Missing commodity metadata in external link", nodeConfig.TemplateClass)
		}

		commMap := make(map[proto.CommodityDTO_CommodityType]DefaultValue)
		for _, commDef := range link.CommodityDefList {
			if commDef.CommType == nil {
				continue
			}
			//commType := TemplateCommodityTypeMap[*commDef.CommType]
			commTypeDef := *commDef.CommType
			if commType, exists := TemplateCommodityTypeMap[commTypeDef]; exists {
				commMap[commType] = DefaultValue{HasKey: commDef.HasKey}
				glog.V(3).Infof("%s : %s-->%s hostedBy comm %s::%v\n", nodeConfig.TemplateClass,
					buyerType, sellerType, commTypeDef, commType)
			} else {
				glog.Warningf("%s: Invalid commodity type %s in external link",
					nodeConfig.TemplateClass, commTypeDef)
			}
		}

		if len(commMap) == 0 {
			return fmt.Errorf("%s: Invalid commodity metadata in external link", nodeConfig.TemplateClass)
		}

		if link.ProbeEntityPropertyList == nil || len(link.ProbeEntityPropertyList) == 0 {
			return fmt.Errorf("%s: Missing internal properties metadata for stitching with external entities",
				nodeConfig.TemplateClass)
		}
		if link.ExternalEntityPropertyList == nil || len(link.ExternalEntityPropertyList) == 0 {
			return fmt.Errorf("%s: Missing external properties metadata for stitching with external entities",
				nodeConfig.TemplateClass)
		}
		var propList []string
		for _, propDef := range link.ProbeEntityPropertyList {
			propList = append(propList, propDef.Name)
		}
		hostedByProviderProps[sellerType] = propList
		hostedByBoughtComms[sellerType] = commMap

		if link.Relationship != nil {
			hostedByProviderType[sellerType] = *link.Relationship
		}
	}

	node.HostedByProviderProps = hostedByProviderProps
	node.HostedByProviderType = hostedByProviderType
	node.HostedByBoughtComms = hostedByBoughtComms

	return nil
}

func (sn *SupplyChainNode) CreateTemplateDTO() (*proto.TemplateDTO, error) {
	snBuilder := supplychain.NewSupplyChainNodeBuilder(sn.NodeType)

	var templateType proto.TemplateDTO_TemplateType
	if _, exists := templateTypeMapping[*sn.nodeConfig.TemplateType]; !exists {
		glog.Warningf("missing template type for node %s", sn.nodeConfig.TemplateClass)
		templateType = proto.TemplateDTO_BASE
	}

	snBuilder.SetTemplateType(templateType)
	snBuilder.SetPriority(sn.nodeConfig.TemplatePriority)

	// Commodity Sold
	for _, commSold := range sn.nodeConfig.CommoditySoldList {
		if commSold.CommodityType == nil {
			continue
		}
		if _, exists := TemplateCommodityTypeMap[*commSold.CommodityType]; exists {
			commType := TemplateCommodityTypeMap[*commSold.CommodityType]
			commTemplate := &proto.TemplateCommodity{
				CommodityType: &commType,
				Key:           commSold.Key,
			}
			glog.V(3).Infof("%s : adding sold comm %++v\n", sn.NodeType, commTemplate)
			snBuilder.Sells(commTemplate)
		} else {
			glog.Errorf("Unsupported sold commodity type %s", *commSold.CommodityType)
		}
	}

	// Commodity Bought
	for _, bought := range sn.nodeConfig.CommodityBoughtList {
		if bought.Provider == nil {
			glog.Warningf("%s: null provider", sn.nodeConfig.TemplateClass)
			continue
		}
		provider := bought.Provider
		if bought.Provider.TemplateClass == nil {
			glog.Warningf("%s: null provider class", sn.nodeConfig.TemplateClass)
			continue
		}
		if _, exists := TemplateEntityTypeMap[*provider.TemplateClass]; !exists {
			continue
		}
		providerType := TemplateEntityTypeMap[*provider.TemplateClass]
		relationship := proto.Provider_LAYERED_OVER
		if provider.ProviderType != nil {
			if _, exists := relationshipMapping[*provider.ProviderType]; exists {
				relationship = relationshipMapping[*provider.ProviderType]
			}
		}

		var commTemplateList []*proto.TemplateCommodity
		for _, comm := range bought.Comms {
			if comm.CommodityType == nil {
				glog.Warningf("%s: null bought comm", sn.nodeConfig.TemplateClass)
			}
			if _, exists := TemplateCommodityTypeMap[*comm.CommodityType]; exists {
				commType := TemplateCommodityTypeMap[*comm.CommodityType]
				glog.V(3).Infof("%s --> %s Bought comm %s::%s\n", sn.nodeConfig.TemplateClass, *provider.TemplateClass,
					*comm.CommodityType, commType)
				commTemplate := &proto.TemplateCommodity{
					CommodityType: &commType,
					Key:           comm.Key,
				}
				commTemplateList = append(commTemplateList, commTemplate)
			}
		}
		snBuilder.Provider(providerType, relationship)
		for _, commTemplate := range commTemplateList {
			glog.V(3).Infof("%s --> %s adding bought comm %++v\n", sn.NodeType, providerType, commTemplate)
			snBuilder.Buys(commTemplate)
		}
	}

	// External Links
	for _, extLink := range sn.nodeConfig.ExternalLinkList {
		if extLink.Value == nil {
			glog.Warningf("%s: null external link", sn.nodeConfig.TemplateClass)
			continue
		}
		link := extLink.Value
		buyerType := TemplateEntityTypeMap[*link.BuyerRef]
		sellerType := TemplateEntityTypeMap[*link.SellerRef]
		relationship := relationshipMapping[*link.Relationship]

		externalLinkBuilder := supplychain.NewExternalEntityLinkBuilder().
			Link(buyerType, sellerType, relationship)
		commDefs := link.CommodityDefList
		for _, commDef := range commDefs {
			commType := TemplateCommodityTypeMap[*commDef.CommType]
			externalLinkBuilder.Commodity(commType, *commDef.HasKey)
		}

		for _, propDef := range link.ProbeEntityPropertyList {
			externalLinkBuilder.ProbeEntityPropertyDef(propDef.Name, propDef.Description)
		}
		for _, propDef := range link.ExternalEntityPropertyList {
			eType := TemplateEntityTypeMap[propDef.Entity]
			var propertyHandler *proto.PropertyHandler
			if propDef.PropHandler != nil {
				propHandlerEntity := TemplateEntityTypeMap[propDef.PropHandler.EntityType]
				propertyHandler = &proto.PropertyHandler{
					MethodName:    &propDef.PropHandler.MethodName,
					EntityType:    &propHandlerEntity,
					DirectlyApply: &propDef.PropHandler.DirectlyApply,
				}
			}

			serverPropDef := &proto.ServerEntityPropDef{
				Entity:          &eType,
				Attribute:       &propDef.Attribute,
				UseTopoExt:      nil,
				PropertyHandler: propertyHandler,
			}
			externalLinkBuilder.ExternalEntityPropertyDef(serverPropDef)
		}
		external, err := externalLinkBuilder.Build()
		if err != nil {
			glog.Errorf("%++v", err)
		}
		snBuilder.ConnectsTo(external)
	}

	snNode, err := snBuilder.Create()
	if err != nil {
		return nil, err
	}

	// Stitching Metadata
	metadata := sn.nodeConfig.MergedEntityMetaData
	if metadata != nil {
		glog.V(3).Infof("metadata %++v\n", metadata)
		var metadataBuilder *builder.MergedEntityMetadataBuilder
		metadataBuilder = builder.NewMergedEntityMetadataBuilder().
			KeepInTopology(metadata.KeepInTopology)
		for _, comm := range metadata.CommSold {
			commType := TemplateCommodityTypeMap[comm]
			metadataBuilder.PatchSoldMetadata(commType, make(map[string][]string))
		}
		for _, bought := range metadata.CommoditiesBought {
			pType := TemplateEntityTypeMap[bought.Provider]
			var commList []proto.CommodityDTO_CommodityType
			for _, comm := range bought.Comm {
				commType := TemplateCommodityTypeMap[comm]
				commList = append(commList, commType)
			}
			metadataBuilder.PatchBoughtList(pType, commList)
		}
		matchingData := metadata.MatchingMetadata
		if matchingData != nil {
			glog.V(3).Infof("matchingData %++v\n", matchingData)
			returnType := returnTypeMapping[matchingData.ReturnType]
			extReturnType := returnTypeMapping[matchingData.ExternalEntityReturnType]
			metadataBuilder.InternalMatchingType(returnType).
				ExternalMatchingType(extReturnType)

			for _, md := range matchingData.MatchingDataList {
				glog.V(3).Infof("internal md %++v\n", md)
				if md.Delimiter == "" {
					metadataBuilder.InternalMatchingProperty(md.MatchingProperty.PropertyName)
				} else {
					metadataBuilder.InternalMatchingPropertyWithDelimiter(md.MatchingProperty.PropertyName, md.Delimiter)
				}
			}

			for _, md := range matchingData.ExternalEntityMatchingPropertyList {
				glog.V(3).Infof("external md %++v\n", md)
				if md.MatchingProperty != nil {
					if md.Delimiter == "" {
						metadataBuilder.ExternalMatchingProperty(md.MatchingProperty.PropertyName)
					} else {
						metadataBuilder.ExternalMatchingPropertyWithDelimiter(md.MatchingProperty.PropertyName, md.Delimiter)
					}
				} else if md.MatchingField != nil {
					if md.Delimiter == "" {
						metadataBuilder.ExternalMatchingField(md.MatchingField.FieldName, md.MatchingField.MessagePath)
					} else {
						metadataBuilder.ExternalMatchingFieldWithDelimiter(md.MatchingField.FieldName,
							md.MatchingField.MessagePath,
							md.Delimiter)
					}
				}
			}
		}

		mergingMetadata, err := metadataBuilder.Build()
		if err != nil {
			return nil, err
		}
		snNode.MergedEntityMetaData = mergingMetadata
	}
	return snNode, nil
}
