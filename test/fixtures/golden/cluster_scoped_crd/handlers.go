package clusterwidgets

import (
	"fmt"
)

// Basic handler implementation
func HandleGlobalConfigOperations(operation string, params map[string]interface{}) (interface{}, error) {
	return fmt.Sprintf("Operation %s not implemented for GlobalConfig", operation), nil
}
