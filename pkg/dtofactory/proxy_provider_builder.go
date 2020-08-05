package dtofactory

import (
	"github.com/golang/glog"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type ProxyProviderEntityBuilder struct {
	entityType      proto.EntityDTO_EntityType
	scope           string
	keepStandalone  bool
	entityId        string
	supplyChainNode *registration.SupplyChainNode
}

func NewProxyProviderEntityBuilder(entityType proto.EntityDTO_EntityType, entityId string,
	scope string, keepStandalone bool,
	supplyChainNode *registration.SupplyChainNode) *ProxyProviderEntityBuilder {
	return &ProxyProviderEntityBuilder{
		entityType:      entityType,
		entityId:        entityId,
		keepStandalone:  keepStandalone,
		scope:           scope,
		supplyChainNode: supplyChainNode,
	}
}

func (eb *ProxyProviderEntityBuilder) BuildEntity() (*proto.EntityDTO, error) {
	var dto *proto.EntityDTO

	glog.Infof("Building proxy provider %s", eb.entityId)
	entityBuilder := builder.NewEntityDTOBuilder(eb.entityType, eb.entityId).
		DisplayName(eb.entityId)

	// no matching id
	// set properties
	propName := "IP"
	propVal := eb.entityId
	entityBuilder.WithProperty(getEntityPropertyNameValue(propName, propVal))

	// set sold commodities
	// Setting commodities and properties in the entity builder using the supply chain template
	supplyChainNode := eb.supplyChainNode

	supportedComms := supplyChainNode.SupportedComms
	supportedAccessComms := supplyChainNode.SupportedAccessComms

	key := getKey(eb.entityType, eb.entityId, eb.scope)

	var soldCommodities []*proto.CommodityDTO
	for commType, commVal := range supportedComms {
		commBuilder := builder.NewCommodityDTOBuilder(commType)
		if commVal.Key != nil {
			commBuilder.Key(key)
		}
		soldCommodity, err := commBuilder.Create()
		if err != nil {
			glog.Errorf("Failed to create sold commodity %v for %v: %v", commType, eb.entityId, err)
			continue
		}
		soldCommodities = append(soldCommodities, soldCommodity)
	}

	for commType := range supportedAccessComms {
		//using the provider id as the key
		soldCommodity, err := builder.NewCommodityDTOBuilder(commType).Key(key).Create()
		if err != nil {
			glog.Errorf("Failed to create sold commodity %v for %v: %v", commType, eb.entityId, err)
			continue
		}
		soldCommodities = append(soldCommodities, soldCommodity)
	}
	entityBuilder.SellsCommodities(soldCommodities)

	// mark the  provider as proxy
	replacementEntityMetaDataBuilder := builder.NewReplacementEntityMetaDataBuilder()
	metaData := replacementEntityMetaDataBuilder.Matching(propName).Build()
	entityBuilder.ReplacedBy(metaData)

	dto, err := entityBuilder.Create()
	//keepStandalone := false
	//dto.KeepStandalone = &keepStandalone

	if err != nil {
		return nil, err
	}
	logDebug(protobuf.MarshalTextString(dto))

	glog.Infof("Proxy provider DTO: %+v", dto)

	return dto, nil
}
