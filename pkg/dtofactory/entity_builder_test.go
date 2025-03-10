package dtofactory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/registration"
	difdata "github.ibm.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
)

func createSupplyChainTemplates() (map[proto.EntityDTO_EntityType]*registration.SupplyChainNode, error) {
	// Load the supply chain conf(ig
	supplyChainConf := "../../configs/app-supply-chain-config.yaml"
	supplyChainConfig, err := conf.LoadSupplyChain(supplyChainConf)
	if err != nil {
		return nil, fmt.Errorf("error while parsing the supply chain config file %s: %+v", supplyChainConf, err)
	}

	supplyChain, err := registration.NewSupplyChain(supplyChainConfig)
	if err != nil {
		return nil, fmt.Errorf("error while parsing the supply chain config %+v", err)
	}
	supplyChainNodeMap := supplyChain.GetSupplyChainNodes()

	return supplyChainNodeMap, nil
}

func TestEntityBuilder(t *testing.T) {
	supplyChainNodeMap, err := createSupplyChainTemplates()
	if err != nil {
		t.Errorf("Failed to create supply chain template: %v", err)
	}
	scope := "test"
	ENTITY =
		"{" +
			"	\"type\": \"application\"," +
			"	\"uniqueId\": \"456\"," +
			"	\"name\": \"My App Name 2\"," +
			"	\"metrics\" : {" +
			RESPONSE_TIME +
			"]," +
			TRANSACTION +
			"]," +
			HEAP_ARRAY +
			"]" +
			"		}," +
			"	\"hostedOn\":{" +
			"		\"hostType\":[" +
			"			\"virtualMachine\"" +
			"		]," +
			"		\"ipAddress\":\"10.10.168.193\"," +
			"		\"hostUuid\":\"\"" +
			"	}" +
			"}"

	parsedDifEntity := parseEntity(ENTITY)
	eType := data.ParseEntityType(parsedDifEntity.Type)
	if eType == "" {
		t.Errorf("Invalid entity type for entity: %v", parsedDifEntity)
	}

	eId := parsedDifEntity.UID
	if eId == "" {
		t.Errorf("Invalid entity ID for entity: %v", parsedDifEntity)
	}

	difEntity := data.NewBasicDIFEntity(eType, eId, eId)
	difEntity.SetDIFEntities([]*difdata.DIFEntity{parsedDifEntity})

	vmHost := make(map[string][]*difdata.DIFEntity)
	vmHost["10.10.168.193"] = []*difdata.DIFEntity{}
	difEntity.HostsByIP[data.VM] = vmHost

	entityTypeStr := data.DIFEntityTypeToTemplateEntityStringMap[eType]
	entityType := registration.TemplateEntityTypeMap[entityTypeStr]

	eb := NewGenericEntityBuilder(entityType, difEntity,
		scope, true, supplyChainNodeMap[entityType])

	dto, err := eb.BuildEntity()

	if err != nil {
		t.Errorf("DTO BUILD ERROR %v", err)
	}
	// created from the built DTO
	testEntity := entityDTOToTestEntity(dto)
	var sold1 []proto.CommodityDTO_CommodityType
	for s := range testEntity.soldComms {
		sold1 = append(sold1, s)
	}
	var bought1 []proto.CommodityDTO_CommodityType
	var providers1 []proto.EntityDTO_EntityType
	for p, bc := range testEntity.boughtComms {
		providers1 = append(providers1, p)
		for b := range bc {
			bought1 = append(bought1, b)
		}
	}

	// expected entity properties
	expectedTestEntity := &TestEntity{
		id:          difEntity.EntityId,
		displayName: difEntity.EntityId,
		eType:       entityType,
		soldComms:   make(map[proto.CommodityDTO_CommodityType]*proto.CommodityDTO),
		boughtComms: make(map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]*proto.CommodityDTO),
		entityProps: make(map[string]string),
	}
	expectedTestEntity.soldComms[proto.CommodityDTO_RESPONSE_TIME] = &proto.CommodityDTO{}
	expectedTestEntity.soldComms[proto.CommodityDTO_TRANSACTION] = &proto.CommodityDTO{}
	expectedTestEntity.soldComms[proto.CommodityDTO_HEAP] = &proto.CommodityDTO{}
	expectedTestEntity.soldComms[proto.CommodityDTO_APPLICATION] = &proto.CommodityDTO{}

	comms := make(map[proto.CommodityDTO_CommodityType]*proto.CommodityDTO)
	expectedTestEntity.boughtComms[proto.EntityDTO_VIRTUAL_MACHINE] = comms
	comms[proto.CommodityDTO_APPLICATION] = &proto.CommodityDTO{}
	var sold2 []proto.CommodityDTO_CommodityType
	for s := range expectedTestEntity.soldComms {
		sold2 = append(sold2, s)
	}

	var bought2 []proto.CommodityDTO_CommodityType
	var providers2 []proto.EntityDTO_EntityType
	for p, bc := range expectedTestEntity.boughtComms {
		providers2 = append(providers2, p)
		for b := range bc {
			bought2 = append(bought2, b)
		}
	}

	assert.ElementsMatch(t, sold1, sold2)
	assert.ElementsMatch(t, bought1, bought2)
	assert.ElementsMatch(t, providers1, providers2)
}

type TestEntity struct {
	id          string
	displayName string
	eType       proto.EntityDTO_EntityType
	soldComms   map[proto.CommodityDTO_CommodityType]*proto.CommodityDTO
	providers   []proto.EntityDTO_EntityType
	boughtComms map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]*proto.CommodityDTO
	entityProps map[string]string
}

func entityDTOToTestEntity(dto *proto.EntityDTO) *TestEntity {
	testEntity := &TestEntity{
		id:          dto.GetId(),
		displayName: dto.GetDisplayName(),
		eType:       dto.GetEntityType(),
		soldComms:   make(map[proto.CommodityDTO_CommodityType]*proto.CommodityDTO),
		boughtComms: make(map[proto.EntityDTO_EntityType]map[proto.CommodityDTO_CommodityType]*proto.CommodityDTO),
		entityProps: make(map[string]string),
	}

	soldComms := dto.GetCommoditiesSold()
	for _, comm := range soldComms {
		testEntity.soldComms[*comm.CommodityType] = comm
	}

	boughtComms := dto.GetCommoditiesBought()
	for _, commBought := range boughtComms {
		provider := *commBought.ProviderType
		testEntity.providers = append(testEntity.providers, provider)
		testEntity.boughtComms[provider] = make(map[proto.CommodityDTO_CommodityType]*proto.CommodityDTO)

		bought := commBought.Bought
		for _, comm := range bought {
			testEntity.boughtComms[provider][*comm.CommodityType] = comm
		}
	}

	for _, entityProp := range dto.GetEntityProperties() {
		testEntity.entityProps[*entityProp.Name] = *entityProp.Value
	}

	return testEntity
}

func TestMergingMetadataProperty(t *testing.T) {
	supplyChainNodeMap, err := createSupplyChainTemplates()
	if err != nil {
		t.Errorf("Failed to create supply chain template: %v", err)
	}
	scope := "test"
	ENTITY =
		"{" +
			"	\"type\": \"application\"," +
			"	\"uniqueId\":\"10.233.90.114\"," +
			"	\"name\":\"10.233.90.114\"," +
			"	\"matchIdentifiers\":{" +
			"		\"ipAddress\":\"10.233.90.114\"" +
			"	 }," +
			"	\"metrics\" : {" +
			RESPONSE_TIME +
			" 		]," +
			TRANSACTION +
			"       ]" +
			"     }" +
			"}"

	parsedDifEntity := parseEntity(ENTITY)
	eType := data.ParseEntityType(parsedDifEntity.Type)
	if eType == "" {
		t.Errorf("Invalid entity type for entity: %v", parsedDifEntity)
	}

	eId := parsedDifEntity.UID
	if eId == "" {
		t.Errorf("Invalid entity ID for entity: %v", parsedDifEntity)
	}

	entityType := EntityType(eType)
	difEntity := data.NewBasicDIFEntity(eType, eId, eId)
	difEntity.SetDIFEntities([]*difdata.DIFEntity{parsedDifEntity})

	eb := NewGenericEntityBuilder(*entityType, difEntity,
		scope, true, supplyChainNodeMap[*entityType])

	dto, err := eb.BuildEntity()

	if err != nil {
		t.Errorf("DTO BUILD ERROR %v", err)
	}

	dtoProps := dto.GetEntityProperties()
	for _, dtoProp := range dtoProps {
		if dtoProp.GetName() == "IP" {
			assert.EqualValues(t, dtoProp.GetValue(), parsedDifEntity.MatchingIdentifiers.IPAddress)
		}
	}
}

// Test proxy application hosted on proxy VM:
// - HOST_IP appears in the entity properties
// - Provider VM's ID is not the same as the application's ID
// - Commodity key for the bought application is correctly set
func TestExternalLinkMetadataProperty(t *testing.T) {
	supplyChainNodeMap, err := createSupplyChainTemplates()
	if err != nil {
		t.Errorf("Failed to create supply chain template: %v", err)
	}

	scope := "test"
	ENTITY =
		"{" +
			"	\"type\": \"application\"," +
			"	\"uniqueId\":\"10.10.168.193\"," +
			"	\"name\":\"10.10.168.193\"," +
			"	\"metrics\" : {" +
			RESPONSE_TIME +
			" 		]," +
			TRANSACTION +
			"       ]" +
			"     }," +
			"	\"hostedOn\":{" +
			"		\"hostType\":[" +
			"			\"virtualMachine\"" +
			"		]," +
			"		\"ipAddress\":\"10.10.168.193\"" +
			"	}" +
			"}"

	parsedDifEntity := parseEntity(ENTITY)
	eType := data.ParseEntityType(parsedDifEntity.Type)
	if eType == "" {
		t.Errorf("Invalid entity type for entity: %v", parsedDifEntity)
	}

	eId := parsedDifEntity.UID
	if eId == "" {
		t.Errorf("Invalid entity ID for entity: %v", parsedDifEntity)
	}

	difEntity := data.NewBasicDIFEntity(eType, eId, eId)
	difEntity.SetDIFEntities([]*difdata.DIFEntity{parsedDifEntity})

	hostIP := parsedDifEntity.HostedOn.IPAddress
	vmHost := make(map[string][]*difdata.DIFEntity)
	vmHost[hostIP] = []*difdata.DIFEntity{}
	difEntity.HostsByIP[data.VM] = vmHost

	entityTypeStr := data.DIFEntityTypeToTemplateEntityStringMap[eType]
	entityType := registration.TemplateEntityTypeMap[entityTypeStr]

	eb := NewGenericEntityBuilder(entityType, difEntity,
		scope, true, supplyChainNodeMap[entityType])

	dto, err := eb.BuildEntity()

	if err != nil {
		t.Errorf("DTO BUILD ERROR %v", err)
	}

	dtoProps := dto.GetEntityProperties()
	for _, dtoProp := range dtoProps {
		if dtoProp.GetName() == "HOST_IP" {
			assert.EqualValues(t, dtoProp.GetValue(), hostIP)
		}
	}
	assert.NotEqual(t, dto.GetCommoditiesBought()[0].GetProviderId(), hostIP)
	assert.EqualValues(t, dto.GetCommoditiesBought()[0].GetProviderId(),
		getProxyEntityId(dto.GetCommoditiesBought()[0].GetProviderType(), dto.GetId(), scope))
	for _, boughtComms := range dto.GetCommoditiesBought() {
		for _, boughtComm := range boughtComms.GetBought() {
			if boughtComm.GetCommodityType() != proto.CommodityDTO_APPLICATION {
				continue
			}
			boughtCommKeys := NewCommodityKeyBuilder(dto.GetEntityType(), difEntity).
				SetScope(scope).
				GetBoughtCommKey(false)
			assert.EqualValues(t, boughtComm.GetKey(), boughtCommKeys[0])
			return
		}
	}
	assert.Fail(t, "Failed to find Application Commodity in %+v", dto)
}
