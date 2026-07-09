package mongodb

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDecodeResponseFromDoc_SearchURLs(t *testing.T) {
	doc := bson.M{
		"_id":           "resp-1",
		"prompt_id":     "prompt-1",
		"response_text": "example",
		"search_urls": primitive.A{
			primitive.M{
				"search_query":   "fashion brands",
				"url":            "https://www.example.com/article",
				"title":          "Example Article",
				"citation_index": int32(1),
			},
		},
	}

	response := decodeResponseFromDoc(doc)
	if len(response.SearchURLs) != 1 {
		t.Fatalf("expected 1 search URL, got %d", len(response.SearchURLs))
	}

	url := response.SearchURLs[0]
	if url.URL != "https://www.example.com/article" {
		t.Fatalf("unexpected url: %q", url.URL)
	}
	if url.SearchQuery != "fashion brands" {
		t.Fatalf("unexpected search query: %q", url.SearchQuery)
	}
	if url.Title != "Example Article" {
		t.Fatalf("unexpected title: %q", url.Title)
	}
	if url.CitationIndex != 1 {
		t.Fatalf("unexpected citation index: %d", url.CitationIndex)
	}
}
