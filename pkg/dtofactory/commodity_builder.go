package dtofactory

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/data-ingestion-framework/pkg/data"
	"github.com/turbonomic/data-ingestion-framework/pkg/registration"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	difdata "github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type GenericCommodityBuilder struct {
	entity *difdata.DIFEntity
}

func NewGenericCommodityBuilder(entity *difdata.DIFEntity) *GenericCommodityBuilder {
	return &GenericCommodityBuilder{entity: entity}
}

func (cb *GenericCommodityBuilder) BuildCommodity() (map[proto.CommodityDTO_CommodityType][]*builder.CommodityDTOBuilder, error) {
	result := make(map[proto.CommodityDTO_CommodityType][]*builder.CommodityDTOBuilder)

	metrics := cb.entity.Metrics
	for metricKey, metricList := range metrics { // Metrics is array of metric map [name,metric Value]
		// Check is metric is supported
		metricName := data.DIFMetricToTemplateCommodityStringMap[metricKey]
		commodityType, exists := registration.TemplateCommodityTypeMap[metricName]
		if !exists {
			glog.Errorf("%s:%s data has unsupported metric %s",
				cb.entity.Type, cb.entity.UID, metricName)
			continue
		}

		commodities, err := convertFromMetricValueListToCommodityList(commodityType, metricList)

		if err != nil {
			glog.Errorf("%v", err)
		}

		result[commodityType] = commodities
	}
	return result, nil
}

func convertFromMetricValueListToCommodityList(commType proto.CommodityDTO_CommodityType,
	responseMetrics []*difdata.DIFMetricVal) ([]*builder.CommodityDTOBuilder, error) { //([]*proto.CommodityDTO, error) {

	var commodityList []*builder.CommodityDTOBuilder //[]*proto.CommodityDTO

	for _, responseMetric := range responseMetrics {
		if responseMetric.Average == nil {
			return nil, fmt.Errorf("invalid commodity, missing average value for %v", commType)
		}
		commBuilder := builder.NewCommodityDTOBuilder(commType)
		if responseMetric.Capacity != nil {
			capacity := *responseMetric.Capacity
			commBuilder.Capacity(capacity) //TODO: do not set for bought commodity
		}
		if responseMetric.Average != nil {
			average := *responseMetric.Average
			commBuilder.Used(average)
		}
		if responseMetric.Key != nil {
			commKey := *responseMetric.Key
			commBuilder.Key(commKey)
		}

		commodityList = append(commodityList, commBuilder)
	}
	return commodityList, nil
}

func setResizable(entityType proto.EntityDTO_EntityType,
	commMap map[proto.CommodityDTO_CommodityType][]*builder.CommodityDTOBuilder) {
	// We must explicitly set resizable to false on an non-resizable commodity
	for _, nonResizableCommodity := range registration.NonResizableCommodities {
		if commList, found := commMap[nonResizableCommodity]; found {
			for _, comm := range commList {
				comm.Resizable(false)
			}
		}
	}
	switch entityType {
	case proto.EntityDTO_APPLICATION_COMPONENT:
		if commList, foundHeap := commMap[proto.CommodityDTO_HEAP]; foundHeap {
			resizable := false
			if _, foundGC := commMap[proto.CommodityDTO_REMAINING_GC_CAPACITY]; foundGC {
				resizable = true
			}
			for _, comm := range commList {
				comm.Resizable(resizable)
			}
		}
	case proto.EntityDTO_DATABASE_SERVER:
		if commList, foundHeap := commMap[proto.CommodityDTO_DB_MEM]; foundHeap {
			resizable := false
			if _, foundGC := commMap[proto.CommodityDTO_DB_CACHE_HIT_RATE]; foundGC {
				resizable = true
			}
			for _, comm := range commList {
				comm.Resizable(resizable)
			}
		}
	}
}
