package auth

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/models"
)

const (
	contextKeyUserID = "auth_user_id"
	contextKeyRole   = "auth_role"
)

func SetAuthContext(c *gin.Context, userID string, role models.Role) {
	c.Set(contextKeyUserID, userID)
	c.Set(contextKeyRole, role)
}

func GetUserID(c *gin.Context) (string, error) {
	value, exists := c.Get(contextKeyUserID)
	if !exists {
		return "", errors.New("user id not found in context")
	}
	userID, ok := value.(string)
	if !ok || userID == "" {
		return "", errors.New("invalid user id in context")
	}
	return userID, nil
}

func GetRole(c *gin.Context) (models.Role, error) {
	value, exists := c.Get(contextKeyRole)
	if !exists {
		return "", errors.New("role not found in context")
	}
	role, ok := value.(models.Role)
	if !ok || !role.Valid() {
		return "", errors.New("invalid role in context")
	}
	return role, nil
}
