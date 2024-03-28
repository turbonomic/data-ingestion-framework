package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.ibm.com/turbonomic/turbo-api/pkg/api"
)

var (
	API                   = "api"
	TopologyProcessor     = "topology-processor"
	HYDRA                 = "hydra"
	AUTH                  = "auth"
	APIPath               = "/vmturbo/rest/"
	TopologyProcessorPath = "/"
	HydraPath             = "/oauth2/"
	AuthPath              = "/vmturbo/auth/"

	defaultRESTAPIEndpoints = map[string]string{
		API:               APIPath,
		TopologyProcessor: TopologyProcessorPath,
		HYDRA:             HydraPath,
		AUTH:              AuthPath,
	}
)

type Client interface {
	AddTarget(target *api.Target) error
	DiscoverTarget(uuid string) (*Result, error)
	GetHydraAccessToken() (string, error)
	GetJwtToken(hydraToken string) (string, error)
}

// TurboClient manages REST clients to Turbonomic services
type TurboClient struct {
	clients map[string]Client // A map that maps service name to REST client
}

func NewTurboClient(c *Config) (*TurboClient, error) {
	// Build httpClient, which can be shared by multiple connections
	httpClient := http.DefaultClient
	proxy := c.proxy
	if proxy != "" || c.serverAddress.Scheme == "https" {
		var tr http.Transport
		if c.serverAddress.Scheme == "https" {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

		if proxy != "" {
			//Check if the proxy server requires authentication or not
			//Authenticated proxy format: http://username:password@ip:port
			//Non-Aunthenticated proxy format: http://ip:port
			if strings.Index(proxy, "@") != -1 {
				//Extract the username password portion, with @
				usernamePassword := proxy[strings.Index(proxy, "//")+2 : strings.LastIndex(proxy, "@")+1]
				username := usernamePassword[:strings.Index(usernamePassword, ":")]
				password := usernamePassword[strings.Index(usernamePassword, ":")+1 : strings.LastIndex(usernamePassword, "@")]
				//Extract Proxy address by remove the username_password
				proxyAddr := strings.ReplaceAll(proxy, usernamePassword, "")
				proxyURL, err := url.Parse(proxyAddr)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse proxy\n")
				}
				proxyURL.User = url.UserPassword(username, password)
				tr.Proxy = http.ProxyURL(proxyURL)
			} else {
				proxyURL, err := url.Parse(proxy)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse proxy\n")
				}
				tr.Proxy = http.ProxyURL(proxyURL)
			}
		}
		httpClient = &http.Client{Transport: &tr}
	}
	// Build the map of clients
	turboClient := &TurboClient{
		clients: make(map[string]Client),
	}
	for service, endpoint := range defaultRESTAPIEndpoints {
		turboClient.clients[service] = newClient(httpClient, c.serverAddress,
			c.basicAuth, service, endpoint, c.clientId, c.clientSecret)
	}
	return turboClient, nil
}

func newClient(client *http.Client, baseURL *url.URL, basicAuth *BasicAuthentication,
	service, endpoint, clientId, clientSecret string) Client {
	restClient := NewRESTClient(client, baseURL, endpoint).BasicAuthentication(basicAuth)
	if service == TopologyProcessor {
		// Create a Turbo client without authentication
		return &TPClient{
			restClient,
		}
	}
	// Create a Turbo client based on basic authentication
	return &APIClient{
		restClient.BasicAuthentication(basicAuth),
		nil, clientId, clientSecret,
	}
}

// GetHydraAccessToken gets the access token from Hydra service
func (turboClient *TurboClient) GetHydraAccessToken() (string, error) {
	client, ok := turboClient.clients[HYDRA]
	if !ok {
		return "", fmt.Errorf("client for service %v is not registered", HYDRA)
	}
	return client.GetHydraAccessToken()
}

// GetJwtToken gets the JwtToken from Hydra access token
func (turboClient *TurboClient) GetJwtToken(hydraToken string) (string, error) {
	client, ok := turboClient.clients[AUTH]
	if !ok {
		return "", fmt.Errorf("client for service %v is not registered", AUTH)
	}
	return client.GetJwtToken(hydraToken)
}

// AddTarget adds a target via a given service
func (turboClient *TurboClient) AddTarget(target *api.Target, service string) error {
	client, ok := turboClient.clients[service]
	if !ok {
		return fmt.Errorf("client for service %v is not registered", service)
	}
	return client.AddTarget(target)
}

// AddTarget discovers a target via a given service
func (turboClient *TurboClient) DiscoverTarget(uuid, service string) (*Result, error) {
	client, ok := turboClient.clients[service]
	if !ok {
		return nil, fmt.Errorf("client for service %v is not registered", service)
	}
	return client.DiscoverTarget(uuid)
}

// Get the target identifier for the given target
func getTargetId(target *api.Target) string {
	for _, inputField := range target.InputFields {
		field := inputField.Name
		if field == "targetIdentifier" {
			tgtId := inputField.Value
			return tgtId
		}
	}
	return ""
}

func buildResponseError(requestDesc string, status string, content string) error {
	errorMsg := fmt.Sprintf("unsuccessful %s response: %s.", requestDesc, status)
	errorDTO, err := parseAPIErrorDTO(content)
	if err == nil && errorDTO.Message != "" {
		// Add error message only if we can parse result content to errorDTO.
		errorMsg = errorMsg + fmt.Sprintf(" %s.", errorDTO.Message)
	}
	return fmt.Errorf("%s", errorMsg)
}
