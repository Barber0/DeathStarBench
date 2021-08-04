package registry

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"math/rand"
	"time"
)

// NewClient returns a new Client with connection to consul
func NewClient(addr string) (*Client, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = addr

	c, err := consul.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: c,
	}, nil
}

// Client provides an interface for communicating with registry
type Client struct {
	*consul.Client
	registryId string
}

// Register a service with registry
func (c *Client) Register(name string, ip string, port int) error {
	if c.registryId == "" {
		c.registryId = fmt.Sprintf("%s-%d-%d", name, time.Now().UnixNano(), rand.Int())
	}

	reg := &consul.AgentServiceRegistration{
		ID:      c.registryId,
		Name:    name,
		Port:    port,
		Address: ip,
	}
	return c.Agent().ServiceRegister(reg)
}

// Deregister removes the service address from registry
func (c *Client) Deregister() error {
	return c.Agent().ServiceDeregister(c.registryId)
}
