package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	dif "github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

const (
	path            = "/metrics"
	port            = 8081
	url             = "https://vmturbo.atlassian.net/rest/api/2/search?jql=project='Operations%20Manager'&maxResults=0"
	auth            = "fakeAuth"
	cookie          = "fakeCookie"
	appID           = "Turbonomic-turbonomic-http://prometheus-server:9090"
	stitchingID     = "BUSINESS_APPLICATION-Turbonomic-turbonomic-http://prometheus-server:9090"
	scope           = "Prometheus"
	defaultCapacity = 60000
)

type JiraResult struct {
	StartAt    int64
	MaxResults int64
	Total      int64
}

func getTickets() (float64, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to create http request GET %v: %v", url, err)
	}
	req.Header.Add("Authorization", auth)
	req.Header.Add("Cookie", cookie)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return 0, fmt.Errorf(
			"http request GET %v failed with error %v", url, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to read response of http request GET %v: %v", url, err)
	}
	var jiraResult JiraResult
	if err := json.Unmarshal(body, &jiraResult); err != nil {
		return 0, fmt.Errorf(
			"failed to unmarshal json [%v]: %v", string(body), err)
	}
	return float64(jiraResult.Total), nil
}

func sendTopology(w http.ResponseWriter, r *http.Request) {
	tickets, err := getTickets()
	if err != nil {
		glog.Errorf("Failed to get ticket: %v", err)
		sendFailure(w, r, err)
		return
	}
	// Create business app entity
	bizApp := dif.
		NewDIFEntity(appID, "businessApplication").
		Matching(stitchingID)
	key := "Total Tickets"
	capacity := float64(defaultCapacity)
	bizApp.AddMetrics("kpi", []*dif.DIFMetricVal{{
		Average:  &tickets,
		Capacity: &capacity,
		Key:      &key,
	}})
	// Create topology
	topology := dif.NewTopology().SetUpdateTime()
	topology.Scope = scope
	topology.AddEntity(bizApp)
	// Marshal to json
	result, err := json.Marshal(topology)
	if err != nil {
		glog.Errorf("Failed to marshal %v into json: %v.", topology, err)
		sendFailure(w, r, fmt.Errorf(
			"failed to marshal %v into json: %v", topology, err))
		return
	}
	glog.Infof("%v", tickets)
	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(result); err != nil {
		glog.Errorf("Failed to send response: %v.", err)
	}
	return
}

func sendFailure(w http.ResponseWriter, r *http.Request, e error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	status := fmt.Sprintf("{\"status\": \"%v\"}", e)
	if _, err := w.Write([]byte(status)); err != nil {
		glog.Errorf("Failed to send response: %v.", err)
	}
	return
}

func main() {
	http.HandleFunc(path, sendTopology)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		glog.Fatalf("Failed to create server: %v.", err)
	}
}
