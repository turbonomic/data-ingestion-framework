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
	proxy string
}

type ConfigBuilder struct {
	serverAddress *url.URL
	basicAuth     *BasicAuthentication
	proxy         string
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
	}
}
