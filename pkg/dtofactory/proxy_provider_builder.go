package dtofactory

import (
	"fmt"
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

	id := getEntityId(eb.entityType, eb.entityId, eb.scope)
	glog.Infof("****** building proxy provider ... %s", id)
	entityBuilder := builder.NewEntityDTOBuilder(eb.entityType, id).
		DisplayName(getDisplayEntityName(eb.entityType, eb.entityId, eb.scope))

	// no matching id
	// set properties
	propName := "HOST_IP"
	propVal := eb.entityId
	entityBuilder.WithProperty(getEntityPropertyNameValue(propName, propVal))

	// set sold commodities
	// Setting commodities and properties in the entity builder using the supply chain template
	supplyChainNode := eb.supplyChainNode

	supportedComms := supplyChainNode.SupportedComms
	supportedAccessComms := supplyChainNode.SupportedAccessComms

	var soldCommodities []*proto.CommodityDTO
	for commType, commVal := range supportedComms {
		if *commVal.Key != "" {
			key := id //using the provider id as the key
			soldCommodities = append(soldCommodities, createCommodityWithKey(commType, key))
		} else {
			soldCommodities = append(soldCommodities, createCommodity(commType))
		}
	}
	for commType, _ := range supportedAccessComms {
		key := id //using the provider id as the key
		soldCommodities = append(soldCommodities, createCommodityWithKey(commType, key))
	}
	entityBuilder.SellsCommodities(soldCommodities)

	dto, err := entityBuilder.Create()
	if err != nil {
		return nil, err
	}
	logDebug(fmt.Printf, protobuf.MarshalTextString(dto))

	return dto, nil
}
