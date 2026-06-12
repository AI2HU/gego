package auth

import (
	"github.com/AI2HU/gego/internal/models"
)

type Permission string

const (
	PermLLMsRead       Permission = "llms:read"
	PermLLMsWrite      Permission = "llms:write"
	PermPromptsRead    Permission = "prompts:read"
	PermPromptsWrite   Permission = "prompts:write"
	PermSchedulesRead  Permission = "schedules:read"
	PermSchedulesWrite Permission = "schedules:write"
	PermStatsRead      Permission = "stats:read"
	PermSearchExecute  Permission = "search:execute"
	PermAuthProfile    Permission = "auth:profile"
)

type EndpointPolicy struct {
	Method     string
	Path       string
	Permission Permission
}

var RolePermissions = map[models.Role][]Permission{
	models.RoleAdmin: {
		PermLLMsRead, PermLLMsWrite,
		PermPromptsRead, PermPromptsWrite,
		PermSchedulesRead, PermSchedulesWrite,
		PermStatsRead, PermSearchExecute, PermAuthProfile,
	},
	models.RoleMember: {
		PermLLMsRead,
		PermPromptsRead,
		PermSchedulesRead,
		PermStatsRead,
		PermSearchExecute,
		PermAuthProfile,
	},
}

var EndpointPolicies = []EndpointPolicy{
	{Method: "GET", Path: "/llms", Permission: PermLLMsRead},
	{Method: "GET", Path: "/llms/:id", Permission: PermLLMsRead},
	{Method: "POST", Path: "/llms", Permission: PermLLMsWrite},
	{Method: "PUT", Path: "/llms/:id", Permission: PermLLMsWrite},
	{Method: "DELETE", Path: "/llms/:id", Permission: PermLLMsWrite},
	{Method: "GET", Path: "/prompts", Permission: PermPromptsRead},
	{Method: "GET", Path: "/prompts/:id", Permission: PermPromptsRead},
	{Method: "POST", Path: "/prompts", Permission: PermPromptsWrite},
	{Method: "PUT", Path: "/prompts/:id", Permission: PermPromptsWrite},
	{Method: "DELETE", Path: "/prompts/:id", Permission: PermPromptsWrite},
	{Method: "GET", Path: "/schedules", Permission: PermSchedulesRead},
	{Method: "GET", Path: "/schedules/:id", Permission: PermSchedulesRead},
	{Method: "POST", Path: "/schedules", Permission: PermSchedulesWrite},
	{Method: "PUT", Path: "/schedules/:id", Permission: PermSchedulesWrite},
	{Method: "DELETE", Path: "/schedules/:id", Permission: PermSchedulesWrite},
	{Method: "GET", Path: "/stats", Permission: PermStatsRead},
	{Method: "GET", Path: "/stats/urls", Permission: PermStatsRead},
	{Method: "GET", Path: "/stats/query-urls", Permission: PermStatsRead},
	{Method: "GET", Path: "/stats/keyword-domains", Permission: PermStatsRead},
	{Method: "POST", Path: "/search", Permission: PermSearchExecute},
	{Method: "GET", Path: "/auth/me", Permission: PermAuthProfile},
}

func HasPermission(role models.Role, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
