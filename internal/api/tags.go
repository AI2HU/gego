package api

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

func parseTagsFromQuery(c *gin.Context) []string {
	raw := c.QueryArray("tags")
	if len(raw) == 0 {
		if value := strings.TrimSpace(c.Query("tags")); value != "" {
			raw = strings.Split(value, ",")
		}
	}

	tags := make([]string, 0, len(raw))
	seen := make(map[string]bool, len(raw))
	for _, tag := range raw {
		tag = strings.TrimSpace(tag)
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
	}

	return tags
}

func (s *Server) resolvePromptIDsByTags(ctx context.Context, tags []string) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	prompts, err := s.promptService.GetPromptsByTags(ctx, tags)
	if err != nil {
		return nil, err
	}

	promptIDs := make([]string, 0, len(prompts))
	for _, prompt := range prompts {
		promptIDs = append(promptIDs, prompt.ID)
	}

	return promptIDs, nil
}
