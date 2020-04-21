package discovery

import (
	"bufio"
	//"bytes"
	"fmt"
	"github.com/tamerh/jsparser"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"os"
	"strconv"
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
	for key, val := range matchIdMap.ObjectVals {
		if key == "ipAddress" {
			ipAddress = val.StringVal
		}
	}

	var matchId *data.DIFMatchingIdentifiers
	if len(ipAddress) > 0 {
		matchId = &data.DIFMatchingIdentifiers{
			IPAddress: ipAddress,
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

	//fmt.Printf("Parsing %++v\n", metricEntry)
	if metricEntry.ValueType != jsparser.Array {
		return []*data.DIFMetricVal{}
	}

	for _, mVal := range metricEntry.ArrayVals {
		// One DIFMetricVal per array element
		difMetricVal := &data.DIFMetricVal{
			Average:     nil,
			Min:         nil,
			Max:         nil,
			Capacity:    nil,
			Unit:        nil,
			Key:         nil,
			Description: nil,
			RawMetrics:  nil,
		}
		// Each array element is a map of key and values
		for k, v := range mVal.ObjectVals {
			if k == "average" {
				if v.ValueType == jsparser.Number {
					numVal, _ := strconv.ParseFloat(v.StringVal, 64)
					difMetricVal.Average = &numVal
				}
			}
			if k == "max" {
				if v.ValueType == jsparser.Number {
					numVal, _ := strconv.ParseFloat(v.StringVal, 64)
					difMetricVal.Max = &numVal
				}
			}
			if k == "min" {
				if v.ValueType == jsparser.Number {
					numVal, _ := strconv.ParseFloat(v.StringVal, 64)
					difMetricVal.Min = &numVal
				}
			}
			if k == "capacity" {
				fmt.Printf("*** FOUND CAPACITY %++v\n", v.StringVal)
				if v.ValueType == jsparser.Number {
					numVal, _ := strconv.ParseFloat(v.StringVal, 64)
					difMetricVal.Capacity = &numVal
				}
			}
			if k == "unit" {
				unitVal := parseUnitValue(v.StringVal)
				difMetricVal.Unit = &unitVal
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

//==================================================================================

func printMetricVal(m *data.DIFMetricVal) {
	if m.Average != nil {
		fmt.Printf("	Average: %v ", *m.Average)
	}
	if m.Capacity != nil {
		fmt.Printf("Capacity: %v ", *m.Capacity)
	}
	if m.Unit != nil {
		fmt.Printf("Unit: %v ", *m.Unit)
	}
	if m.Key != nil {
		fmt.Printf("Key: %v ", *m.Key)
	}
	fmt.Printf("\n")
}

func DIFEntityToString(entity *data.DIFEntity) {
	fmt.Printf("*** [%s]%s:%s\n", entity.Type, entity.UID, entity.Name)

	if entity.MatchingIdentifiers != nil {
		fmt.Printf("	MatchingIdentifiers:\n")
		fmt.Printf("		IPAddress : %++v\n", entity.MatchingIdentifiers.IPAddress)
	}
	if entity.PartOf != nil {
		fmt.Printf("	PartOf: %d\n", len(entity.PartOf))
		for _, partOf := range entity.PartOf {
			fmt.Printf("		%s:%s\n", partOf.ParentEntity, partOf.UniqueId)
		}
	}

	if entity.HostedOn != nil {
		fmt.Printf("		%s:%s\n", entity.HostedOn.HostUuid, entity.HostedOn.IPAddress)
	}

	if entity.Metrics != nil {
		for metricName, metricList := range entity.Metrics {
			for _, metric := range metricList {
				fmt.Printf("	Metric %s:\n", metricName)
				printMetricVal(metric)
			}
		}
	}
}

//==================================================================================
// For unit testing only
func ReadDIFTopologyStream(path string) {
	//var resp []byte
	//rb := bytes.NewBuffer(resp)
	//reader := bufio.NewReader(rb)
	//br := bufio.NewReaderSize(reader, 65536)
	//parser := jsparser.NewJSONParser(br, "topology")

	fmt.Printf("========================================\n")
	f, _ := os.Open(path)
	br := bufio.NewReaderSize(f, 65536)
	parser1 := jsparser.NewJSONParser(br, "scope")
	for json := range parser1.Stream() {
		fmt.Printf("Scope %++v\n", json.StringVal)
	}

	f, _ = os.Open(path)
	br = bufio.NewReaderSize(f, 65536)
	parser := jsparser.NewJSONParser(br, "topology")
	var entities []*data.DIFEntity
	for json := range parser.Stream() {
		// Create DIFEntity
		entity := &data.DIFEntity{
			UID:                 json.ObjectVals["uniqueId"].StringVal,
			Type:                json.ObjectVals["type"].StringVal,
			Name:                json.ObjectVals["name"].StringVal,
			HostedOn:            nil,
			MatchingIdentifiers: nil,
			PartOf:              nil,
			Metrics:             nil,
		}
		fmt.Printf("%s::%s\n", entity.UID, entity.Type)
		// ----- Matching Identifiers
		entity.MatchingIdentifiers = parseMatchingIdentifiers(json)

		// ----- Part Of
		entity.PartOf = parsePartOf(json)

		// ----- HostedOn
		entity.HostedOn = parseHostedOn(json)

		// ------ Metrics
		if json.ObjectVals["metrics"] != nil {
			difMetricValMap := make(map[string][]*data.DIFMetricVal)

			metricsMap := json.ObjectVals["metrics"].ObjectVals
			for metricName, metrics := range metricsMap {
				fmt.Printf("%s --> metrics: %++v\n", metricName, metrics)
				metricList := parseMetricVal(metrics)
				difMetricValMap[metricName] = metricList
			}
			entity.Metrics = difMetricValMap
		}

		DIFEntityToString(entity)
		entities = append(entities, entity)
		fmt.Printf("------------------------------\n")
	} //end of topology

	var topology *data.Topology
	topology = &data.Topology{
		Version:    "",
		Updatetime: 0,
		Scope:      "",
		Source:     "",
		Entities:   nil,
	}
	topology.Entities = entities
	fmt.Printf("========================================\n")
}
