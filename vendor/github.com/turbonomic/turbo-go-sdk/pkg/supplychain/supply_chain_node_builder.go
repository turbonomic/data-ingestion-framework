package supplychain

import (
	"fmt"
	"math"

	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

// provider creates a template entity that sells commodities to a consumer template.
type provider struct {
	provider *proto.Provider
	// Specifies if the provider is optional or not.
	// For example the provider of Pod is WorkloadController if Pod is deployed by K8s controller, otherwise the provider
	// of Pod is Namespace. So WorkloadController and Namespace are optional providers for Pod.
	optional *bool
}
type SupplyChainNodeBuilder struct {
	templateClass              *proto.EntityDTO_EntityType
	templateType               *proto.TemplateDTO_TemplateType
	priority                   *int32
	commoditiesSold            []*proto.TemplateCommodity
	providerCommodityBoughtMap map[*provider][]*proto.TemplateCommodity
	externalLinks              []*proto.TemplateDTO_ExternalEntityLinkProp
	currentProvider            *provider

	err error
}

// Create a new SupplyChainNode Builder.
// All the new supply chain node are default to use the base template type and priority 0.
func NewSupplyChainNodeBuilder(entityType proto.EntityDTO_EntityType) *SupplyChainNodeBuilder {
	templateType := proto.TemplateDTO_BASE
	priority := int32(0)
	return &SupplyChainNodeBuilder{
		templateClass: &entityType,
		templateType:  &templateType,
		priority:      &priority,
	}
}

// Create a SupplyChainNode
func (scnb *SupplyChainNodeBuilder) Create() (*proto.TemplateDTO, error) {
	if scnb.err != nil {
		return nil, fmt.Errorf("Cannot create supply chain node because of error: %v", scnb.err)
	}
	return &proto.TemplateDTO{
		TemplateClass:    scnb.templateClass,
		TemplateType:     scnb.templateType,
		TemplatePriority: scnb.priority,
		CommoditySold:    scnb.commoditiesSold,
		CommodityBought:  buildCommodityBought(scnb.providerCommodityBoughtMap),
		ExternalLink:     scnb.externalLinks,
	}, nil
}

// The very basic selling method. If want others, use other names
func (scnb *SupplyChainNodeBuilder) Sells(templateComm *proto.TemplateCommodity) *SupplyChainNodeBuilder {
	if scnb.err != nil {
		return scnb
	}

	if scnb.commoditiesSold == nil {
		scnb.commoditiesSold = []*proto.TemplateCommodity{}
	}
	scnb.commoditiesSold = append(scnb.commoditiesSold, templateComm)
	return scnb
}

// Provider sets the mandatory provider of the SupplyChainNode
func (scnb *SupplyChainNodeBuilder) Provider(providerType proto.EntityDTO_EntityType,
	pType proto.Provider_ProviderType) *SupplyChainNodeBuilder {
	return scnb.ProviderOpt(providerType, pType, nil)
}

// ProviderOpt sets the optional provider of the SupplyChainNode with min and max cardinality
func (scnb *SupplyChainNodeBuilder) ProviderOpt(providerType proto.EntityDTO_EntityType,
	pType proto.Provider_ProviderType, optional *bool) *SupplyChainNodeBuilder {
	if scnb.err != nil {
		return scnb
	}
	var minCardinality int32
	if optional != nil && *optional {
		// If provider is optional, set minCardinality to 0
		minCardinality = int32(0)
	} else {
		// else minCardinality is 0
		minCardinality = int32(1)
	}
	var maxCardinality int32
	if pType == proto.Provider_LAYERED_OVER {
		maxCardinality = int32(math.MaxInt32)
	} else {
		maxCardinality = int32(1)
	}
	scnb.currentProvider = &provider{
		provider: &proto.Provider{
			TemplateClass:  &providerType,
			ProviderType:   &pType,
			CardinalityMax: &maxCardinality,
			CardinalityMin: &minCardinality,
		},
		optional: optional,
	}
	return scnb
}

// Add a commodity this node buys from the current provider. The provider must already been specified.
// If there is no provider for this node, does not add the commodity.
func (scnb *SupplyChainNodeBuilder) Buys(templateComm *proto.TemplateCommodity) *SupplyChainNodeBuilder {
	if scnb.err != nil {
		return scnb
	}

	if hasProvider := scnb.requireProvider(); !hasProvider {
		scnb.err = fmt.Errorf("Provider must be set before calling Buys().")
		return scnb
	}

	if scnb.providerCommodityBoughtMap == nil {
		scnb.providerCommodityBoughtMap = make(map[*provider][]*proto.TemplateCommodity)
	}

	templateCommoditiesSoldByCurrentProvider, exist := scnb.providerCommodityBoughtMap[scnb.currentProvider]
	if !exist {
		templateCommoditiesSoldByCurrentProvider = []*proto.TemplateCommodity{}
	}
	templateCommoditiesSoldByCurrentProvider = append(templateCommoditiesSoldByCurrentProvider, templateComm)
	scnb.providerCommodityBoughtMap[scnb.currentProvider] = templateCommoditiesSoldByCurrentProvider

	return scnb
}

// Adds an external entity link to the current node.
// This means the current node will connect an entity discovered by other probe in the full supply chain. The external
// entity can be a provider or a consumer.The connection configurations, for example type, id and connection type are
// specified in externalEntityLink.
func (scnb *SupplyChainNodeBuilder) ConnectsTo(extEntityLink *proto.ExternalEntityLink) *SupplyChainNodeBuilder {
	if scnb.err != nil {
		return scnb
	}

	linkProp, err := scnb.buildExternalEntityLinkProperty(extEntityLink)
	if err != nil {
		scnb.err = err
		return scnb
	}
	currentLinks := scnb.externalLinks
	if currentLinks == nil {
		currentLinks = []*proto.TemplateDTO_ExternalEntityLinkProp{}
	}
	scnb.externalLinks = append(currentLinks, linkProp)
	return scnb
}

func (scnb SupplyChainNodeBuilder) buildExternalEntityLinkProperty(
	extEntityLink *proto.ExternalEntityLink) (*proto.TemplateDTO_ExternalEntityLinkProp, error) {
	entityType := scnb.templateClass
	var key *proto.EntityDTO_EntityType
	if extEntityLink.GetBuyerRef() == *entityType {
		sellerType := extEntityLink.GetSellerRef()
		key = &sellerType
	} else if extEntityLink.GetSellerRef() == *entityType {
		buyerType := extEntityLink.GetBuyerRef()
		key = &buyerType
	} else {
		return nil, fmt.Errorf("Template entity type %v does match types in this external link", entityType)
	}
	return &proto.TemplateDTO_ExternalEntityLinkProp{
		Key:   key,
		Value: extEntityLink,
	}, nil
}

// Check if the provider has been set.
func (scnb *SupplyChainNodeBuilder) requireProvider() bool {
	if scnb.currentProvider == nil {
		return false
	}
	return true
}

func (scnb *SupplyChainNodeBuilder) SetPriority(p int32) {
	scnb.priority = &p
}

func (scnb *SupplyChainNodeBuilder) SetTemplateType(t proto.TemplateDTO_TemplateType) {
	scnb.templateType = &t
}

func buildCommodityBought(providerCommodityBoughtMap map[*provider][]*proto.TemplateCommodity) []*proto.TemplateDTO_CommBoughtProviderProp {
	if len(providerCommodityBoughtMap) == 0 {
		return nil
	}
	commBought := []*proto.TemplateDTO_CommBoughtProviderProp{}
	for provider, templateCommodities := range providerCommodityBoughtMap {
		commBought = append(commBought, &proto.TemplateDTO_CommBoughtProviderProp{
			Key:        provider.provider,
			Value:      templateCommodities,
			IsOptional: provider.optional,
		})
	}
	return commBought
}
