package dtofactory

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.ibm.com/turbonomic/data-ingestion-framework/pkg/registration"
	difdata "github.ibm.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
)

var TEST_ENTITY = "{" +
	"	\"type\": \"application\"," +
	"	\"uniqueId\": \"456\"," +
	"	\"name\": \"My App Name\"," +
	"	\"hostedOn\": {" +
	"		\"hostType\": [" +
	"			\"virtualMachine\"" +
	"		]," +
	"		\"ipAddress\": \"10.10.168.193\"," +
	"		\"hostUuid\": \"\"" +
	"	}," +
	"	\"metrics\" : {" +
	"		\"connections\": [" +
	"			{" +
	"				\"average\": 33.2," +
	"				\"unit\": \"tps\"" +
	"			}" +
	"		]," +
	"		\"cpu\": [" +
	"			{" +
	"				\"average\": 33.2," +
	"				\"unit\": \"mhz\"," +
	"				\"rawData\": {" +
	"					\"utilization\": [" +
	"						{" +
	"							\"1579607218\": \"60.5\"" +
	"						}," +
	"						{" +
	"							\"1579607218\": \"60.5\"" +
	"						}" +
	"					]" +
	"				}," +
	"				\"units\": \"mhz\"" +
	"			}" +
	"		]," +
	"		\"kpi\": [" +
	"			{" +
	"				\"unit\": \"tps\"," +
	"				\"average\": 33.2," +
	"				\"key\": \"xxx\"" +
	"			}," +
	"			{" +
	"				\"unit\": \"tps\"," +
	"				\"average\": 33.2," +
	"				\"key\": \"xxx1\"" +
	"			}" +
	"		]" +
	"	}" +
	"}"

func TestCommodity(t *testing.T) {

	difEntity := parseEntity(TEST_ENTITY)
	cb := NewGenericCommodityBuilder(difEntity)
	metricMap := cb.entity.Metrics
	for metricKey, metricList := range metricMap { //Metrics is array of metric map [name,metric Value]
		//description := metricEntry.Description
		//if description != nil {
		//	fmt.Printf("DESCRIPTION %s\n", *description )
		//}
		//metricMap := metricEntry.MetricMap
		// each metric is a map of metric name and its value
		//if len(metricMap) > 1 {
		//	continue
		//}
		//for metricKey, metricList := range metricMap {
		// Parse metric
		metricName := data.DIFMetricToTemplateCommodityStringMap[metricKey]
		commodityType := registration.TemplateCommodityTypeMap[metricName]

		commodities, err := convertFromMetricValueListToCommodityList(commodityType, metricList)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		fmt.Printf("comm %v\n", commodities)

		//}
	}
}

var ENTITY string
var INVALID_COMM_NAME = "    	  \"INVALID_COMM_NAME\": [" +
	"			{" +
	"		          \"average\": 33.2," +
	"		          \"unit\": \"tps\"" +
	"			}" +
	"	   ]"

var INVALID_METRIC = "    	  \"responseTime\": [" +
	"			{" +
	"		          \"unit\": \"tps\"," +
	"				  \"capacity\": 100" +
	"	        }" +
	"		]"

var CONNECTION = "    	  \"connection\": [" +
	"			{" +
	"		          \"average\": 33.2," +
	"		          \"unit\": \"tps\"" +
	"			}"

var RESPONSE_TIME = "    	  \"responseTime\": [" +
	"			{" +
	"		          \"average\": 33.2," +
	"		          \"unit\": \"tps\"" +
	"			}"

var TRANSACTION = "    	  \"transaction\": [" +
	"			{" +
	"		          \"average\": 33.2," +
	"		          \"unit\": \"tps\"" +
	"			}"

var HEAP_ARRAY = "\"heap\":[" +
	"	{" +
	"		\"average\":449195.6953125," +
	"		\"unit\":\"\"," +
	"		\"key\":\"\"" +
	"	}," +
	"	{" +
	"		\"average\":75434.1234565," +
	"		\"unit\":\"\"," +
	"		\"key\":\"\"" +
	"	}"

var DBMEM_RESIZE = "\"dbMem\":[" +
	"	{" +
	"		\"average\":100," +
	"		\"resizable\":false" +
	"	}"

var DBMEM = "\"dbMem\":[" +
	"	{" +
	"		\"average\":100" +
	"	}"

var DBCACHEHITRATE = "\"dbCacheHitRate\":[" +
	"	{" +
	"		\"average\":100" +
	"	}"

var KPI_ARRAY = "\"kpi\":[" +
	"	{" +
	"		\"average\":100," +
	"		\"unit\":\"\"," +
	"		\"key\":\"KPI1\"" +
	"	}," +
	"	{" +
	"		\"average\":200," +
	"		\"unit\":\"\"," +
	"		\"key\":\"KPI2\"" +
	"	}" +
	"]"

func parseEntity(entityString string) *difdata.DIFEntity {

	var difEntity *difdata.DIFEntity
	err := json.Unmarshal([]byte(entityString), &difEntity)

	if err != nil {
		log.Fatalf("PARSE error: %v", err)
	}

	return difEntity
}

func TestCommodityArray(t *testing.T) {
	ENTITY =
		"{" +
			"	\"type\": \"application\"," +
			"	\"uniqueId\": \"456\"," +
			"	\"name\": \"My App Name\"," +
			"	\"metrics\" : {" +
			CONNECTION +
			" 		]," +
			TRANSACTION +
			"       ]," +
			HEAP_ARRAY +
			"		]" +
			"     }" +
			"}"

	difEntity := parseEntity(ENTITY)
	cb := NewGenericCommodityBuilder(difEntity)
	commMap, err := cb.BuildCommodity()
	if err != nil {
		log.Fatalf(" ERROR: %v", err)
	}

	expectedCommMap := map[proto.CommodityDTO_CommodityType]int{
		proto.CommodityDTO_HEAP:        2,
		proto.CommodityDTO_TRANSACTION: 1,
		proto.CommodityDTO_CONNECTION:  1,
	}

	for commType, num := range expectedCommMap {
		if _, exists := commMap[commType]; !exists {
			assert.Fail(t, fmt.Sprintf("Commodity %v was not created", commType))
		}
		if num != len(commMap[commType]) {
			assert.Fail(t,
				fmt.Sprintf("Commodity %v : num of commodities [%d] created is not equal to expected value [%d]",
					commType, len(commMap[commType]), num))
		}
	}
}

func TestCommodityKPI(t *testing.T) {
	ENTITY =
		"{" +
			"	\"type\": \"application\"," +
			"	\"uniqueId\": \"456\"," +
			"	\"name\": \"My App Name\"," +
			"	\"metrics\" : {" +
			KPI_ARRAY +
			" }" +
			"}"

	difEntity := parseEntity(ENTITY)
	cb := NewGenericCommodityBuilder(difEntity)
	commMap, err := cb.BuildCommodity()
	if err != nil {
		log.Fatalf(" ERROR: %v", err)
	}

	for key, commList := range commMap {
		for _, cb := range commList {
			comm, _ := cb.Create()
			fmt.Printf("key %s --> %++v\n", key, comm)
		}
	}

	expectedCommMap := map[proto.CommodityDTO_CommodityType]int{
		proto.CommodityDTO_KPI: 2,
	}

	for commType, num := range expectedCommMap {
		if _, exists := commMap[commType]; !exists {
			assert.Fail(t, fmt.Sprintf("Commodity %v was not created", commType))
		}
		if num != len(commMap[commType]) {
			assert.Fail(t,
				fmt.Sprintf("Commodity %v : num of commodities [%d] created is not equal to expected value [%d]",
					commType, len(commMap[commType]), num))
		}
	}
}

func TestCommodityInvalid(t *testing.T) {
	ENTITY =
		"{" +
			"	\"type\": \"application\"," +
			"	\"uniqueId\": \"456\"," +
			"	\"name\": \"My App Name\"," +
			"	\"metrics\" : {" +
			INVALID_COMM_NAME +
			"     }" +
			"}"

	difEntity := parseEntity(ENTITY)

	cb := NewGenericCommodityBuilder(difEntity)
	commMap, err := cb.BuildCommodity()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	assert.True(t, len(commMap) == 0, "Pass: Commmodity with invalid name is not created")
}

func TestCommodityMissingMetrics(t *testing.T) {
	ENTITY =
		"{" +
			"	\"type\": \"application\"," +
			"	\"uniqueId\": \"456\"," +
			"	\"name\": \"My App Name\"," +
			"	\"metrics\" : {" +
			INVALID_METRIC +
			"     }" +
			"}"

	difEntity := parseEntity(ENTITY)

	cb := NewGenericCommodityBuilder(difEntity)
	commMap, err := cb.BuildCommodity()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for _, commList := range commMap {
		assert.True(t, len(commList) == 0, "Pass: Commmodity with missing used value is not created")
	}
}

func TestCommodityResizableDefault(t *testing.T) {
	ENTITY =
		"{" +
			"	\"type\": \"databaseServer\"," +
			"	\"uniqueId\": \"456\"," +
			"	\"name\": \"My App Name\"," +
			"	\"metrics\" : {" +
			DBMEM +
			"       ]," +
			DBCACHEHITRATE +
			"		]" +
			"     }" +
			"}"
	entityType := proto.EntityDTO_DATABASE_SERVER
	commType := proto.CommodityDTO_DB_MEM
	difEntity := parseEntity(ENTITY)
	cb := NewGenericCommodityBuilder(difEntity)
	commMap, err := cb.BuildCommodity()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	setResizable(entityType, commMap)
	for _, commList := range commMap {
		for _, cb := range commList {
			comm, _ := cb.Create()
			if comm.GetCommodityType() != commType {
				continue
			}
			assert.True(t, comm.GetResizable())
		}
	}
}

func TestCommodityResizableJSONOverrideDefault(t *testing.T) {
	ENTITY =
		"{" +
			"	\"type\": \"databaseServer\"," +
			"	\"uniqueId\": \"456\"," +
			"	\"name\": \"My App Name\"," +
			"	\"metrics\" : {" +
			DBMEM_RESIZE +
			"       ]," +
			DBCACHEHITRATE +
			"		]" +
			"     }" +
			"}"
	entityType := proto.EntityDTO_DATABASE_SERVER
	commType := proto.CommodityDTO_DB_MEM
	difEntity := parseEntity(ENTITY)
	cb := NewGenericCommodityBuilder(difEntity)
	commMap, err := cb.BuildCommodity()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	setResizable(entityType, commMap)
	for _, commList := range commMap {
		for _, cb := range commList {
			comm, _ := cb.Create()
			if comm.GetCommodityType() != commType {
				continue
			}
			assert.False(t, comm.GetResizable())
		}
	}
}
