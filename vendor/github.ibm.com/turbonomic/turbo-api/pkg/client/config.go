package client

import (
	"net/url"
)

type Config struct {
	// ServerAddress is A URL, including <Scheme>://<IP>:<Port>.
	serverAddress *url.URL
	// For Basic authentication.
	basicAuth *BasicAuthentication
	// For proxy
	proxy        string
	clientId     string
	clientSecret string
}

type ConfigBuilder struct {
	serverAddress *url.URL
	basicAuth     *BasicAuthentication
	proxy         string
	clientId      string
	clientSecret  string
}

func NewConfigBuilder(serverAddress *url.URL) *ConfigBuilder {
	return &ConfigBuilder{
		serverAddress: serverAddress,
	}
}

func (cb *ConfigBuilder) SetProxy(proxy string) *ConfigBuilder {
	cb.proxy = proxy
	return cb
}

func (cb *ConfigBuilder) SetClientId(clientId string) *ConfigBuilder {
	cb.clientId = clientId
	return cb
}

func (cb *ConfigBuilder) SetClientSecret(clientSecret string) *ConfigBuilder {
	cb.clientSecret = clientSecret
	return cb
}

func (cb *ConfigBuilder) BasicAuthentication(usrn, passd string) *ConfigBuilder {
	cb.basicAuth = &BasicAuthentication{
		username: usrn,
		password: passd,
	}
	return cb
}

func (cb *ConfigBuilder) Create() *Config {
	return &Config{
		serverAddress: cb.serverAddress,
		basicAuth:     cb.basicAuth,
		proxy:         cb.proxy,
		clientId:      cb.clientId,
		clientSecret:  cb.clientSecret,
	}
}
