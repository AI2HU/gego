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
	PermStatsRead           Permission = "stats:read"
	PermSearchExecute       Permission = "search:execute"
	PermExclusionWordsRead  Permission = "exclusion-words:read"
	PermExclusionWordsWrite Permission = "exclusion-words:write"
	PermAuthProfile         Permission = "auth:profile"
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
		PermStatsRead, PermSearchExecute,
		PermExclusionWordsRead, PermExclusionWordsWrite,
		PermAuthProfile,
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
	{Method: "GET", Path: "/providers", Permission: PermLLMsRead},
	{Method: "GET", Path: "/providers/:provider/api-keys", Permission: PermLLMsRead},
	{Method: "POST", Path: "/providers/:provider/models", Permission: PermLLMsWrite},
	{Method: "GET", Path: "/models", Permission: PermLLMsRead},
	{Method: "GET", Path: "/models/:id", Permission: PermLLMsRead},
	{Method: "POST", Path: "/models", Permission: PermLLMsWrite},
	{Method: "PUT", Path: "/models/:id", Permission: PermLLMsWrite},
	{Method: "POST", Path: "/models/:id/test", Permission: PermLLMsRead},
	{Method: "DELETE", Path: "/models/:id", Permission: PermLLMsWrite},
	{Method: "GET", Path: "/prompts", Permission: PermPromptsRead},
	{Method: "POST", Path: "/prompts/generate", Permission: PermPromptsWrite},
	{Method: "GET", Path: "/prompts/:id", Permission: PermPromptsRead},
	{Method: "POST", Path: "/prompts", Permission: PermPromptsWrite},
	{Method: "PUT", Path: "/prompts/:id", Permission: PermPromptsWrite},
	{Method: "DELETE", Path: "/prompts/:id", Permission: PermPromptsWrite},
	{Method: "GET", Path: "/schedules", Permission: PermSchedulesRead},
	{Method: "GET", Path: "/schedules/:id", Permission: PermSchedulesRead},
	{Method: "POST", Path: "/schedules", Permission: PermSchedulesWrite},
	{Method: "PUT", Path: "/schedules/:id", Permission: PermSchedulesWrite},
	{Method: "DELETE", Path: "/schedules/:id", Permission: PermSchedulesWrite},
	{Method: "POST", Path: "/schedules/:id/run", Permission: PermSchedulesWrite},
	{Method: "GET", Path: "/scheduler/status", Permission: PermSchedulesRead},
	{Method: "POST", Path: "/scheduler/start", Permission: PermSchedulesWrite},
	{Method: "POST", Path: "/scheduler/stop", Permission: PermSchedulesWrite},
	{Method: "POST", Path: "/scheduler/reload", Permission: PermSchedulesWrite},
	{Method: "GET", Path: "/stats", Permission: PermStatsRead},
	{Method: "GET", Path: "/stats/urls", Permission: PermStatsRead},
	{Method: "GET", Path: "/stats/query-urls", Permission: PermStatsRead},
	{Method: "GET", Path: "/stats/keyword-domains", Permission: PermStatsRead},
	{Method: "GET", Path: "/exclusion-words", Permission: PermExclusionWordsRead},
	{Method: "POST", Path: "/exclusion-words", Permission: PermExclusionWordsWrite},
	{Method: "DELETE", Path: "/exclusion-words/:id", Permission: PermExclusionWordsWrite},
	{Method: "GET", Path: "/brands", Permission: PermExclusionWordsRead},
	{Method: "GET", Path: "/brands/suggestions", Permission: PermExclusionWordsRead},
	{Method: "POST", Path: "/brands", Permission: PermExclusionWordsWrite},
	{Method: "POST", Path: "/brands/map", Permission: PermExclusionWordsWrite},
	{Method: "PUT", Path: "/brands/:id", Permission: PermExclusionWordsWrite},
	{Method: "DELETE", Path: "/brands/:id", Permission: PermExclusionWordsWrite},
	{Method: "POST", Path: "/brands/:id/aliases", Permission: PermExclusionWordsWrite},
	{Method: "PUT", Path: "/brands/:id/aliases/:aliasId", Permission: PermExclusionWordsWrite},
	{Method: "DELETE", Path: "/brands/:id/aliases/:aliasId", Permission: PermExclusionWordsWrite},
	{Method: "POST", Path: "/search", Permission: PermSearchExecute},
	{Method: "GET", Path: "/logs/errors", Permission: PermSchedulesRead},
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
