package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-api/pkg/api"
	"mime/multipart"
	"net/http"
	"strings"
)

// APIClient connects to api service through ingress
type APIClient struct {
	*RESTClient
	SessionCookie *http.Cookie
	ClientId      string
	ClientSecret  string
}

const (
	SessionCookie string = "JSESSIONID"
)

type HydraTokenBody struct {
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
}

func (c *APIClient) GetJwtToken(hydraToken string) (string, error) {
	if hydraToken == "" {
		glog.V(4).Infof("The hydra token is empty")
		return "", nil
	}
	// the format of getting jwtToken with auth requires the following header
	// key: x-oauth2, value: "hydra"
	// key: x-auth-token, value: hydra access token
	request := c.Post().Resource(api.Resource_Type_auth_token).Header("x-oauth2", "hydra").Header("x-auth-token", hydraToken)
	// Execute the request
	response, err := request.Do()
	if err != nil {
		return "", fmt.Errorf("request to exchange for auth token %v failed: %s", request, err)
	}
	if response.statusCode == 401 {
		// When we receive the 401 status code, means that the credentials are not valid.
		// We return error, so getJwtToken() method in tap_service will continue
		// to retry authentication until the credentials are corrected
		return "", fmt.Errorf("Auth service authentication failed using the given client_id and secret")
	}
	if response.statusCode == 403 {
		// When we receive the 403 status code, meaning the security feature is currently not available.
		// We will retry websocket connection without the JWTToken
		glog.Errorf("Auth service is not accessible or disabled [%v:%s]", response.statusCode, response.status)
		return "", nil
	}
	if response.statusCode == 502 {
		// When we receive the 502 status code, meaning the auth service is currently not available.
		// We return error, so getJwtToken() method in tap_service will continue
		// to retry authentication until the service is restored
		return "", fmt.Errorf("Auth service is not available [%v:%s]", response.statusCode, response.status)
	}
	return response.body, nil
}

func (c *APIClient) GetHydraAccessToken() (string, error) {
	if c.ClientId == "" || c.ClientSecret == "" {
		glog.V(4).Infof("The client id or client secret are not provided")
		return "", nil
	}
	// Create the form-data format payload
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	writer.WriteField("client_id", c.ClientId)
	writer.WriteField("client_secret", c.ClientSecret)
	writer.WriteField("grant_type", "client_credentials")
	err := writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to Close the writer: %v", err)
	}
	// Create the rest api request
	// the format of get hydra access token requires the following in the body, as form-data format
	// key: client_id, value: client id value
	// key: client_secret, value: client secret value
	// key: grant_type, value: client_credentials
	request := c.Post().Resource(api.Resource_Type_hydra_token).Header("Content-Type", writer.FormDataContentType()).BufferData(payload)
	// Execute the request
	response, err := request.Do()
	if err != nil {
		return "", fmt.Errorf("failed to get hydra access token: %s", err)
	}
	if response.statusCode == 401 {
		// When we receive the 401 status code, means that the credentials are not valid.
		// We return error, so getJwtToken() method in tap_service will continue
		// to retry authentication until the credentials are corrected
		return "", fmt.Errorf("Hydra service authentication failed using the given client_id and secret. " +
			"Redeploy the secret containing the correct credentials and restart the probe pod")
	}
	if response.statusCode == 502 {
		// When we receive the 502 status code, meaning the hydra service is currently not available.
		// We return error, so getJwtToken() method in tap_service will continue
		// to retry authentication until the service is restored
		return "", fmt.Errorf("Hydra service is not available [%v:%s]", response.statusCode, response.status)
	}
	if response.statusCode == 403 {
		// When we receive the 403 status code, means that the hydra service is currently not available,
		// This is when the security feature is not available.
		// We are not returning error here to handle the case when customer enabled probe security in the first place then disabled
		// In the case above, there'll be client_id and secret in the k8s secret, but we shouldn't use them in websocket connection
		// If the hydra service is temporarily not accessible, we have retry in performWebSocketConnection in turbo-go-sdk
		glog.Errorf("Hydra service is not accessible or disabled [%v:%s]", response.statusCode, response.status)
		return "", nil
	}
	var hydraToken HydraTokenBody
	err = json.Unmarshal([]byte(response.body), &hydraToken)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshall get hydra token response: %v", err)
	}
	return hydraToken.AccessToken, nil
}

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
