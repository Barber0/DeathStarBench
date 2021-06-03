package registry

import (
	consul "github.com/hashicorp/consul/api"
	"strings"
)

// NewClient returns a new Client with connection to consul
func NewClient(addr string) (*Client, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = addr

	c, err := consul.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{c}, nil
}

// Client provides an interface for communicating with registry
type Client struct {
	*consul.Client
}

// Register a service with registry
func (c *Client) Register(name string, ip string, port int) error {
	srvNames := strings.Split(name, "-")

	srvLabel := "app|"
	if len(srvNames) > 1 {
		srvLabel += srvNames[1]
	} else {
		srvLabel += name
	}

	reg := &consul.AgentServiceRegistration{
		ID:      name,
		Name:    name,
		Port:    port,
		Address: ip,
		Tags:    []string{srvLabel},
	}
	return c.Agent().ServiceRegister(reg)
}

// Deregister removes the service address from registry
func (c *Client) Deregister(id string) error {
	return c.Agent().ServiceDeregister(id)
}
