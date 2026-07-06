package mongodb

import (
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/AI2HU/gego/internal/shared"
)

func applyKeywordFilter(query bson.M, keyword string) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return
	}

	applySearchTermsFilter(query, shared.ExpandBrandSearchTerms(keyword))
}

func applySearchTermsFilter(query bson.M, searchTerms []shared.BrandSearchTerm) {
	insensitivePatterns := make([]string, 0, len(searchTerms))
	caseSensitiveConditions := make([]bson.M, 0)

	for _, term := range searchTerms {
		trimmed := strings.TrimSpace(term.Term)
		if trimmed == "" {
			continue
		}
		pattern := regexp.QuoteMeta(trimmed)
		if term.CaseSensitive {
			caseSensitiveConditions = append(caseSensitiveConditions, bson.M{
				"response_text": bson.M{"$regex": pattern},
			})
			continue
		}
		insensitivePatterns = append(insensitivePatterns, pattern)
	}

	orConditions := make([]bson.M, 0, 1+len(caseSensitiveConditions))
	if len(insensitivePatterns) > 0 {
		regex := insensitivePatterns[0]
		if len(insensitivePatterns) > 1 {
			regex = strings.Join(insensitivePatterns, "|")
		}
		orConditions = append(orConditions, bson.M{
			"response_text": bson.M{"$regex": regex, "$options": "i"},
		})
	}
	orConditions = append(orConditions, caseSensitiveConditions...)

	switch len(orConditions) {
	case 0:
		return
	case 1:
		query["response_text"] = orConditions[0]["response_text"]
	default:
		query["$or"] = orConditions
	}
}
