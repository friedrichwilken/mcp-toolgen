package widgets_readonly

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WidgetClient provides operations for Widget custom resources
type WidgetClient struct {
	client    client.Client
	namespace string
}

// NewWidgetClient creates a new client for Widget resources
func NewWidgetClient(c client.Client, namespace string) *WidgetClient {
	return &WidgetClient{
		client:    c,
		namespace: namespace,
	}
}

// Create creates a new Widget resource
func (c *WidgetClient) Create(ctx context.Context, obj *Widget) error {
	return c.client.Create(ctx, obj)
}
