package registration

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/conf"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
	"gopkg.in/yaml.v2"
)

func TestSupplyChainNode(t *testing.T) {
	filename := "test_app_node.yaml"

	var supplyChainConfig *conf.SupplyChainConfig
	supplyChainConfig, err := conf.LoadSupplyChain(filename)
	if err != nil {
		fmt.Printf("%++v\n", err)
	}
	assert.True(t, err == nil)
	assert.True(t, len(supplyChainConfig.Nodes) > 0)

	supplyChain, err := NewSupplyChain(supplyChainConfig)
	if err != nil {
		fmt.Printf("%++v\n", err)
	}
	assert.True(t, err == nil)

	entityType := proto.EntityDTO_APPLICATION_COMPONENT
	nodes := supplyChain.nodeMap

	if _, exists := nodes[entityType]; !exists {
		t.Errorf("Missing %s node", entityType)
		assert.Fail(t, "Missing %s node", entityType)
	}

	appNode := nodes[entityType]

	expectedSoldComms := []proto.CommodityDTO_CommodityType{
		proto.CommodityDTO_TRANSACTION,
		proto.CommodityDTO_RESPONSE_TIME,
		proto.CommodityDTO_REMAINING_GC_CAPACITY,
		proto.CommodityDTO_THREADS,
		proto.CommodityDTO_HEAP,
		proto.CommodityDTO_KPI,
	}
	expectedSoldAccessComms := []proto.CommodityDTO_CommodityType{
		proto.CommodityDTO_APPLICATION,
	}

	var soldCommsList []proto.CommodityDTO_CommodityType
	var soldAccessCommsList []proto.CommodityDTO_CommodityType
	for key := range appNode.SupportedComms {
		soldCommsList = append(soldCommsList, key)
	}
	for key := range appNode.SupportedAccessComms {
		soldAccessCommsList = append(soldAccessCommsList, key)
	}

	assert.ElementsMatch(t, expectedSoldComms, soldCommsList)
	assert.ElementsMatch(t, expectedSoldAccessComms, soldAccessCommsList)

	expectedProviders := []proto.EntityDTO_EntityType{
		proto.EntityDTO_VIRTUAL_MACHINE,
	}
	expectedProviderComms := []proto.CommodityDTO_CommodityType{
		proto.CommodityDTO_VMEM,
		proto.CommodityDTO_VCPU,
	}
	var providers []proto.EntityDTO_EntityType
	var providerComms []proto.CommodityDTO_CommodityType
	for key, boughtMap := range appNode.SupportedBoughtComms {
		providers = append(providers, key)
		for comm := range boughtMap {
			providerComms = append(providerComms, comm)
		}
	}
	assert.EqualValues(t, expectedProviders, providers)
	assert.ElementsMatch(t, expectedProviderComms, providerComms)
}

var SERVICE_NODE string

func TestSupplyChainNodeMissingSoldCommodity(t *testing.T) {
	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   commoditySold:\n" +
			"   - key: key-placeholder \n"

	data := []byte(SERVICE_NODE)
	var sc *conf.SupplyChainConfig
	err := yaml.Unmarshal([]byte(data), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)
}

func TestSupplyChainNodeInvalidType(t *testing.T) {
	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE_ENTITY\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n"

	var sc *conf.SupplyChainConfig
	err := yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)
}

func TestSupplyChainNodeInvalidBought(t *testing.T) {
	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   commodityBought:\n" +
			"     - key:\n" +
			"         templateClass: APPLICATION_COMPONENT\n" +
			"         providerType: LAYERED_OVER\n"

	var sc *conf.SupplyChainConfig
	err := yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)

	SERVICE_NODE =
		"supplyChainNode:\n" +
			" - templateClass: SERVICE\n" +
			"   templateType: BASE\n" +
			"   templatePriority: -1\n" +
			"   commodityBought:\n" +
			"     - key:\n" +
			"         templateClass: APP_COMPONENT\n" +
			"         providerType: LAYERED_OVER\n"

	err = yaml.Unmarshal([]byte(SERVICE_NODE), &sc)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = NewSupplyChain(sc)
	fmt.Printf("ERR %v\n", err)
	assert.True(t, err != nil)
}

func TestSupplyChainNodeIgnoreIfPresent(t *testing.T) {
	supplyChainConfig, err := conf.LoadSupplyChain("test_app_node.yaml")
	assert.True(t, err == nil)
	assert.True(t, len(supplyChainConfig.Nodes) > 0)
	// Default supply chain
	supplyChain, err := NewSupplyChain(supplyChainConfig)
	assert.True(t, err == nil)
	templateDtoMap := supplyChain.CreateSupplyChainNodeTemplates()
	assert.NotEmpty(t, templateDtoMap)
	templateDTO, ok := templateDtoMap[proto.EntityDTO_APPLICATION_COMPONENT]
	assert.True(t, ok)
	ignoreIfPresent, err := getIgnoreIfPresent(templateDTO)
	assert.True(t, err == nil)
	assert.False(t, ignoreIfPresent)
	// Supply chain with ignoreIfPresent true
	supplyChain, err = NewSupplyChain(supplyChainConfig)
	assert.True(t, err == nil)
	templateDtoMap = supplyChain.IgnoreIfPresent(true).CreateSupplyChainNodeTemplates()
	assert.NotEmpty(t, templateDtoMap)
	templateDTO, ok = templateDtoMap[proto.EntityDTO_APPLICATION_COMPONENT]
	assert.True(t, ok)
	ignoreIfPresent, err = getIgnoreIfPresent(templateDTO)
	assert.True(t, err == nil)
	assert.True(t, ignoreIfPresent)
}

func getIgnoreIfPresent(templateDTO *proto.TemplateDTO) (bool, error) {
	mergedEntityMetadata := templateDTO.MergedEntityMetaData
	if mergedEntityMetadata == nil {
		return false, fmt.Errorf("missing mergedEntityMetadata")
	}
	commoditiesSoldMetadata := mergedEntityMetadata.CommoditiesSoldMetadata
	if commoditiesSoldMetadata == nil || len(commoditiesSoldMetadata) == 0 {
		return false, fmt.Errorf("missing commoditiesSoldMetadata")
	}
	ignoreIfPresent := commoditiesSoldMetadata[0].IgnoreIfPresent
	if ignoreIfPresent == nil {
		return false, fmt.Errorf("missing ignoreIfPresent")
	}
	return *ignoreIfPresent, nil
}
