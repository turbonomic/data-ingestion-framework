package dtofactory

import (
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"strings"
)

type CommodityKeyBuilder struct {
	entityType proto.EntityDTO_EntityType
	difEntity  *data.BasicDIFEntity
}

func NewCommodityKeyBuilder(entityType proto.EntityDTO_EntityType,
	difEntity *data.BasicDIFEntity) *CommodityKeyBuilder {
	return &CommodityKeyBuilder{
		entityType: entityType,
		difEntity:  difEntity,
	}
}

func (kb *CommodityKeyBuilder) GetKey() *string {

	if _, exists := registration.KeySupplierMapping[kb.entityType]; !exists {
		return nil
	}

	difEntity := kb.difEntity

	keySupplierType := registration.KeySupplierMapping[kb.entityType]
	if keySupplierType == kb.entityType {
		return &difEntity.EntityId
	}

	// check in provider connections
	for pType, pIds := range difEntity.GetProviders() {
		eType := EntityType(pType)
		if eType == nil {
			continue
		}

		// Key should come from the provider of this type
		if keySupplierType == *eType {
			key := strings.Join(pIds, ",")
			return &key
		}
	}

	// check in consumer connections
	for cType, cIds := range difEntity.GetConsumers() {
		eType := EntityType(cType)
		if eType == nil {
			continue
		}
		// Key should come from the provider of this type
		if keySupplierType == *eType {
			key := strings.Join(cIds, ",")
			return &key
		}
	}

	return nil
}
