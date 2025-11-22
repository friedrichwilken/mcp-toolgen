package clusterwidgets

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GlobalConfigClient provides operations for GlobalConfig custom resources
type GlobalConfigClient struct {
	client    client.Client
	namespace string
}

// NewGlobalConfigClient creates a new client for GlobalConfig resources
func NewGlobalConfigClient(c client.Client, namespace string) *GlobalConfigClient {
	return &GlobalConfigClient{
		client:    c,
		namespace: namespace,
	}
}

// Create creates a new GlobalConfig resource
func (c *GlobalConfigClient) Create(ctx context.Context, obj *GlobalConfig) error {
	return c.client.Create(ctx, obj)
}
