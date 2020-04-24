package client

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-api/pkg/api"
	"net/http"
)

// APIClient connects to api service through ingress
type APIClient struct {
	*RESTClient
	SessionCookie *http.Cookie
}

const (
	SessionCookie string = "JSESSIONID"
)

// Discover a target using API
// This function is called by turboctl which is not being maintained
func (c *APIClient) DiscoverTarget(uuid string) (*Result, error) {
	response, err := c.Post().Resource(api.Resource_Type_Targets).Name(uuid).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to discover target %s: %s", uuid, err)
	}
	if response.statusCode != 200 {
		return nil, buildResponseError("target discovery", response.status, response.body)
	}
	return &response, nil
}

// AddTarget adds a target via api service
func (c *APIClient) AddTarget(target *api.Target) error {
	// Login first
	if _, err := c.login(); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	// Find if the target exists
	targetExists, _ := c.findTarget(target)
	if targetExists {
		glog.V(2).Infof("Target %v already exists.", getTargetId(target))
		return nil
	}

	// Construct the Target required by the rest api
	targetData, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("failed to marshall target instance: %v", err)
	}

	// Create the rest api request
	request := c.Post().Resource(api.Resource_Type_Targets).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Header("Cookie", fmt.Sprintf("%s=%s", c.SessionCookie.Name, c.SessionCookie.Value)).
		Data(targetData)

	glog.V(4).Infof("[AddTarget] %v.", request)
	glog.V(4).Infof("[AddTarget] Data: %s.", targetData)

	// Execute the request
	response, err := request.Do()
	if err != nil {
		return fmt.Errorf("request %v failed: %s", request, err)
	}
	glog.V(4).Infof("Response %+v.", response)

	if response.statusCode != 200 {
		return buildResponseError("target addition", response.status, response.body)
	}

	glog.V(2).Infof("Successfully added target via API service: %v.", response)
	return nil
}

// Login to the Turbo API server
func (c *APIClient) login() (*Result, error) {
	if c.SessionCookie != nil {
		// Already logged in
		return nil, nil
	}
	if c.basicAuth == nil ||
		c.basicAuth.username == "" ||
		c.basicAuth.password == "" {
		return nil, fmt.Errorf("missing username and password")
	}
	var data []byte
	data = []byte(fmt.Sprintf("username=%s&password=%s", c.basicAuth.username, c.basicAuth.password))
	request := c.Post().Resource("login").
		Header("Content-Type", "application/x-www-form-urlencoded").
		Data(data)

	response, err := request.Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to login  %s: %s", c.baseURL, err)
	}
	if response.statusCode != 200 {
		return nil, buildResponseError("Turbo server login", response.status, response.body)
	}

	// Save the session cookie
	sessionCookie, ok := response.cookies[SessionCookie]
	if ok {
		c.SessionCookie = sessionCookie
		glog.V(2).Infof("Successfully logged in to Turbonomic server.")
		glog.V(4).Infof("Session Cookie = %s:%s.", c.SessionCookie.Name, c.SessionCookie.Value)
	} else {
		return nil, buildResponseError("Invalid session cookie", response.status,
			fmt.Sprintf("%s", response.cookies))
	}
	return &response, nil
}

func (c *APIClient) findTarget(target *api.Target) (bool, error) {
	// Get a list of targets from the Turbo server
	request := c.Get().Resource(api.Resource_Type_Targets).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Header("Cookie", fmt.Sprintf("%s=%s", c.SessionCookie.Name, c.SessionCookie.Value))

	response, err := request.Do()
	if err != nil {
		glog.Errorf("Failed to execute find target request: %s.", err)
		return false, fmt.Errorf("failed to execute find target request: %v", err)
	}

	glog.V(4).Infof("Received response from find target request %v: %+v.",
		request, response)

	if response.statusCode != 200 {
		return false, buildResponseError("find target", response.status, response.body)
	}

	// The target identifier for the given target
	targetId := getTargetId(target)

	// Parse the response - list of targets
	var targetList []interface{}
	if err := json.Unmarshal([]byte(response.body), &targetList); err != nil {
		return false, fmt.Errorf("failed to unmarshall find target response: %v", err)
	}

	// Iterate over the list of targets to look for the given target
	// by comparing the category, target type and identifier fields
	for _, tgt := range targetList {
		targetInstance, isMap := tgt.(map[string]interface{})
		if !isMap {
			continue
		}
		var category, targetType, tgtId string
		category, _ = targetInstance["category"].(string)
		targetType, _ = targetInstance["type"].(string)
		inputFields, ok := targetInstance["inputFields"].([]interface{})
		if ok {
			// array of InputFields
			for _, value := range inputFields {
				// an instance of InputField
				inputFieldMap, isMap := value.(map[string]interface{})
				// Get the target identifier value from the
				// input field named 'targetIdentifier'
				if isMap {
					field, isNameField := inputFieldMap["name"].(string)
					//
					if isNameField && field == "targetIdentifier" {
						tgtId, _ = inputFieldMap["value"].(string)
					}
				}
			}
		}
		glog.V(4).Infof("%s::%s::%s", category, targetType, tgtId)
		if target.Category == category && target.Type == targetType && tgtId == targetId {
			return true, nil
		}
	}
	return false, nil
}
