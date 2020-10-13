package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
	dif "github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"gopkg.in/yaml.v2"
)

const (
	path                       = "/metrics"
	port                       = 8081
	testsURL                   = "https://api.thousandeyes.com/v6/tests"
	netMetricsURL              = "https://api.thousandeyes.com/v6/net/metrics"
	scope                      = "Prometheus"
	defaultCapacity            = 60000
	defaultCredentialsLocation = "/etc/credentials"
)

var (
	username string
	token    string
	client   *http.Client
)

type ThousandEyes struct {
	Test []Test `json:"test"`
}

type Test struct {
	CreatedDate           string    `json:"createdDate,omitempty"`
	CreatedBy             string    `json:"createdBy,omitempty"`
	Enabled               int       `json:"enabled"`
	SavedEvent            int       `json:"savedEvent"`
	TestID                int       `json:"testId"`
	TestName              string    `json:"testName"`
	Type                  string    `json:"type"`
	Interval              int       `json:"interval"`
	URL                   string    `json:"url"`
	Protocol              string    `json:"protocol"`
	Ipv6Policy            string    `json:"ipv6Policy"`
	NetworkMeasurements   int       `json:"networkMeasurements"`
	MtuMeasurements       int       `json:"mtuMeasurements"`
	BandwidthMeasurements int       `json:"bandwidthMeasurements"`
	BgpMeasurements       int       `json:"bgpMeasurements"`
	UsePublicBgp          int       `json:"usePublicBgp"`
	AlertsEnabled         int       `json:"alertsEnabled"`
	LiveShare             int       `json:"liveShare"`
	HTTPTimeLimit         int       `json:"httpTimeLimit"`
	HTTPTargetTime        int       `json:"httpTargetTime"`
	HTTPVersion           int       `json:"httpVersion"`
	FollowRedirects       int       `json:"followRedirects"`
	SslVersionID          int       `json:"sslVersionId"`
	VerifyCertificate     int       `json:"verifyCertificate"`
	UseNtlm               int       `json:"useNtlm"`
	AuthType              string    `json:"authType"`
	ContentRegex          string    `json:"contentRegex"`
	ProbeMode             string    `json:"probeMode"`
	PathTraceMode         string    `json:"pathTraceMode"`
	NumPathTraces         int       `json:"numPathTraces"`
	APILinks              []APILink `json:"apiLinks"`
	SslVersion            string    `json:"sslVersion"`
}

type APILink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type NetMetrics struct {
	Net   Net   `json:"net"`
	Pages Pages `json:"pages"`
}

type Net struct {
	Test    Test      `json:"test"`
	Metrics []Metrics `json:"metrics"`
}

type Metrics struct {
	AvgLatency float64 `json:"avgLatency"`
	Loss       float64 `json:"loss"`
	MaxLatency float64 `json:"maxLatency"`
	Jitter     float64 `json:"jitter"`
	MinLatency float64 `json:"minLatency"`
	ServerIP   string  `json:"serverIp"`
	AgentName  string  `json:"agentName"`
	CountryID  string  `json:"countryId"`
	Date       string  `json:"date"`
	AgentID    int     `json:"agentId"`
	RoundID    int     `json:"roundId"`
	Permalink  string  `json:"permalink"`
}

type Pages struct {
	Current int `json:"current"`
}

type Credentials struct {
	UserName string `yaml:"username"`
	Token    string `yaml:"token"`
}

func init() {
	_ = flag.Set("alsologtostderr", "true")
	_ = flag.Set("stderrthreshold", "INFO")
	_ = flag.Set("v", "2")
	flag.Parse()
	credentialsLocation := os.Getenv("TARGET_INFO_LOCATION")
	if credentialsLocation == "" {
		credentialsLocation = defaultCredentialsLocation
	}
	credentialsFile, err := ioutil.ReadFile(credentialsLocation)
	if err != nil {
		glog.Fatalf("Failed to read target info from file %v: %v", credentialsLocation, err)
	}
	var credentials Credentials
	err = yaml.Unmarshal(credentialsFile, &credentials)
	if err != nil {
		glog.Fatalf("Failed to unmarshal target info from file %v: %v", credentialsLocation, err)
	}
	username = credentials.UserName
	if username == "" {
		glog.Fatalf("username is missing.")
	}
	token = credentials.Token
	if token == "" {
		glog.Fatalf("token is missing.")
	}
	client = &http.Client{}
}

func getTests() ([]Test, error) {
	req, err := http.NewRequest("GET", testsURL+".json", nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create http request GET %v: %v", testsURL, err)
	}
	req.Header.Set("Authorization", "Basic "+basicAuth(username, token))
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf(
			"http request GET %v failed with error %v", testsURL, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read response of http request GET %v: %v", testsURL, err)
	}
	var tests ThousandEyes
	if err := json.Unmarshal(body, &tests); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal json [%v]: %v", string(body), err)
	}
	return tests.Test, nil
}

func getMetrics(metricsUrl string) ([]Metrics, error) {
	req, err := http.NewRequest("GET", metricsUrl+".json", nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create http request GET %v: %v", testsURL, err)
	}
	req.Header.Set("Authorization", "Basic "+basicAuth(username, token))
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf(
			"http request GET %v failed with error %v", testsURL, err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read response of http request GET %v: %v", testsURL, err)
	}
	var netMetrics NetMetrics
	if err := json.Unmarshal(body, &netMetrics); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal json [%v]: %v", string(body), err)
	}
	return netMetrics.Net.Metrics, nil
}

func sendTopology(w http.ResponseWriter, r *http.Request) {
	// Create topology
	topology := dif.NewTopology().SetUpdateTime()
	topology.Scope = scope

	tests, err := getTests()
	if err != nil {
		glog.Errorf("Failed to get test: %v", err)
		sendFailure(w, r, err)
		return
	}

	for _, test := range tests {
		hrefs := test.APILinks
		url := test.URL
		for _, href := range hrefs {
			if strings.HasPrefix(href.Href, netMetricsURL) {
				metrics, err := getMetrics(href.Href)
				if err != nil {
					glog.Errorf("Failed to get metrics: %v", err)
					break
				}
				for _, metric := range metrics {
					// Create service entity
					service := dif.NewDIFEntity(test.TestName, "service").Matching(metric.ServerIP)
					capacity := float64(defaultCapacity)
					service.AddMetrics("responseTime", []*dif.DIFMetricVal{{
						Average:  &metric.AvgLatency,
						Max:      &metric.MaxLatency,
						Capacity: &capacity,
						Key:      &url,
					}})
					topology.AddEntity(service)
					break
				}
			}
		}
	}

	// Marshal to json
	result, err := json.Marshal(topology)
	if err != nil {
		glog.Errorf("Failed to marshal %v into json: %v.", topology, err)
		sendFailure(w, r, fmt.Errorf(
			"failed to marshal %v into json: %v", topology, err))
		return
	}
	glog.Infof("%s", result)
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

// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func main() {
	http.HandleFunc(path, sendTopology)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		glog.Fatalf("Failed to create server: %v.", err)
	}
}
