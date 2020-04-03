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
	for _, metricMap := range metrics { //Metrics is array of metric map [name,metric Value]

		if len(metricMap) > 1 {
			glog.Errorf("Invalid metric data %++v", metricMap)
			continue
		}
		for metricKey, metricList := range metricMap {
			// Check is metric is supported
			metricName := data.DIFMetricToTemplateCommodityStringMap[metricKey]
			commodityType, exists := registration.TemplateCommodityTypeMap[metricName]
			if !exists {
				glog.Errorf("%s:%s data has unsupported metric %s\n",
					cb.entity.Type, cb.entity.UID, metricName)
				continue
			}

			commodities, err := cb.convertFromMetricValueListToCommodityList(commodityType, metricList)

			if err != nil {
				glog.Errorf("%v", err)
			}

			result[commodityType] = commodities
		}
	}

	return result, nil
}

func (cb *GenericCommodityBuilder) convertFromMetricValueListToCommodityList(commType proto.CommodityDTO_CommodityType,
	responseMetrics []*difdata.DIFMetricVal) ([]*builder.CommodityDTOBuilder, error) { //([]*proto.CommodityDTO, error) {

	var commodityList []*builder.CommodityDTOBuilder //[]*proto.CommodityDTO

	for _, responseMetric := range responseMetrics {
		if responseMetric.Average == nil {
			return nil, fmt.Errorf("Invalid commodity, missing average value for %v\n", commType)
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
		//commodity, err := commBuilder.Create()
		//if err != nil {
		//	glog.Errorf("%v", err)
		//	return nil, err
		//}
		//
		//commodityList = append(commodityList, commodity)
	}
	return commodityList, nil
}
