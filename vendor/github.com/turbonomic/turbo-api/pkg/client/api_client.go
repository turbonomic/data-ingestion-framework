package client

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-api/pkg/api"
	"net/http"
	"strings"
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
	existingTarget, err := c.findTarget(target)
	if err != nil {
		return err
	}
	// Update the target if it already exists
	if existingTarget != nil {
		glog.V(2).Infof("Target %v already exists.", getTargetId(target))
		if target.Type == existingTarget.Type {
			return c.updateTarget(existingTarget, target)
		}
		glog.V(2).Infof("Delete and re-add the target since the probe type has changed "+
			"(old probe type: %v, new probe type: %v, old target: %v, new target: %v)",
			existingTarget.Type, target.Type, existingTarget, target)
		if err := c.deleteTarget(existingTarget); err != nil {
			return fmt.Errorf("failed to delete target %v of type %v which is necessary due to probe type changed"+
				" to %v; error: %v", existingTarget.DisplayName, existingTarget.Type, target.Type, err)
		}
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

func (c *APIClient) findTarget(target *api.Target) (*api.Target, error) {
	// Get a list of targets from the Turbo server
	request := c.Get().Resource(api.Resource_Type_Targets).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Header("Cookie", fmt.Sprintf("%s=%s", c.SessionCookie.Name, c.SessionCookie.Value))

	response, err := request.Do()
	if err != nil {
		glog.Errorf("Failed to execute find target request: %s.", err)
		return nil, fmt.Errorf("failed to execute find target request: %v", err)
	}

	glog.V(4).Infof("Received response from find target request %v: %+v.",
		request, response)

	if response.statusCode != 200 {
		return nil, buildResponseError("find target", response.status, response.body)
	}

	// Parse the response - list of targets
	var targetList []api.Target
	if err := json.Unmarshal([]byte(response.body), &targetList); err != nil {
		return nil, fmt.Errorf("failed to unmarshall find target response: %v", err)
	}

	// The target identifier for the given target
	targetId := getTargetId(target)

	// Iterate over the list of targets to look for the given target
	// by comparing the category, target type and identifier fields
	// target type is regarded the same if the old one only differs by an extra suffix
	for _, tgt := range targetList {
		// array of InputFields
		for _, inputField := range tgt.InputFields {
			if inputField.Name == "targetIdentifier" &&
				inputField.Value == targetId &&
				tgt.Category == target.Category &&
				strings.HasPrefix(tgt.Type, target.Type) {
				return &tgt, nil
			}
		}
	}

	glog.V(4).Infof("target %v does not exist", targetId)
	return nil, nil
}

func (c *APIClient) updateTarget(existing, input *api.Target) error {
	// Update the input fields
	existing.InputFields = input.InputFields
	targetData, err := json.Marshal(existing)
	if err != nil {
		return fmt.Errorf("failed to marshall target instance: %v", err)
	}

	// Create the rest api request
	request := c.Put().Resource(api.Resource_Type_Targets).Name(existing.UUID).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Header("Cookie", fmt.Sprintf("%s=%s", c.SessionCookie.Name, c.SessionCookie.Value)).
		Data(targetData)

	glog.V(4).Infof("[UpdateTarget] %v.", request)
	glog.V(4).Infof("[UpdateTarget] Data: %s.", targetData)

	// Execute the request
	response, err := request.Do()
	if err != nil {
		return fmt.Errorf("request %v failed: %s", request, err)
	}
	glog.V(4).Infof("Response %+v.", response)

	if response.statusCode != 200 {
		return buildResponseError("target update", response.status, response.body)
	}

	glog.V(2).Infof("Successfully updated target via API service.")
	return nil
}

// deleteTarget deletes an existing target
func (c *APIClient) deleteTarget(existing *api.Target) error {
	// Create the rest api request
	request := c.Delete().Resource(api.Resource_Type_Targets).Name(existing.UUID).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Header("Cookie", fmt.Sprintf("%s=%s", c.SessionCookie.Name, c.SessionCookie.Value))

	glog.V(4).Infof("[DeleteTarget] %v.", request)

	// Execute the request
	response, err := request.Do()
	if err != nil {
		return fmt.Errorf("request %v failed: %s", request, err)
	}
	glog.V(4).Infof("Response %+v.", response)

	if response.statusCode != 200 {
		return buildResponseError("target delete", response.status, response.body)
	}

	glog.V(2).Infof("Successfully deleted target %v of type %v via API service.",
		existing.DisplayName, existing.Type)
	return nil
}
