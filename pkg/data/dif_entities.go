package data

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

type IsDIFEntity interface {
	GetDIFEntities() []*data.DIFEntity
	GetProviders() map[DIFEntityType][]string
	SetDIFEntities(entityList []*data.DIFEntity)
	SetProvider(pType DIFEntityType, pId string)
}

type BasicDIFEntity struct {
	EntityId   string
	EntityType DIFEntityType
	// list of entities created from multiple DIF JSON entities
	entityList []*data.DIFEntity
	// Map of internal providers
	Providers map[DIFEntityType][]string //provider type providerIds
	Consumers map[DIFEntityType][]string //consumer type providerIds
	// Map of external providers
	HostsByIP   map[DIFEntityType]map[string][]*data.DIFEntity //host type, host ip,
	HostsByUUID map[DIFEntityType]map[string][]*data.DIFEntity //host type, host uuid,
}

func NewBasicDIFEntity(entityType DIFEntityType, entityId string) *BasicDIFEntity {
	return &BasicDIFEntity{
		EntityId:   entityId,
		EntityType: entityType,
		entityList: []*data.DIFEntity{},
		Providers:  make(map[DIFEntityType][]string),
		// Example:
		// map[virtualMachine: map[1.1.1.1:[...]]]
		HostsByIP:   make(map[DIFEntityType]map[string][]*data.DIFEntity),
		HostsByUUID: make(map[DIFEntityType]map[string][]*data.DIFEntity),
	}
}

func (entity *BasicDIFEntity) GetDIFEntities() []*data.DIFEntity {
	return entity.entityList
}

func (entity *BasicDIFEntity) SetDIFEntities(entityList []*data.DIFEntity) {
	entity.entityList = entityList
}

func (entity *BasicDIFEntity) SetProvider(pType DIFEntityType, pId string) {
	if entity.Providers == nil {
		entity.Providers = make(map[DIFEntityType][]string)
	}

	if _, exists := entity.Providers[pType]; !exists {
		entity.Providers[pType] = []string{}
	}

	pList := entity.Providers[pType]
	pList = append(pList, pId)
	entity.Providers[pType] = pList
}

func (entity *BasicDIFEntity) SetConsumer(cType DIFEntityType, cId string) {
	if entity.Consumers == nil {
		entity.Consumers = make(map[DIFEntityType][]string)
	}

	if _, exists := entity.Consumers[cType]; !exists {
		entity.Consumers[cType] = []string{}
	}

	cList := entity.Consumers[cType]
	cList = append(cList, cId)
	entity.Consumers[cType] = cList
}

// Returns map of providers by provider entity type -> list of provider IDs
func (entity *BasicDIFEntity) GetProviders() map[DIFEntityType][]string {
	return entity.Providers
}

// Returns map of consumers by consumer entity type -> list of consumer IDs
func (entity *BasicDIFEntity) GetConsumers() map[DIFEntityType][]string {
	return entity.Consumers
}

func (entity *BasicDIFEntity) setExternalProvider(pType DIFEntityType, pId string, hosts []*data.DIFEntity, byIP bool) {
	if entity.HostsByIP == nil {
		entity.HostsByIP = make(map[DIFEntityType]map[string][]*data.DIFEntity)
	}

	if byIP {
		if _, exists := entity.HostsByIP[pType]; !exists {
			entity.HostsByIP[pType] = make(map[string][]*data.DIFEntity)
		}
		pMap := entity.HostsByIP[pType]
		if _, exists := pMap[pId]; !exists {
			pMap[pId] = []*data.DIFEntity{}
		}
		pMap[pId] = hosts
		entity.HostsByIP[pType] = pMap

	} else {
		if _, exists := entity.HostsByUUID[pType]; !exists {
			entity.HostsByUUID[pType] = make(map[string][]*data.DIFEntity)
		}

		pMap := entity.HostsByUUID[pType]
		if _, exists := pMap[pId]; !exists {
			pMap[pId] = []*data.DIFEntity{}
		}
		pMap[pId] = hosts
		entity.HostsByUUID[pType] = pMap
	}
}

// Returns external provider map by entity type -> map of entityID -> external provider dif entity list
func (entity *BasicDIFEntity) GetExternalProviderByIP() map[DIFEntityType]map[string][]*data.DIFEntity {
	return entity.HostsByIP
}

// Returns external provider map by entity type -> map of entityID -> external provider dif entity list
func (entity *BasicDIFEntity) GetExternalProviderByUUID() map[DIFEntityType]map[string][]*data.DIFEntity {
	return entity.HostsByUUID
}
