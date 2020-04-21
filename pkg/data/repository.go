package data

import (
	"github.com/golang/glog"
	difdata "github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

type ProviderMap map[string]DIFEntityType

type ExternalProviderMap map[DIFEntityType][]string

type DIFRepository struct {

	// Map of entities by EntityType, EntityId
	EntityMap map[DIFEntityType]map[string]*BasicDIFEntity

	// Maps for connecting entities to internal providers as specified by the 'partOf' section of the DIF Entity JSON
	// The entity with the partOf section is a provider for the entities listed in the partOf section.
	// In the example below, application-10.233.90.212 is provider for service 'topology-processor'
	// {
	//	"uniqueId": "10.233.90.212",
	//	"type": "application",
	//	"name": "10.233.90.212",
	//	"hostedOn": null,
	//	"partOf": [
	//	 {
	//		"entity": "service",
	//		"uniqueId": "topology-processor"
	//	 }
	//	],
	// }
	// Ex:
	// Service 	-> S1 	-> [ A1 -> Application]
	//		   	-> S2 	-> [ A2 -> Application]
	// BizTrans -> BT1 	-> [ S1 -> Service]
	//		   	-> BT2 	-> [ S2 -> Service]
	// BizApp 	-> BA1 	-> [ S1 -> Service]
	//			   BA1 	-> [ S2 -> Service]
	// 			   BA1 	-> [ BT1 -> BizTrans]
	//		       BA1 	-> [ BT2 -> BizTrans]
	ConsumerToProviderMap map[DIFEntityType]map[string]ProviderMap
	ProviderToConsumerMap map[DIFEntityType]map[string]ProviderMap

	// Maps for connecting entities to external providers as specified in the 'hostedOn' section of the DIF Entity JSON
	// Host is specified using IP
	// In the example below, the application is hosted by a VM with the specified IP address
	// {
	//   "uniqueId": "10.10.172.113",
	//   "type": "application",
	//   "name": "10.10.172.113",
	//   "hostedOn": {
	//    "hostType": [
	//     "virtualMachine"
	//    ],
	//    "ipAddress": "10.10.172.113",
	//    "hostUuid": ""
	//   },
	//  }
	// Ex:
	//	Application -> A1: map[VM:IP1 map[CONTAINER:IP1]
	//				-> A2: map[VM:IP1]
	//
	ExternalProvidersMapByIP map[DIFEntityType]map[string]ExternalProviderMap

	// Maps for connecting entities to external providers as specified in the 'hostedOn' section of the DIF Entity JSON
	// Host is specified using UUID
	// Ex:
	//	Application -> A1: map[VM:UUID1]
	//				-> A2: map[VM:UUID1]
	//
	ExternalProvidersMapByUUID map[DIFEntityType]map[string]ExternalProviderMap

	// Map of external providers without any DIF Json data - used to create proxy entity DTOs by the DiscoveryCliebt
	ExternalProxyProvidersByIP   map[DIFEntityType][]string
	ExternalProxyProvidersByUUID map[DIFEntityType][]string
}

func NewDIFRepository() *DIFRepository {
	return &DIFRepository{
		EntityMap:                    make(map[DIFEntityType]map[string]*BasicDIFEntity),
		ConsumerToProviderMap:        make(map[DIFEntityType]map[string]ProviderMap),
		ExternalProvidersMapByIP:     make(map[DIFEntityType]map[string]ExternalProviderMap),
		ExternalProvidersMapByUUID:   make(map[DIFEntityType]map[string]ExternalProviderMap),
		ExternalProxyProvidersByIP:   make(map[DIFEntityType][]string),
		ExternalProxyProvidersByUUID: make(map[DIFEntityType][]string),
	}
}

// Initialize the structures
func (r *DIFRepository) InitRepository(parsedEntities []*difdata.DIFEntity) {
	r.addEntities(parsedEntities)
	r.addEntityConnections()
	r.addHostedByConnections()
	r.getProxyProviders()
}

// Add list of DIF entities that are parsed from the a DIF JSON source
// Entities are consolidated by entity type and entity ID
func (r *DIFRepository) addEntities(parsedEntities []*difdata.DIFEntity) {
	if parsedEntities == nil || len(parsedEntities) <= 0 {
		return
	}

	entityMap := r.EntityMap
	for _, entity := range parsedEntities {
		//glog.V(2).Infof("%++v", entity)
		eType := ParseEntityType(entity.Type)
		if eType == "" {
			glog.Errorf("Invalid entity type for entity: %v\n", entity)
			continue
		}

		eId := entity.UID
		if eId == "" {
			glog.Errorf("Invalid entity ID for entity: %v\n", entity)
			continue
		}

		if _, exists := entityMap[eType]; !exists {
			entityMap[eType] = make(map[string]*BasicDIFEntity)
		}
		eMap := entityMap[eType]

		var difEntity *BasicDIFEntity
		if _, exists := eMap[eId]; !exists {
			eMap[eId] = NewBasicDIFEntity(eType, eId)
		} else {
			glog.V(4).Infof("Adding parsed entity to an existing entity %s:%s ---> %v", eType, eId, eMap[eId])
		}
		difEntity = eMap[eId]

		// save parsed entity
		eList := difEntity.GetDIFEntities()
		eList = append(eList, entity)
		difEntity.SetDIFEntities(eList)

		// save provider-consumer connections
		if entity.PartOf != nil {
			partOfList := entity.PartOf
			for _, partOf := range partOfList {
				cType := ParseEntityType(partOf.ParentEntity)
				if cType == "" {
					glog.Errorf("Invalid entity type for partOf: %v\n", partOf.ParentEntity)
					continue
				}
				cId := partOf.UniqueId
				difEntity.SetConsumer(cType, cId)

				// save the consumer connection, the entity will be added as provider
				pType := eType
				pId := eId
				r.addConsumerProviderConnection(pType, cType, pId, cId)

			}
		}

		// save external provider connections
		if entity.HostedOn != nil {
			for _, pTypeStr := range entity.HostedOn.HostType {
				pType := ParseEntityType(string(pTypeStr))
				if pType == "" {
					glog.Errorf("Invalid entity type for HostedBy: %v\n", pTypeStr)
					continue
				}
				cType := eType
				cId := eId
				if entity.HostedOn.IPAddress != "" {
					pId := entity.HostedOn.IPAddress
					r.addExternalProviderConnection(pType, cType, pId, cId, true)
				}
				if entity.HostedOn.HostUuid != "" {
					pId := entity.HostedOn.HostUuid
					r.addExternalProviderConnection(pType, cType, pId, cId, false)
				}
			}
		}
		eMap[eId] = difEntity
	}
}

func (r *DIFRepository) addConsumerProviderConnection(pType, cType DIFEntityType, pId, cId string) {
	//glog.Infof("%v::%s Adding provider %v %s\n", cType, cId, pType, pId)
	consumerToProviderMap := r.ConsumerToProviderMap
	if _, exists := consumerToProviderMap[cType]; !exists {
		consumerToProviderMap[cType] = make(map[string]ProviderMap)
	}
	cMap := consumerToProviderMap[cType]
	if _, exists := cMap[cId]; !exists {
		cMap[cId] = make(map[string]DIFEntityType)
	}

	pMap := cMap[cId]
	pMap[pId] = pType
}

func (r *DIFRepository) addExternalProviderConnection(pType, cType DIFEntityType, pId, cId string, isIP bool) {
	//glog.Infof("%v::%s Adding host %v %s\n", cType, cId, pType, pId)
	var externalProviderMap map[DIFEntityType]map[string]ExternalProviderMap
	if isIP {
		externalProviderMap = r.ExternalProvidersMapByIP
	} else {
		externalProviderMap = r.ExternalProvidersMapByUUID
	}

	if _, exists := externalProviderMap[cType]; !exists {
		externalProviderMap[cType] = make(map[string]ExternalProviderMap)
	}
	cMap := externalProviderMap[cType]
	if _, exists := cMap[cId]; !exists {
		cMap[cId] = make(map[DIFEntityType][]string)
	}

	pMap := cMap[cId]
	if _, exists := pMap[pType]; !exists {
		pMap[pType] = []string{}
	}
	pMap[pType] = append(pMap[pType], pId)
}

// Go over the internal providers map and add the providers in the entities
func (r *DIFRepository) addEntityConnections() {
	for cType, cMap := range r.ConsumerToProviderMap {
		for cId, pMap := range cMap {
			//glog.Infof("Consumer %s::%s ---> ", cType, cId)
			consumerEntity := r.EntityMap[cType][cId]
			if consumerEntity == nil {
				glog.Errorf("NULL DIF Entity for %s::%s", cType, cId)
				continue
			}
			for pId, pType := range pMap {
				//glog.Infof("		provider %s::%s ---> ", pType, pId)
				consumerEntity.SetProvider(pType, pId)
			}
		}
	}
}

// Go over the external providers map and add the providers in the entitities
func (r *DIFRepository) addHostedByConnections() {
	r.addExternalProvider(true)
	r.addExternalProvider(false)
}

func (r *DIFRepository) addExternalProvider(byIP bool) {
	var externalProviderMap map[DIFEntityType]map[string]ExternalProviderMap
	if byIP {
		externalProviderMap = r.ExternalProvidersMapByIP
	} else {
		externalProviderMap = r.ExternalProvidersMapByUUID
	}

	for cType, cMap := range externalProviderMap {
		for cId, pMap := range cMap {
			consumerEntity := r.EntityMap[cType][cId]
			if consumerEntity == nil {
				glog.Errorf("NULL consumer Entity for %s::%s", cType, cId)
				continue
			}
			for pType, pIds := range pMap {
				for _, pId := range pIds {
					var pEntity *BasicDIFEntity
					var hostEntities []*difdata.DIFEntity
					if _, exists := r.EntityMap[pType]; exists {
						pEntity = r.EntityMap[pType][pId]
						if pEntity != nil {
							hostEntities = pEntity.GetDIFEntities()
						}
					}
					consumerEntity.setExternalProvider(pType, pId, hostEntities, byIP)
				}
			}
		}
	}
}

func (r *DIFRepository) getProxyProviders() {
	r.getProxyProviderByIP()
	r.getProxyProviderByUUID()
	glog.V(3).Infof("ExternalProxyProvidersByIP: %v\n", r.ExternalProxyProvidersByIP)
	glog.V(3).Infof("ExternalProxyProvidersByUUID: %v\n", r.ExternalProxyProvidersByUUID)

}

func (r *DIFRepository) getProxyProviderByIP() {
	for _, cMap := range r.ExternalProvidersMapByIP {
		for _, pMap := range cMap {
			for pType, pIds := range pMap {
				for _, pId := range pIds {
					if _, exists := r.EntityMap[pType]; !exists {
						// host entity without DIF entity data
						if _, exists := r.ExternalProxyProvidersByIP[pType]; !exists {
							r.ExternalProxyProvidersByIP[pType] = []string{}
						}
						providerIds := r.ExternalProxyProvidersByIP[pType]
						foundId := false
						for _, id := range providerIds {
							if id == pId {
								foundId = true
								break
							}
						}
						if !foundId {
							providerIds = append(providerIds, pId)
							r.ExternalProxyProvidersByIP[pType] = providerIds
						}
					}
				}
			}
		}
	}
}

func (r *DIFRepository) getProxyProviderByUUID() {

	for _, cMap := range r.ExternalProvidersMapByUUID {
		for _, pMap := range cMap {
			for pType, pIds := range pMap {
				for _, pId := range pIds {
					if _, exists := r.EntityMap[pType]; !exists {
						// host entity without DIF entity data
						if _, exists := r.ExternalProxyProvidersByUUID[pType]; !exists {
							r.ExternalProxyProvidersByUUID[pType] = []string{}
						}
						providerIds := r.ExternalProxyProvidersByUUID[pType]
						foundId := false
						for _, id := range providerIds {
							if id == pId {
								foundId = true
								break
							}
						}
						if !foundId {
							providerIds = append(providerIds, pId)
							r.ExternalProxyProvidersByUUID[pType] = providerIds
						}
					}
				}
			}
		}
	}
}
