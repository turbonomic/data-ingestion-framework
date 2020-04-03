package discovery

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"io/ioutil"
)

type DIFConsolidationResult struct {
	Warnings       []string
	Scope          string
	ParsedEntities []*data.DIFEntity
}

// Pulls JSON DIF Data from different configured data sources
func GetDIFData(datasources []MetricDataSource) (*DIFConsolidationResult, error) {
	consolidationResult := &DIFConsolidationResult{
		Warnings:       nil,
		ParsedEntities: nil,
		Scope:          "",
	}
	var endpoints []string
	//TODO: parallel execution for different endpoints
	for _, datasource := range datasources {
		endpoints = append(endpoints, datasource.GetMetricEndpoint())
		var topology *data.Topology
		topology, err := datasource.GetMetricData()
		if err != nil {
			return nil, err
		}
		if topology.Entities != nil {
			consolidationResult.ParsedEntities = append(consolidationResult.ParsedEntities, topology.Entities...)
		}
		consolidationResult.Scope = topology.Scope
	}

	//3. marshal all the entities to json format
	if glog.V(4) {
		topo := data.NewTopology()
		topo.SetUpdateTime()
		topo.Entities = consolidationResult.ParsedEntities
		topo.Scope = consolidationResult.Scope

		jdata, err := json.MarshalIndent(topo, "", " ")
		if err != nil {
			glog.Errorf("Failed to marshal json: %v.", err)
		}

		ioutil.WriteFile("./dif-data.json", jdata, 0644)
	}

	return consolidationResult, nil
}
