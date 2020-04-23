package discovery

import (
	"bufio"
	"bytes"
	"github.com/tamerh/jsparser"

	//"bufio"
	//"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/dataingestionframework/data"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	FILE_URL = "file://"
	HTTP_URL = "http://"
)

type MetricDataSource interface {
	GetMetricData() (*data.Topology, error)
	GetMetricEndpoint() string
}

func ValidateMetricDataSource(metricEndpoint string) error {
	metricEndpoint = strings.TrimSpace(metricEndpoint)

	var valid bool
	switch {
	// HTTP endpoint - send request and retrieve the JSON response
	case strings.HasPrefix(metricEndpoint, HTTP_URL):
		valid = true
	// File based
	case strings.HasPrefix(metricEndpoint, FILE_URL):
		valid = true
	default:
		valid = false
		glog.Errorf("Unsupported metric endpoint: %s", metricEndpoint)
	}

	if valid {
		return nil
	}
	return fmt.Errorf("unsupported metric endpoint: %s", metricEndpoint)
}

func CreateMetricDataSource(metricEndpoint string) MetricDataSource {
	var metricDataSource MetricDataSource
	metricEndpoint = strings.TrimSpace(metricEndpoint)

	switch {
	// HTTP endpoint - send request and retrieve the JSON response
	case strings.HasPrefix(metricEndpoint, HTTP_URL):
		metricDataSource = NewHTTPBasedMetricDatSource(metricEndpoint)
	// File based
	case strings.HasPrefix(metricEndpoint, FILE_URL):
		metricDataSource = NewHFileBasedMetricDatSource(metricEndpoint)
	default:
		glog.Errorf("Unsupported metric endpoint: %s", metricEndpoint)
	}

	return metricDataSource
}

type HTTPBasedMetricDatSource struct {
	metricsUrl string
}

func NewHTTPBasedMetricDatSource(metricEndpoint string) *HTTPBasedMetricDatSource {
	return &HTTPBasedMetricDatSource{
		metricsUrl: metricEndpoint,
	}
}
func (md *HTTPBasedMetricDatSource) GetMetricEndpoint() string {
	return md.metricsUrl
}

func (md *HTTPBasedMetricDatSource) GetMetricData() (*data.Topology, error) {
	params := map[string]string{}
	var resp []byte
	resp, err := sendRequest(md.metricsUrl, params)
	if err != nil {
		return nil, err
	}

	//topology, err := loadJSON(md.metricsUrl, resp)
	topology, err := loadJSONStream(md.metricsUrl, resp)

	return topology, nil
}

// Send a request to the given endpoint. Params are encoded as query parameters
func sendRequest(endpoint string, params map[string]string) ([]byte, error) {
	url, err := encodeRequest(endpoint, params)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("Sending request to %s", url)

	resp, err := http.Get(url)
	if err != nil {
		glog.Errorf("Failed getting response from %s: %v", url, err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Error reading the response %v: %v", resp, err)
		return nil, err
	}
	glog.V(4).Infof("Received response: %s", string(body))
	return body, nil
}

func encodeRequest(endpoint string, params map[string]string) (string, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	return req.URL.String(), nil
}

type FileBasedMetricDatSource struct {
	metricsFile string
}

func (md *FileBasedMetricDatSource) GetMetricEndpoint() string {
	return md.metricsFile
}

func NewHFileBasedMetricDatSource(metricEndpoint string) *FileBasedMetricDatSource {
	return &FileBasedMetricDatSource{
		metricsFile: metricEndpoint,
	}
}

func (md *FileBasedMetricDatSource) GetMetricData() (*data.Topology, error) {
	var resp []byte
	filename := string(md.metricsFile[7:])
	resp, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("file error: %v", err.Error())
	}

	topology, err := loadJSON(md.metricsFile, resp)

	return topology, err
}

func loadJSON(metricEndpoint string, resp []byte) (*data.Topology, error) {

	var topology *data.Topology
	err := json.Unmarshal(resp, &topology)
	if err != nil {
		glog.Errorf("Topology JSON unmarshall error: %v\n", err)
		return nil, fmt.Errorf("json unmarshall error: %v", err.Error())
	}
	if topology.Entities != nil {
		glog.Infof("%s: parsed %d entities\n", metricEndpoint, len(topology.Entities))
	}

	return topology, nil
}

func loadJSONStream(metricEndpoint string, resp []byte) (*data.Topology, error) {

	rb := bytes.NewBuffer(resp)
	reader := bufio.NewReader(rb)
	parser := jsparser.NewJSONParser(reader, "Scope")
	var scope string
	for json := range parser.Stream() {
		glog.Infof("Scope %++v\n", json.StringVal)
		scope = json.StringVal
	}

	rb = bytes.NewBuffer(resp)
	reader = bufio.NewReader(rb)
	parser = jsparser.NewJSONParser(reader, "topology")
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
		glog.V(4).Infof("%v", entity)
		entities = append(entities, entity)
	} //end of topology

	var topology *data.Topology
	topology = &data.Topology{
		Version:    "",
		Updatetime: 0,
		Scope:      scope,
		Source:     scope,
		Entities:   entities,
	}

	if topology.Entities != nil {
		glog.Infof("%s: parsed %d entities", metricEndpoint, len(topology.Entities))
	}
	return topology, nil

}
