package conf

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"log"
	"testing"
)

var (
	BIZ_APP_NODE = "supplyChainNode:\n" +
		"  - templateClass: BUSINESS_APPLICATION\n" +
		"    templateType: BASE\n" +
		"    templatePriority: -1\n" +
		"    commodityBought:\n" +
		"    - key:\n" +
		"        templateClass: SERVICE\n" +
		"        providerType: LAYERED_OVER\n" +
		"        cardinalityMax: 2147483647\n" +
		"        cardinalityMin:\n" +
		"      value:\n" +
		"        - commodityType: TRANSACTION\n" +
		"          key: key-placeholder\n" +
		"        - commodityType: RESPONSE_TIME\n" +
		"          key: key-placeholder\n" +
		"        - commodityType: APPLICATION\n" +
		"    - key:\n" +
		"        templateClass: BUSINESS_TRANSACTION\n" +
		"        providerType: LAYERED_OVER\n" +
		"        cardinalityMax: 2147483647\n" +
		"        cardinalityMin:\n" +
		"      value:\n" +
		"        - commodityType: TRANSACTION\n" +
		"          key: key-placeholder\n" +
		"        - commodityType: RESPONSE_TIME\n" +
		"          key: key-placeholder\n" +
		"        - commodityType: APPLICATION\n"
)

func TestSupplyChainConfig(t *testing.T) {

	var data []byte
	data = []byte(BIZ_APP_NODE)
	var sc SupplyChainConfig
	err := yaml.Unmarshal([]byte(data), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	assert.True(t, err == nil)
	assert.True(t, len(sc.Nodes) > 0)

	entityType := BUSINESS_APPLICATION
	var nc *NodeConfig
	for _, node := range sc.Nodes {
		if node.TemplateClass == entityType {
			nc = node
			break
		}
	}

	assert.True(t, (nc != nil))
	assert.True(t, nc.CommoditySoldList == nil)
	assert.True(t, len(nc.CommodityBoughtList) == 2)

	providerMap := make(map[string]*CommodityBoughtConfig)
	for _, bought := range nc.CommodityBoughtList {
		assert.True(t, bought.Provider != nil)
		assert.True(t, bought.Provider.TemplateClass != nil)
		providerMap[*bought.Provider.TemplateClass] = bought
	}

	assert.True(t, providerMap[SERVICE] != nil)
	svcProvider := providerMap[SERVICE]
	assert.True(t, len(svcProvider.Comms) == 3)

	assert.True(t, providerMap[BUSINESS_TRANSACTION] != nil)
	bizTransProvider := providerMap[BUSINESS_TRANSACTION]
	assert.True(t, len(bizTransProvider.Comms) == 3)

	commKey := "key-placeholder"
	emptyKey := ""
	commMap := map[string]string{
		TRANSACTION:   commKey,
		RESPONSE_TIME: commKey,
		APPLICATION:   emptyKey,
	}

	for _, bought := range svcProvider.Comms {
		comm := *bought.CommodityType
		if _, exists := commMap[comm]; !exists {
			assert.Fail(t, "Comm %s is not expected", comm)
		} else {
			if bought.Key != nil {
				assert.EqualValues(t, *bought.Key, commMap[comm])
			}
		}
	}
	for _, bought := range bizTransProvider.Comms {
		comm := *bought.CommodityType
		if _, exists := commMap[comm]; !exists {
			assert.Fail(t, "Comm %s is not expected", comm)
		} else {
			if bought.Key != nil {
				assert.EqualValues(t, *bought.Key, commMap[comm])
			}
		}
	}
}

const (
	SERVICE               string = "SERVICE"
	BUSINESS_APPLICATION  string = "BUSINESS_APPLICATION"
	BUSINESS_TRANSACTION  string = "BUSINESS_TRANSACTION"
	APPLICATION_COMPONENT string = "APPLICATION_COMPONENT"
	VIRTUAL_MACHINE       string = "VIRTUAL_MACHINE"
	DATABASE_SERVER       string = "DATABASE_SERVER"

	TRANSACTION     string = "TRANSACTION"
	RESPONSE_TIME   string = "RESPONSE_TIME"
	APPLICATION     string = "APPLICATION"
	COLLECTION_TIME string = "COLLECTION_TIME"
	THREADS         string = "THREADS"
	HEAP            string = "HEAP"
	VCPU            string = "VCPU"
	VMEM            string = "VMEM"

	LIST_STRING string = "LIST_STRING"
	STRING      string = "STRING"

	LAYERED_OVER string = "LAYERED_OVER"
	HOSTING      string = "HOSTING"
)

var (
	SERVICE_NODE = "supplyChainNode:\n" +
		" - templateClass: SERVICE\n" +
		"   templateType: BASE\n" +
		"   templatePriority: -1\n" +
		"   commoditySold:\n" +
		"     - commodityType: TRANSACTION\n" +
		"       key: key-placeholder\n" +
		"     - commodityType: RESPONSE_TIME\n" +
		"       key: key-placeholder\n" +
		"     - commodityType: APPLICATION\n" +
		"   commodityBought:\n" +
		"     - key:\n" +
		"         templateClass: APPLICATION_COMPONENT\n" +
		"         providerType: LAYERED_OVER\n" +
		"         cardinalityMax: 2147483647\n" +
		"         cardinalityMin: 0\n" +
		"       value:\n" +
		"         - commodityType: TRANSACTION\n" +
		"           key: key-placeholder\n" +
		"         - commodityType: RESPONSE_TIME\n" +
		"           key: key-placeholder\n" +
		"         - commodityType: APPLICATION\n" +
		"     - key:\n" +
		"         templateClass: DATABASE_SERVER\n" +
		"         providerType: LAYERED_OVER\n" +
		"         cardinalityMax: 2147483647\n" +
		"         cardinalityMin: 0\n" +
		"       value:\n" +
		"         - commodityType: TRANSACTION\n" +
		"           key: key-placeholder\n" +
		"         - commodityType: RESPONSE_TIME\n" +
		"           key: key-placeholder\n" +
		"         - commodityType: APPLICATION\n" +
		"   mergedEntityMetaData:\n" +
		"     keepStandalone: false\n" +
		"     matchingMetadata:\n" +
		"       returnType: STRING\n" +
		"       matchingData:\n" +
		"         - matchingProperty:\n" +
		"             propertyName: IP\n" +
		"       externalEntityReturnType: LIST_STRING\n" +
		"       externalEntityMatchingProperty:\n" +
		"         - matchingProperty:\n" +
		"             propertyName: IP\n" +
		"           delimiter: \",\"\n" +
		"     commoditiesSold:\n" +
		"       - RESPONSE_TIME\n" +
		"       - TRANSACTION\n" +
		"     commoditiesBought:\n" +
		"       - providerType: APPLICATION_COMPONENT\n" +
		"         commodityMetadata:\n" +
		"           - RESPONSE_TIME\n" +
		"           - TRANSACTION\n"
)

var transComm = TRANSACTION
var rtComm = RESPONSE_TIME
var appComm = APPLICATION
var appEntity = APPLICATION_COMPONENT
var dbEntity = DATABASE_SERVER
var relLayeredOver = LAYERED_OVER
var keyStr = "key-placeholder"

var NodeSoldMap = map[string]*CommodityConfig{
	TRANSACTION:   {CommodityType: &transComm, Key: &keyStr},
	RESPONSE_TIME: {CommodityType: &rtComm, Key: &keyStr},
	APPLICATION:   {CommodityType: &appComm},
}

var NodeBoughtMap = map[string]*CommodityBoughtConfig{
	APPLICATION_COMPONENT: {
		Provider: &ProviderConfig{
			TemplateClass:  &appEntity,
			ProviderType:   &relLayeredOver,
			CardinalityMax: 2147483647,
			CardinalityMin: 0,
		},
		Comms: []*CommodityConfig{
			{CommodityType: &appComm},
			{CommodityType: &rtComm, Key: &keyStr},
			{CommodityType: &transComm, Key: &keyStr},
		},
	},

	DATABASE_SERVER: {
		Provider: &ProviderConfig{
			TemplateClass:  &dbEntity,
			ProviderType:   &relLayeredOver,
			CardinalityMax: 2147483647,
			CardinalityMin: 0,
		},
		Comms: []*CommodityConfig{
			{CommodityType: &appComm},
			{CommodityType: &rtComm, Key: &keyStr},
			{CommodityType: &transComm, Key: &keyStr},
		},
	},
}

var NodeMergedEntityMetadata = &MergedEntityMetaDataConfig{
	KeepInTopology: false,
	MatchingMetadata: &MatchingMetadataConfig{
		ReturnType:               STRING,
		ExternalEntityReturnType: LIST_STRING,
		MatchingDataList: []*MatchingDataConfig{
			{
				MatchingProperty: &MatchingPropertyConfig{
					PropertyName: "IP",
				},
				Delimiter: ",",
			},
		},
		ExternalEntityMatchingPropertyList: []*MatchingDataConfig{
			{
				MatchingProperty: &MatchingPropertyConfig{
					PropertyName: "IP",
				},
				Delimiter: ",",
			},
		},
	},
	CommSold: []string{RESPONSE_TIME, TRANSACTION},
	CommoditiesBought: []*MergedMetadataBoughtConfig{
		{
			Provider: APPLICATION_COMPONENT,
			Comm:     []string{RESPONSE_TIME, TRANSACTION},
		},
	},
}

func TestSupplyChainBoughtComms(t *testing.T) {
	var data []byte
	data = []byte(SERVICE_NODE)
	var sc SupplyChainConfig
	err := yaml.Unmarshal([]byte(data), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	entityType := SERVICE
	var nc *NodeConfig
	for _, node := range sc.Nodes {
		if node.TemplateClass == entityType {
			nc = node
			break
		}
	}
	assert.True(t, (nc != nil))

	boughtMap := make(map[string]*CommodityBoughtConfig)
	for _, cb := range nc.CommodityBoughtList {
		assert.True(t, cb.Provider != nil)
		assert.True(t, cb.Provider.TemplateClass != nil)
		assert.True(t, cb.Comms != nil)
		assert.True(t, len(cb.Comms) > 0)
		boughtMap[*cb.Provider.TemplateClass] = cb
		assert.ElementsMatch(t, cb.Comms, NodeBoughtMap[*cb.Provider.TemplateClass].Comms)
	}

	soldMap := make(map[string]*CommodityConfig)
	for _, s := range nc.CommoditySoldList {
		assert.True(t, s.CommodityType != nil)
		soldMap[*s.CommodityType] = s
		assert.EqualValues(t, NodeSoldMap[*s.CommodityType], s)
	}

	m := nc.MergedEntityMetaData
	assert.True(t, m != nil)
	assert.ElementsMatch(t, m.CommSold, NodeMergedEntityMetadata.CommSold)
	assert.EqualValues(t, m.CommoditiesBought, NodeMergedEntityMetadata.CommoditiesBought)
	assert.EqualValues(t, m.KeepInTopology, NodeMergedEntityMetadata.KeepInTopology)
	assert.True(t, m.MatchingMetadata != nil)
	assert.EqualValues(t, m.MatchingMetadata.ReturnType, NodeMergedEntityMetadata.MatchingMetadata.ReturnType)
	assert.EqualValues(t, m.MatchingMetadata.ExternalEntityReturnType, NodeMergedEntityMetadata.MatchingMetadata.ExternalEntityReturnType)

	assert.EqualValues(t, len(m.MatchingMetadata.MatchingDataList), 1)
	md := m.MatchingMetadata.MatchingDataList[0]
	assert.True(t, md.MatchingProperty != nil)
	assert.EqualValues(t, md.MatchingProperty.PropertyName,
		NodeMergedEntityMetadata.MatchingMetadata.MatchingDataList[0].MatchingProperty.PropertyName)

	assert.EqualValues(t, len(m.MatchingMetadata.ExternalEntityMatchingPropertyList), 1)
	extMd := m.MatchingMetadata.MatchingDataList[0]
	assert.True(t, extMd.MatchingProperty != nil)
	assert.EqualValues(t, extMd.MatchingProperty.PropertyName,
		NodeMergedEntityMetadata.MatchingMetadata.MatchingDataList[0].MatchingProperty.PropertyName)
}
