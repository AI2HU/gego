package shared

import (
	"github.com/gin-gonic/gin"
)

// ParseEnabledFilter parses the enabled query parameter and returns a pointer to bool or nil
func ParseEnabledFilter(c *gin.Context) *bool {
	enabledStr := c.Query("enabled")
	if enabledStr == "" {
		return nil
	}

	switch enabledStr {
	case "true":
		return &[]bool{true}[0]
	case "false":
		return &[]bool{false}[0]
	default:
		return nil
	}
}
