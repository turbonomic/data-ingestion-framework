package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/golang/glog"
	dif "github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
)

const (
	metricPath         = "/metrics"
	port               = 8081
	baseLoginURL       = "https://login.microsoftonline.com/"
	baseQueryURL       = "https://api.loganalytics.io/"
	defaultResource    = "https://api.loganalytics.io"
	defaultRedirectURI = "http://localhost:3000/login"
	defaultGrantType   = "client_credentials"
	queryComputer      = `
Heartbeat | summarize arg_max(TimeGenerated, *) by Computer | project Computer, ComputerIP`
	queryMemory = `
Perf
| where TimeGenerated > ago(10m)
| where ObjectName == "Memory" and
(CounterName == "Used Memory MBytes" or // the name used in Linux records
CounterName == "Committed Bytes") // the name used in Windows records
| summarize avg(CounterValue) by Computer, CounterName, bin(TimeGenerated, 10m)
| order by TimeGenerated
`
)

var (
	workspaces   []string
	tenantID     string
	clientID     string
	clientSecret string
	hostMap      map[string]string
	client       *http.Client
)

type Column struct {
	Name string
	Type string
}

type Table struct {
	Columns []*Column
	Name    string
	Rows    [][]interface{}
}

type LogAnalyticsQueryRequest struct {
	Query string `json:"query"`
}

type LogAnalyticsQueryResults struct {
	Tables []*Table
}

type ErrorDetail struct {
	Code      string
	Message   string
	Resources []string
	Target    string
	Value     string
}

type ErrorInfo struct {
	Details []*ErrorDetail
	Code    string
	Message string
}

type LogAnalyticsErrorResponse struct {
	Errors *ErrorInfo
}

type AccessToken struct {
	Type        string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	ExpiresOn   string `json:"expires_on"`
	Resource    string `json:"resource"`
	AccessToken string `json:"access_token"`
}

func init() {
	_ = flag.Set("alsologtostderr", "true")
	_ = flag.Set("stderrthreshold", "INFO")
	_ = flag.Set("v", "2")
	flag.Parse()
	workspaceIDs := os.Getenv("AZURE_LOG_ANALYTICS_WORKSPACES")
	if workspaceIDs == "" {
		glog.Fatalf("AZURE_LOG_ANALYTICS_WORKSPACES is missing.")
	}
	workspaces = strings.Split(workspaceIDs, ",")
	tenantID = os.Getenv("AZURE_TENANT_ID")
	if tenantID == "" {
		glog.Fatalf("AZURE_TENANT_ID is missing.")
	}
	clientID = os.Getenv("AZURE_CLIENT_ID")
	if clientID == "" {
		glog.Fatalf("AZURE_CLIENT_ID is missing.")
	}
	clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	if clientSecret == "" {
		glog.Fatalf("AZURE_CLIENT_SECRET is missing")
	}
	client = &http.Client{}
	hostMap = make(map[string]string)
}

func login() (string, error) {
	loginURL, _ := url.Parse(baseLoginURL)
	loginURL.Path = path.Join(tenantID, "oauth2", "token")
	data := url.Values{}
	data.Set("grant_type", defaultGrantType)
	data.Set("resource", defaultResource)
	data.Set("redirect_uri", defaultRedirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	req, err := http.NewRequest(http.MethodPost, loginURL.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf(
			"failed to create request POST %v: %v", loginURL, err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return "", fmt.Errorf(
			"request GET %v failed with error %v", loginURL, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf(
			"failed to read response of request POST %v: %v", loginURL, err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request POST %v failed with status %v", loginURL, res.StatusCode)
	}
	var accessToken AccessToken
	if err := json.Unmarshal(body, &accessToken); err != nil {
		return "", fmt.Errorf(
			"failed to unmarshal json [%v]: %v", string(body), err)
	}
	return accessToken.AccessToken, nil

}

func createAndSendTopology(w http.ResponseWriter, r *http.Request) {
	// Always obtain a new token first
	token, err := login()
	if err != nil {
		glog.Errorf("Failed to login: %v", err)
	}
	glog.V(2).Infof("Token: %s", token)
	topology, err := createTopology(token)
	if err != nil {
		glog.Errorf("Failed to create topology: %v", err)
	}
	sendResult(topology, w, r)
}

func createTopology(token string) (*dif.Topology, error) {
	// Create topology
	topology := dif.NewTopology().SetUpdateTime()
	// Iterate through all workspaces
	for _, workspace := range workspaces {
		queryResults, err := doQuery(token, queryComputer, workspace)
		if err != nil {
			glog.Errorf("failed to query computer: %v", err)
			continue
		}
		for _, table := range queryResults.Tables {
			for _, row := range table.Rows {
				// row[0]: Computer
				// row[1]: ComputerIP
				computer := row[0].(string)
				computerIP := row[1].(string)
				hostMap[computer] = computerIP
			}
		}
		glog.V(2).Infof(spew.Sdump(hostMap))
		queryResults, err = doQuery(token, queryMemory, workspace)
		if err != nil {
			glog.Errorf("failed to query memory: %v", err)
			continue
		}
		hostSeen := make(map[string]bool)
		for _, table := range queryResults.Tables {
			for _, row := range table.Rows {
				// row[0]: Computer
				// row[1]: CounterName
				// row[2]: TimeGenerated
				// row[3]: CounterValue
				computer := row[0].(string)
				computerIP, found := hostMap[computer]
				if !found {
					glog.Warningf("Cannot find IP address for computer %s.", computer)
					continue
				}
				seen, _ := hostSeen[computer]
				if seen {
					continue
				}
				hostSeen[computer] = true
				counterName := row[1].(string)
				var avgMemUsedKB float64
				if counterName == "Used Memory MBytes" {
					avgMemUsedKB = row[3].(float64) * 1024
				} else if counterName == "Committed Bytes" {
					avgMemUsedKB = row[3].(float64) / 1024
				} else {
					glog.Warningf("Unrecognized CounterName %s.", counterName)
				}
				// Create the VM entity
				vm := dif.NewDIFEntity(computer, "virtualMachine").Matching(computerIP)
				// Add the memory metrics
				vm.AddMetrics("memory", []*dif.DIFMetricVal{{
					Average: &avgMemUsedKB,
				}})
				topology.AddEntity(vm)
			}
		}
	}
	return topology, nil
}

func doQuery(token, query, workspace string) (*LogAnalyticsQueryResults, error) {
	queryURL, _ := url.Parse(baseQueryURL)
	queryURL.Path = path.Join("v1", "workspaces", workspace, "query")
	data, _ := json.Marshal(&LogAnalyticsQueryRequest{Query: query})
	glog.V(2).Infof("Marshalled JSON %s", string(data))
	req, err := http.NewRequest(http.MethodPost, queryURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create request POST %v: %v", queryURL, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf(
			"request POST %v failed with error %v", queryURL, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read response of request POST %v: %v", queryURL, err)
	}
	if res.StatusCode != http.StatusOK {
		var errorResponse LogAnalyticsErrorResponse
		_ = json.Unmarshal(body, &errorResponse)
		return nil, fmt.Errorf("request POST %v failed with status %v and error %v",
			queryURL, res.StatusCode, errorResponse)
	}
	glog.V(2).Infof("Response body: %s", string(body))
	var queryResults LogAnalyticsQueryResults
	if err := json.Unmarshal(body, &queryResults); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal json [%v]: %v", string(body), err)
	}
	return &queryResults, nil
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
	glog.V(2).Infof("Sending result: %v.", string(result))
}

func main() {
	http.HandleFunc(metricPath, createAndSendTopology)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		glog.Fatalf("Failed to create server: %v.", err)
	}
}
