package shared

import (
	"testing"
)

func TestAttributeBrandCitationDomains_numericMarkersSameLine(t *testing.T) {
	text := "For CRM tools, Salesforce [2] and HubSpot [1] are popular choices."
	urls := []SearchURLInput{
		{URL: "https://www.hubspot.com/crm", CitationIndex: 0},
		{URL: "https://www.4degrees.com/best-crm", CitationIndex: 1},
	}
	terms := []BrandSearchTerm{{Term: "Salesforce", CaseSensitive: false}}

	hits := AttributeBrandCitationDomains(text, urls, terms)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Domain != "4degrees.com" {
		t.Fatalf("expected 4degrees.com, got %s", hits[0].Domain)
	}
}

func TestAttributeBrandCitationDomains_markdownLink(t *testing.T) {
	text := "Salesforce [The Best CRMs of 2025](https://4degrees.com/article) leads the market."
	terms := []BrandSearchTerm{{Term: "Salesforce", CaseSensitive: false}}

	hits := AttributeBrandCitationDomains(text, []SearchURLInput{{URL: "https://4degrees.com/article", CitationIndex: 0}}, terms)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Domain != "4degrees.com" {
		t.Fatalf("expected 4degrees.com, got %s", hits[0].Domain)
	}
}

func TestAttributeBrandCitationDomains_noNearbyCitation(t *testing.T) {
	text := "Salesforce is a major player in enterprise software."
	terms := []BrandSearchTerm{{Term: "Salesforce", CaseSensitive: false}}
	urls := []SearchURLInput{{URL: "https://example.com/unrelated", CitationIndex: 0}}

	hits := AttributeBrandCitationDomains(text, urls, terms)
	if len(hits) != 0 {
		t.Fatalf("expected no hits, got %d", len(hits))
	}
}

func TestBrandMentionMatcher_reusesCompiledPatterns(t *testing.T) {
	matcher := NewBrandMentionMatcher([]BrandSearchTerm{
		{Term: "Salesforce", CaseSensitive: false},
		{Term: "SFDC", CaseSensitive: true},
	})

	text := "Salesforce and SFDC are mentioned."
	positions := matcher.FindPositions(text)
	if len(positions) != 2 {
		t.Fatalf("expected 2 positions, got %d", len(positions))
	}
}

func TestBrandCitationAttributor_reusedAcrossResponses(t *testing.T) {
	attributor := NewBrandCitationAttributor([]BrandSearchTerm{{Term: "Nike", CaseSensitive: false}})

	first := attributor.Attribute(
		"Nike [1] leads the market.",
		[]SearchURLInput{{URL: "https://forbes.com/nike", CitationIndex: 0}},
	)
	second := attributor.Attribute(
		"Nike [The story](https://espn.com/nike) keeps growing.",
		[]SearchURLInput{{URL: "https://espn.com/nike", CitationIndex: 0}},
	)

	if len(first) != 1 || first[0].Domain != "forbes.com" {
		t.Fatalf("unexpected first hit: %+v", first)
	}
	if len(second) != 1 || second[0].Domain != "espn.com" {
		t.Fatalf("unexpected second hit: %+v", second)
	}
}

func TestBrandMentionMatcher_ignoresBrandInsideMarkdownURL(t *testing.T) {
	matcher := NewBrandMentionMatcher([]BrandSearchTerm{{Term: "Nike", CaseSensitive: false}})
	text := "Nike [The story](https://espn.com/nike) keeps growing."

	positions := matcher.FindPositions(text)
	if len(positions) != 1 {
		t.Fatalf("expected 1 visible mention, got %d (%v)", len(positions), positions)
	}
	if positions[0] != 0 {
		t.Fatalf("expected mention at start, got %d", positions[0])
	}
}
