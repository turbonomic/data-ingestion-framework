package discovery

import (
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"github.com/tamerh/jsparser"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

const (
	DIF_MATCH_IDENTIFIERS = "matchIdentifiers"
)

func parseMatchingIdentifiers(json *jsparser.JSON) *data.DIFMatchingIdentifiers {
	if json.ObjectVals[DIF_MATCH_IDENTIFIERS] == nil {
		return nil
	}

	matchIdMap := json.ObjectVals[DIF_MATCH_IDENTIFIERS]
	var ipAddress string
	var kubernetesFullyQualifiedName string
	hasValue := false
	for key, val := range matchIdMap.ObjectVals {
		if key == "ipAddress" {
			ipAddress = val.StringVal
			hasValue = true
		} else if key == "kubernetesFullyQualifiedName" {
			kubernetesFullyQualifiedName = val.StringVal
			hasValue = true
		}
	}

	var matchId *data.DIFMatchingIdentifiers
	if hasValue {
		matchId = &data.DIFMatchingIdentifiers{
			IPAddress:                    ipAddress,
			KubernetesFullyQualifiedName: kubernetesFullyQualifiedName,
		}
	}
	return matchId
}

func parsePartOf(json *jsparser.JSON) []*data.DIFPartOf {
	if json.ObjectVals["partOf"] == nil {
		return nil
	}
	partOfObj := json.ObjectVals["partOf"]
	var partOfList []*data.DIFPartOf
	for _, partOfEntry := range partOfObj.ArrayVals {
		difPartOf := &data.DIFPartOf{
			ParentEntity: "",
			UniqueId:     "",
		}
		for key, val := range partOfEntry.ObjectVals {
			if key == "entity" {
				difPartOf.ParentEntity = val.StringVal
			}
			if key == "uniqueId" {
				difPartOf.UniqueId = val.StringVal
			}
		}
		partOfList = append(partOfList, difPartOf)
	}

	return partOfList
}

func parseHostedOn(json *jsparser.JSON) *data.DIFHostedOn {
	if json.ObjectVals["hostedOn"] == nil {
		return nil
	}
	hostedOnObj := json.ObjectVals["hostedOn"]

	//var hostTypes []string
	var hostTypes []data.DIFHostType
	var hostIP, hostID string
	for key, hostedOnEntry := range hostedOnObj.ObjectVals { //is a map
		if key == "hostType" && hostedOnEntry.ArrayVals != nil {
			for _, hType := range hostedOnEntry.ArrayVals {
				hostTypes = append(hostTypes, data.DIFHostType(hType.StringVal))
			}
		}
		if key == "ipAddress" {
			hostIP = hostedOnEntry.StringVal
		}
		if key == "hostUuid" {
			hostID = hostedOnEntry.StringVal
		}
	}
	if len(hostTypes) > 0 {
		difHostedOn := &data.DIFHostedOn{
			HostType:  hostTypes,
			IPAddress: hostIP,
			HostUuid:  hostID,
		}
		return difHostedOn
	}

	return nil
}

func parseMetricVal(metricEntry *jsparser.JSON) []*data.DIFMetricVal {
	var mValList []*data.DIFMetricVal
	if metricEntry.ValueType != jsparser.Array {
		return []*data.DIFMetricVal{}
	}

	for _, mVal := range metricEntry.ArrayVals {
		// One DIFMetricVal per array element
		difMetricVal := &data.DIFMetricVal{}
		// Each array element is a map of key and values
		for k, v := range mVal.ObjectVals {
			if k == string(data.AVERAGE) {
				if v.ValueType == jsparser.Number {
					numVal, _ := strconv.ParseFloat(v.StringVal, 64)
					difMetricVal.Average = &numVal
				}
			}
			if k == string(data.MAX) {
				if v.ValueType == jsparser.Number {
					numVal, _ := strconv.ParseFloat(v.StringVal, 64)
					difMetricVal.Max = &numVal
				}
			}
			if k == string(data.MIN) {
				if v.ValueType == jsparser.Number {
					numVal, _ := strconv.ParseFloat(v.StringVal, 64)
					difMetricVal.Min = &numVal
				}
			}
			if k == string(data.CAPACITY) {
				glog.V(4).Infof("Found Capacity %+v", v.StringVal)
				if v.ValueType == jsparser.Number {
					numVal, _ := strconv.ParseFloat(v.StringVal, 64)
					difMetricVal.Capacity = &numVal
				}
			}
			if k == string(data.UNIT) {
				unitVal := parseUnitValue(v.StringVal)
				difMetricVal.Unit = &unitVal
			}
			if k == string(data.KEY) {
				if v.ValueType == jsparser.String {
					keyVal := v.StringVal
					difMetricVal.Key = &keyVal
				}
			}
		}
		//printMetricVal(difMetricVal)
		mValList = append(mValList, difMetricVal)
	} //end of innermost metric array

	return mValList
}

func parseUnitValue(unitVal interface{}) data.DIFMetricUnit {
	metricUnit := fmt.Sprintf("%v", unitVal)
	switch metricUnit {
	case "tps":
		return data.TPS
	case "mhz":
		return data.MHZ
	case "count":
		return data.COUNT
	case "ms":
		return data.MS
	case "mb":
		return data.MB
	case "pct":
		return data.PCT
	default:
		return ""
	}
}
