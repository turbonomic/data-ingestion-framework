package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/avast/retry-go"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-api/pkg/api"
)

var (
	retryAttempts = 5
	retryDelay    = 1 * time.Second
)

// TPClient connects to topology processor service
type TPClient struct {
	*RESTClient
}

// DiscoverTarget adds a target via Topology Processor service
func (c *TPClient) DiscoverTarget(uuid string) (*Result, error) {
	// Not implemented
	return &Result{}, nil
}

// AddTarget adds a target via topology processor service
func (c *TPClient) AddTarget(target *api.Target) error {
	// Check if the given target already exists
	targetName := getTargetId(target)
	existingTarget, err := c.findTarget(targetName)
	if err != nil {
		return err
	}

	if existingTarget != nil {
		glog.V(2).Infof("Target %v already exists with ID %v.",
			targetName, existingTarget)
		return c.updateTarget(existingTarget, target)
	}

	probeID, err := c.getProbeID(target.Type, target.Category)
	if err != nil {
		return fmt.Errorf("failed to get probe ID: %v", err)
	}

	// Add the target which belongs to the probe with the given probeID
	glog.V(2).Infof("Starting to add target %v for probe %v with ID : %v",
		targetName, target.Type, probeID)

	// Construct the TargetSpec required by the rest api
	targetSpec := &api.TargetSpec{
		ProbeID:          probeID,
		DerivedTargetIDs: []string{},
		InputFields:      target.InputFields,
	}

	// Create the rest api request
	targetData, err := json.Marshal(targetSpec)
	if err != nil {
		return fmt.Errorf("failed to marshall target spec: %v", err)
	}
	request := c.Post().Resource(api.Resource_Type_Target).
		Header("Content-Type", "application/json;charset=UTF-8").
		Header("Accept", "application/json;charset=UTF-8").
		Data(targetData)

	glog.V(4).Infof("[AddTarget] %v", request)
	glog.V(4).Infof("[AddTarget] Data: %s", targetData)

	// Execute the request
	response, err := request.Do()
	if err != nil {
		return fmt.Errorf("request %v failed: %s", request, err)
	}
	glog.V(4).Infof("Response %+v", response)

	if response.statusCode != 200 {
		return buildResponseError("target addition", response.status, response.body)
	}

	// Unmarshal the response and parse out the target ID
	var targetInfo api.TargetInfo
	_ = json.Unmarshal([]byte(response.body), &targetInfo)
	glog.V(2).Infof("Successfully added target via Topology Processor service. Target ID: %v",
		targetInfo.TargetID)

	return nil
}

func (c *TPClient) findTarget(targetName string) (*api.TargetInfo, error) {
	// Get a list of targets from the Turbo server
	request := c.Get().Resource(api.Resource_Type_Target).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json")

	response, err := request.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to execute find target request %v: %v",
			request, err)
	}

	glog.V(4).Infof("Received response from find target request %v: %+v",
		request, response)

	if response.statusCode != 200 {
		return nil, buildResponseError("find target", response.status, response.body)
	}

	var targetsMap map[string][]api.TargetInfo
	if err := json.Unmarshal([]byte(response.body), &targetsMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal get target response: %v", err)
	}
	glog.V(4).Infof("Successfully parsed response body into targetsMap: %v",
		spew.Sdump(targetsMap))
	targets, found := targetsMap["targets"]
	if !found {
		return nil, fmt.Errorf("failed to find key \"targets\" from response")
	}

	for _, target := range targets {
		if target.TargetSpec == nil {
			continue
		}
		for _, inputField := range target.TargetSpec.InputFields {
			if inputField.Name == "targetIdentifier" &&
				inputField.Value == targetName {
				return &target, nil
			}
		}
	}
	glog.V(4).Infof("target %v does not exist", targetName)
	return nil, nil
}

func (c *TPClient) getProbeID(probeType, probeCategory string) (int64, error) {
	// Execute get probe request
	request := c.Get().Resource(api.Resource_Type_Probe).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json")
	// Get the Probe ID based on probe type and probe category
	// Retry 5 times with 1 second delay
	var response Result
	var err error
	errs := retry.Do(
		func() error {
			response, err = request.Do()
			if err != nil {
				return fmt.Errorf("failed to execute get probe request %+v: %v",
					request, err)
			}
			if response.statusCode != 200 {
				return buildResponseError("get probe", response.status, response.body)
			}
			return nil
		},
		retry.Attempts(uint(retryAttempts)),
		retry.OnRetry(func(n uint, err error) {
			glog.Warningf("Retry #%d: %v", n, err)
		}),
		retry.Delay(retryDelay),
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
	)
	if errs != nil {
		return 0, errs
	}
	glog.V(4).Infof("Received response from get probe request %+v: %+v", request, response)
	// Parse the response - list of probes
	var probesMap map[string][]api.ProbeDescription
	if err := json.Unmarshal([]byte(response.body), &probesMap); err != nil {
		return 0, fmt.Errorf("failed to unmarshal get probe response: %v", err)
	}
	probes, found := probesMap["probes"]
	if !found {
		return 0, fmt.Errorf("failed to find key \"probes\" from response")
	}
	for _, probe := range probes {
		if probe.Category == probeCategory &&
			probe.Type == probeType {
			return probe.ID, nil
		}
	}
	return 0, fmt.Errorf("failed to find probe with category %v and type %v",
		probeCategory, probeType)
}

func (c *TPClient) updateTarget(existingTarget *api.TargetInfo, input *api.Target) error {
	inputFieldsData, err := json.Marshal(input.InputFields)
	if err != nil {
		return fmt.Errorf("failed to marshall input fields array: %v", err)
	}
	// Create the rest api request
	request := c.Put().Resource(api.Resource_Type_Target).Name(strconv.FormatInt(existingTarget.TargetID, 10)).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Data(inputFieldsData)

	glog.V(4).Infof("[UpdateTarget] %v", request)
	glog.V(4).Infof("[UpdateTarget] Data: %s", inputFieldsData)

	// Execute the request
	response, err := request.Do()
	if err != nil {
		return fmt.Errorf("request %v failed: %s", request, err)
	}
	glog.V(4).Infof("Response %+v", response)

	if response.statusCode != 200 {
		return buildResponseError("target update", response.status, response.body)
	}
	glog.V(2).Infof("Successfully updated target via Topology Processor service: %v.", existingTarget.TargetID)
	return nil
}
