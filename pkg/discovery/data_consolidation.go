package discovery

import (
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

type DIFConsolidationResult struct {
	Warnings       []string
	Scope          string
	ParsedEntities []*data.DIFEntity
}

// Pulls JSON DIF Data from different configured data sources
func GetDIFData(datasources []MetricDataSource) (*DIFConsolidationResult, error) {
	consolidationResult := &DIFConsolidationResult{}
	var endpoints []string
	// Specifying multiple endpoints in a single targetAddress is NOT
	// a common use case.
	// For consolidation of two or more entities coming from multiple data sources to happen,
	// the topology.Scope returned from those data sources must be the same, in addition
	// to the entity ID.
	// TODO: parallel execution for different endpoints
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
	return consolidationResult, nil
}
