package widgets

// WidgetToolset provides MCP tools for managing Widget custom resources
type WidgetToolset struct{}

// GetName returns the name of this toolset
func (t *WidgetToolset) GetName() string {
	return "widgets"
}

// GetDescription returns the description of this toolset
func (t *WidgetToolset) GetDescription() string {
	return "Tools for managing Widget custom resources"
}
