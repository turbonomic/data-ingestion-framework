package builder

import (
	"fmt"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"math"
)

const (
	OneHundredPercent = 100.0
)

type CommodityDTOBuilder struct {
	commodityType           *proto.CommodityDTO_CommodityType
	key                     *string
	used                    *float64
	reservation             *float64
	capacity                *float64
	limit                   *float64
	peak                    *float64
	active                  *bool
	resizable               *bool
	displayName             *string
	thin                    *bool
	computedUsed            *bool
	usedIncrement           *float64
	propMap                 map[string][]string
	isUsedPct               *bool
	utilizationThresholdPct *float64
	pricingMetadata         *proto.CommodityDTO_PricingMetadata
	storageLatencyData      *proto.CommodityDTO_StorageLatencyData
	storageAccessData       *proto.CommodityDTO_StorageAccessData
	vstoragePartitionData   *proto.VStoragePartitionData
	utilizationData         *proto.CommodityDTO_UtilizationData
	vMemData                *proto.CommodityDTO_VMemData
	vCpuData                *proto.CommodityDTO_VCpuData

	err error
}

func NewCommodityDTOBuilder(commodityType proto.CommodityDTO_CommodityType) *CommodityDTOBuilder {
	return &CommodityDTOBuilder{
		commodityType: &commodityType,
	}
}

func (cb *CommodityDTOBuilder) Create() (*proto.CommodityDTO, error) {
	if cb.err != nil {
		return nil, cb.err
	}
	if err := cb.validateAndConvert(); err != nil {
		return nil, err
	}
	commodityDTO := &proto.CommodityDTO{
		CommodityType:   cb.commodityType,
		Key:             cb.key,
		Used:            cb.used,
		Reservation:     cb.reservation,
		Capacity:        cb.capacity,
		Limit:           cb.limit,
		Peak:            cb.peak,
		Active:          cb.active,
		Resizable:       cb.resizable,
		DisplayName:     cb.displayName,
		Thin:            cb.thin,
		ComputedUsed:    cb.computedUsed,
		UsedIncrement:   cb.usedIncrement,
		PropMap:         buildPropertyMap(cb.propMap),
		UtilizationData: cb.utilizationData,
	}

	if cb.storageLatencyData != nil {
		commodityDTO.CommodityData = &proto.CommodityDTO_StorageLatencyData_{cb.storageLatencyData}
	} else if cb.storageAccessData != nil {
		commodityDTO.CommodityData = &proto.CommodityDTO_StorageAccessData_{cb.storageAccessData}
	} else if cb.vstoragePartitionData != nil {
		commodityDTO.CommodityData = &proto.CommodityDTO_VstoragePartitionData{cb.vstoragePartitionData}
	}

	if cb.vCpuData != nil {
		commodityDTO.HotresizeData = &proto.CommodityDTO_VcpuData{cb.vCpuData}
	} else if cb.vMemData != nil {
		commodityDTO.HotresizeData = &proto.CommodityDTO_VmemData{cb.vMemData}
	}

	return commodityDTO, nil
}

func (cb *CommodityDTOBuilder) Key(key string) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.key = &key
	return cb
}

func (cb *CommodityDTOBuilder) Capacity(capacity float64) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.capacity = &capacity
	return cb
}

func (cb *CommodityDTOBuilder) Used(used float64) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.used = &used
	return cb
}

func (cb *CommodityDTOBuilder) Peak(peak float64) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.peak = &peak
	return cb
}

func (cb *CommodityDTOBuilder) Reservation(reservation float64) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.reservation = &reservation
	return cb
}

func (cb *CommodityDTOBuilder) Active(active bool) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.active = &active
	return cb
}

func (cb *CommodityDTOBuilder) Resizable(resizable bool) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.resizable = &resizable
	return cb
}

func (cb *CommodityDTOBuilder) HasResizable() bool {
	return cb.resizable != nil
}

func (cb *CommodityDTOBuilder) UtilizationData(points []float64, lastPointTimestampMs int64, intervalMs int32) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.utilizationData = &proto.CommodityDTO_UtilizationData{
		Point:                points,
		LastPointTimestampMs: &lastPointTimestampMs,
		IntervalMs:           &intervalMs,
	}
	return cb
}

func (cb *CommodityDTOBuilder) validateAndConvert() error {
	// Access commodities could have nil used value.
	if cb.used != nil && *cb.used < 0 {
		return fmt.Errorf("commodity %v has negative used value", cb.commodityType)
	}
	switch *cb.commodityType {
	case proto.CommodityDTO_THREADS, proto.CommodityDTO_CONNECTION:
		if cb.capacity == nil {
			return fmt.Errorf("commodity %v has nil capacity", cb.commodityType)
		}
	case proto.CommodityDTO_REMAINING_GC_CAPACITY:
		// The garbage collection time (gcTime) is in percentage, and is transformed to a commodity
		// with a capacity 100 [percent] and a used value that is 100 - gcTime.
		if cb.used == nil {
			return fmt.Errorf("commodity %v has nil used value", cb.commodityType)
		}
		used := math.Max(0.0, OneHundredPercent-*cb.used)
		cb.Used(used).Capacity(OneHundredPercent)
	case proto.CommodityDTO_DB_CACHE_HIT_RATE:
		cb.Capacity(OneHundredPercent)
	}
	return nil
}

func buildPropertyMap(propMap map[string][]string) (propList []*proto.CommodityDTO_PropertiesList) {
	for name, values := range propMap {
		propList = append(propList, &proto.CommodityDTO_PropertiesList{
			Name:   &name,
			Values: values,
		})
	}
	return
}
