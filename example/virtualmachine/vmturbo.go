package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	dif "github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

const (
	path        = "/metrics"
	port        = 8081
	stitchingID = "172.23.0.5"
	used        = 1.3 * 1024 * 1024
	capacity    = 3.5 * 1024 * 1024
	vmID        = "spcfq9keqj-worker-1"
)

func sendTopology(w http.ResponseWriter, r *http.Request) {
	// Create the VM entity
	vm := dif.NewDIFEntity(vmID, "virtualMachine").Matching(stitchingID)
	// Add the memory metrics
	used := used
	capacity := capacity
	vm.AddMetrics("memory", []*dif.DIFMetricVal{{
		Average:  &used,
		Capacity: &capacity,
	}})
	// Create topology
	topology := dif.NewTopology().SetUpdateTime()
	// Add the VM entity to the topology
	topology.AddEntity(vm)
	// Send the topology
	sendResult(topology, w, r)
}

func sendResult(topology *dif.Topology, w http.ResponseWriter, r *http.Request) {
	var status = http.StatusOK
	var result []byte
	var err error
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(result)
	}()
	// Marshal to json
	if result, err = json.Marshal(topology); err != nil {
		status = http.StatusInternalServerError
		result = []byte(fmt.Sprintf("{\"status\": \"%v\"}", err))
	}
}

func main() {
	http.HandleFunc(path, sendTopology)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		glog.Fatalf("Failed to create server: %v.", err)
	}
}
