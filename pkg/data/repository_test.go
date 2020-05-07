package data

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"testing"
)

func TestMultipleDIFEntitiesForApplication(t *testing.T) {
	var entities []*data.DIFEntity
	appMap := map[string]*data.DIFEntity{}
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("app-%d", i)

		entity := &data.DIFEntity{
			UID:  id,
			Type: "application",
			Name: id,
			HostedOn: &data.DIFHostedOn{
				HostType:  []data.DIFHostType{"virtualMachine"},
				IPAddress: "1.1.1.1",
				HostUuid:  "",
			},
			MatchingIdentifiers: &data.DIFMatchingIdentifiers{
				IPAddress: "1.1.1.1"},

			Metrics: nil,
		}
		appMap[id] = entity
		entities = append(entities, entity)
	}

	var entities2 []*data.DIFEntity
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("app-%d", i)
		entity := &data.DIFEntity{
			UID:     id,
			Type:    "application",
			Name:    id,
			Metrics: nil,
		}

		partOf := &data.DIFPartOf{
			ParentEntity: "service",
			UniqueId:     "service-1",
		}
		entity.PartOf = []*data.DIFPartOf{partOf}

		entities2 = append(entities2, entity)
	}

	svcEntity := &data.DIFEntity{
		UID:  "service-1",
		Type: "service",
		Name: "service-1",

		Metrics: nil,
	}
	svcMap := map[string]*data.DIFEntity{}
	svcMap["service-1"] = svcEntity

	var parsedEntities []*data.DIFEntity
	parsedEntities = append(parsedEntities, entities...)
	parsedEntities = append(parsedEntities, entities2...)
	parsedEntities = append(parsedEntities, svcEntity)

	repository := NewDIFRepository()
	repository.InitRepository(parsedEntities)

	assert.True(t, len(repository.EntityMap) == 2)

	appEntities := repository.EntityMap[APPLICATION]
	assert.True(t, len(appEntities) == 2)

	svcEntities := repository.EntityMap[SERVICE]
	assert.True(t, len(svcEntities) == 1)

	for appId, appEntity := range appEntities {
		_, exists := appMap[appId]
		assert.True(t, exists)
		assert.True(t, len(appEntity.GetProviders()) == 0)
	}

	for svcId, svcEntity := range svcEntities {
		_, exists := svcMap[svcId]
		assert.True(t, exists)

		providers := svcEntity.GetProviders()
		assert.True(t, len(svcEntity.GetProviders()) == 1)

		for key, val := range providers {
			assert.EqualValues(t, APPLICATION, key)
			assert.True(t, contains(val, "app-1"))
			assert.True(t, contains(val, "app-2"))
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestApplicationHostByUUID(t *testing.T) {
	var entities []*data.DIFEntity
	appMap := map[string]*data.DIFEntity{}
	hostUUID := "host-1"
	hostType := data.DIFHostType(VM)
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("app-%d", i)

		entity := &data.DIFEntity{
			UID:  id,
			Type: "application",
			Name: id,
			HostedOn: &data.DIFHostedOn{
				HostType:  []data.DIFHostType{hostType},
				IPAddress: "",
				HostUuid:  hostUUID,
			},
		}

		appMap[id] = entity
		entities = append(entities, entity)
	}

	var parsedEntities []*data.DIFEntity
	parsedEntities = append(parsedEntities, entities...)

	repository := NewDIFRepository()
	repository.InitRepository(parsedEntities)

	//map[application:map[app-1:0xc0000d24b0 app-2:0xc0000d2500]]
	//map[application:map[app-1:map[1.1.1.1:virtualMachine] app-2:map[1.1.1.1:virtualMachine]]]

	appEntities := repository.EntityMap[APPLICATION]
	for appId := range appEntities {
		appEntity := appEntities[appId]

		providers := appEntity.GetExternalProviderByUUID()
		assert.EqualValues(t, 1, len(providers))
		for pType, pMap := range providers {
			assert.EqualValues(t, VM, pType)
			for pId := range pMap {
				assert.EqualValues(t, hostUUID, pId)
			}
		}
	}
}

func TestApplicationHostByIP(t *testing.T) {
	var entities []*data.DIFEntity
	appMap := map[string]*data.DIFEntity{}
	hostIP := "1.1.1.1"
	hostType := data.DIFHostType(VM)
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("app-%d", i)

		entity := &data.DIFEntity{
			UID:  id,
			Type: "application",
			Name: id,
			HostedOn: &data.DIFHostedOn{
				HostType:  []data.DIFHostType{hostType},
				IPAddress: hostIP,
				HostUuid:  "",
			},
		}

		appMap[id] = entity
		entities = append(entities, entity)
	}

	var parsedEntities []*data.DIFEntity
	parsedEntities = append(parsedEntities, entities...)

	repository := NewDIFRepository()
	repository.InitRepository(parsedEntities)

	//map[application:map[app-1:0xc0000d24b0 app-2:0xc0000d2500]]
	//map[application:map[app-1:map[1.1.1.1:virtualMachine] app-2:map[1.1.1.1:virtualMachine]]]

	appEntities := repository.EntityMap[APPLICATION]
	for appId := range appEntities {
		appEntity := appEntities[appId]

		providers := appEntity.GetExternalProviderByIP()
		assert.EqualValues(t, 1, len(providers))
		for pType, pMap := range providers {
			assert.EqualValues(t, VM, pType)
			for pId := range pMap {
				assert.EqualValues(t, hostIP, pId)
			}
		}
	}
}

func TestApplicationMultipleHostTypesByIP(t *testing.T) {
	var entities []*data.DIFEntity
	appMap := map[string]*data.DIFEntity{}
	hostIP := "1.1.1.1"
	hostType := data.DIFHostType(VM)
	containerType := data.DIFHostType(CONTAINER)
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("app-%d", i)

		entity := &data.DIFEntity{
			UID:  id,
			Type: "application",
			Name: id,
			HostedOn: &data.DIFHostedOn{
				HostType:  []data.DIFHostType{hostType, containerType},
				IPAddress: hostIP,
				HostUuid:  "",
			},
		}
		appMap[id] = entity
		entities = append(entities, entity)
	}

	var parsedEntities []*data.DIFEntity
	parsedEntities = append(parsedEntities, entities...)

	repository := NewDIFRepository()
	repository.InitRepository(parsedEntities)

	//map[application:map[app-1:0xc0000d24b0 app-2:0xc0000d2500]]
	//ExternalProvidersMapByIP:
	//	map[application:
	//		map[app-1:map[container:[1.1.1.1] virtualMachine:[1.1.1.1]]
	//			app-2:map[container:[1.1.1.1] virtualMachine:[1.1.1.1]]]]
	//ExternalProvidersMapByUUID: map[]

	hostTypes := []DIFEntityType{VM, CONTAINER}
	appEntities := repository.EntityMap[APPLICATION]
	for appId := range appEntities {
		appEntity := appEntities[appId]

		providersByIP := appEntity.GetExternalProviderByIP()
		for _, hostType := range hostTypes {
			_, exists := providersByIP[hostType]
			assert.True(t, exists)

			idMap := providersByIP[hostType]
			assert.True(t, len(idMap) == 1)
		}

		providersByUUID := appEntity.GetExternalProviderByUUID()
		assert.True(t, len(providersByUUID) == 0)
	}
}

func TestApplicationMultipleHostTypesByUUID(t *testing.T) {
	var entities []*data.DIFEntity
	appMap := map[string]*data.DIFEntity{}
	hostID := "host-1"
	hostType := data.DIFHostType(VM)
	containerType := data.DIFHostType(CONTAINER)
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("app-%d", i)

		entity := &data.DIFEntity{
			UID:  id,
			Type: "application",
			Name: id,
			HostedOn: &data.DIFHostedOn{
				HostType: []data.DIFHostType{hostType, containerType},
				HostUuid: hostID,
			},
		}

		appMap[id] = entity
		entities = append(entities, entity)
	}

	var parsedEntities []*data.DIFEntity
	parsedEntities = append(parsedEntities, entities...)

	repository := NewDIFRepository()
	repository.InitRepository(parsedEntities)

	//map[application:map[app-1:0xc0000d24b0 app-2:0xc0000d2500]]
	//ExternalProvidersMapByUUID:
	//	map[application:
	//		map[app-1:map[container:[host-1] virtualMachine:[host-1]]
	//			app-2:map[container:[host-1] virtualMachine:[host-1]]]]

	hostTypes := []DIFEntityType{VM, CONTAINER}
	appEntities := repository.EntityMap[APPLICATION]
	for appId := range appEntities {
		appEntity := appEntities[appId]

		providersByUUID := appEntity.GetExternalProviderByUUID()
		for _, hostType := range hostTypes {
			_, exists := providersByUUID[hostType]
			assert.True(t, exists)

			idMap := providersByUUID[hostType]
			assert.True(t, len(idMap) == 1)
		}

		providersByIP := appEntity.GetExternalProviderByIP()
		assert.True(t, len(providersByIP) == 0)
	}
}
