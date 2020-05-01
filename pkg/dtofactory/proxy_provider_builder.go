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
	propName := "IP"
	propVal := eb.entityId
	entityBuilder.WithProperty(getEntityPropertyNameValue(propName, propVal))

	// set sold commodities
	// Setting commodities and properties in the entity builder using the supply chain template
	supplyChainNode := eb.supplyChainNode

	supportedComms := supplyChainNode.SupportedComms
	supportedAccessComms := supplyChainNode.SupportedAccessComms

	var soldCommodities []*proto.CommodityDTO
	for commType, commVal := range supportedComms {
		if commVal.Key != nil {
			key := id //using the provider id as the key
			soldCommodities = append(soldCommodities, createCommodityWithKey(commType, key))
		} else {
			soldCommodities = append(soldCommodities, createCommodity(commType))
		}
	}
	for commType := range supportedAccessComms {
		key := id //using the provider id as the key
		soldCommodities = append(soldCommodities, createCommodityWithKey(commType, key))
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
	logDebug(fmt.Printf, protobuf.MarshalTextString(dto))

	glog.Infof("proxy provider dto: %++v\n", dto)

	return dto, nil
}
