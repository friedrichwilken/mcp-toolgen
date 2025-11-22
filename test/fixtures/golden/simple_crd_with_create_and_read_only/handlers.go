package widgets_readonly

import (
	"fmt"
)

// Basic handler implementation
func HandleWidgetOperations(operation string, params map[string]interface{}) (interface{}, error) {
	return fmt.Sprintf("Operation %s not implemented for Widget", operation), nil
}
