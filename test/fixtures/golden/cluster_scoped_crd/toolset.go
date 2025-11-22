package clusterwidgets

// GlobalConfigToolset provides MCP tools for managing GlobalConfig custom resources
type GlobalConfigToolset struct{}

// GetName returns the name of this toolset
func (t *GlobalConfigToolset) GetName() string {
	return "globalconfigs"
}

// GetDescription returns the description of this toolset
func (t *GlobalConfigToolset) GetDescription() string {
	return "Tools for managing GlobalConfig custom resources"
}
