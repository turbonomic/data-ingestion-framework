package dtofactory

import (
	"github.com/golang/glog"
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type CommodityKeyBuilder struct {
	entityType proto.EntityDTO_EntityType
	difEntity  *data.BasicDIFEntity
	scope      string
}

func NewCommodityKeyBuilder(
	entityType proto.EntityDTO_EntityType, difEntity *data.BasicDIFEntity) *CommodityKeyBuilder {
	return &CommodityKeyBuilder{
		entityType: entityType,
		difEntity:  difEntity,
	}
}

func (kb *CommodityKeyBuilder) SetScope(scope string) *CommodityKeyBuilder {
	kb.scope = scope
	return kb
}

func (kb *CommodityKeyBuilder) GetSoldCommKey() (keys []string) {
	defer func() {
		if len(keys) == 0 {
			keys = append(keys, "")
		}
		glog.V(4).Infof("Getting sold commodity key for %v %v: %v",
			kb.difEntity.EntityType, kb.difEntity.EntityId, keys)
	}()
	keySupplierTypes, exists := registration.SoldCommKeySupplierMapping[kb.entityType]
	if !exists || keySupplierTypes.Cardinality() == 0 {
		keys = append(keys, kb.difEntity.EntityId)
		return
	}
	// Check consumers
	for cType, cIds := range kb.difEntity.GetConsumers() {
		eType := EntityType(cType)
		if eType == nil {
			continue
		}
		// Key should come from the consumer of this type
		if keySupplierTypes.Contains(*eType) {
			keys = append(keys, cIds...)
		}
	}
	return
}

func (kb *CommodityKeyBuilder) GetBoughtCommKey(internalProvider bool) (keys []string) {
	defer func() {
		if len(keys) == 0 {
			keys = append(keys, "")
		}
		if internalProvider {
			glog.V(4).Infof("Getting bought commodity key for internal provider %v %v: %v",
				kb.difEntity.EntityType, kb.difEntity.EntityId, keys)
		} else {
			glog.V(4).Infof("Getting bought commodity key for external provider %v %v: %v",
				kb.difEntity.EntityType, kb.difEntity.EntityId, keys)
		}
	}()
	keySupplierTypes, exists := registration.BoughtCommKeySupplierMapping[kb.entityType]
	if !exists || keySupplierTypes.Cardinality() == 0 {
		keys = append(keys, kb.difEntity.EntityId)
		return
	}
	if internalProvider {
		// Check providers
		for pType, pIds := range kb.difEntity.GetProviders() {
			eType := EntityType(pType)
			if eType == nil {
				continue
			}
			// Key should come from the provider of this type
			if keySupplierTypes.Contains(*eType) {
				keys = append(keys, pIds...)
			}
		}
		return
	}
	// Check external providers by IP
	for pType, pMap := range kb.difEntity.HostsByIP {
		eType := EntityType(pType)
		if eType == nil {
			continue
		}
		if keySupplierTypes.Contains(*eType) {
			for pId := range pMap {
				key := getKey(*eType, pId, kb.scope)
				keys = append(keys, key)
			}
		}
	}
	// Check external providers by UUID
	for pType, pMap := range kb.difEntity.HostsByUUID {
		eType := EntityType(pType)
		if eType == nil {
			continue
		}
		if keySupplierTypes.Contains(*eType) {
			for pId := range pMap {
				key := getKey(*eType, pId, kb.scope)
				keys = append(keys, key)
			}
		}
	}
	return
}
