package shared

import (
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const BrandCitationProximityMaxDistance = 250

// SearchURLInput is the minimal URL metadata needed for proximity attribution.
type SearchURLInput struct {
	URL           string
	CitationIndex int
}

// BrandCitationHit links a brand mention to a nearby cited domain.
type BrandCitationHit struct {
	Domain string
	URL    string
}

// BrandMentionMatcher finds brand mention positions using pre-compiled patterns.
type BrandMentionMatcher struct {
	caseInsensitive *regexp.Regexp
	caseSensitive   []*regexp.Regexp
}

// BrandCitationAttributor attributes brand mentions to nearby citations in response text.
type BrandCitationAttributor struct {
	matcher *BrandMentionMatcher
}

type citationAnchor struct {
	position int
	domain   string
	url      string
}

type citationURLIndex struct {
	byCitationIndex map[int]SearchURLInput
	sources         []SearchURLInput
}

type responseCitationData struct {
	text       string
	anchors    []citationAnchor
	lineStarts []int
	urlIndex   citationURLIndex
}

var (
	numericCitationRe = regexp.MustCompile(`\[(\d+)\]`)
	markdownLinkRe    = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

// NewBrandMentionMatcher builds a reusable matcher for a brand's search terms.
func NewBrandMentionMatcher(terms []BrandSearchTerm) *BrandMentionMatcher {
	insensitivePatterns := make([]string, 0, len(terms))
	sensitivePatterns := make([]*regexp.Regexp, 0)

	for _, term := range terms {
		trimmed := strings.TrimSpace(term.Term)
		if trimmed == "" {
			continue
		}

		pattern := regexp.QuoteMeta(trimmed)
		if term.CaseSensitive {
			sensitivePatterns = append(sensitivePatterns, regexp.MustCompile(`\b`+pattern+`\b`))
			continue
		}
		insensitivePatterns = append(insensitivePatterns, pattern)
	}

	matcher := &BrandMentionMatcher{
		caseSensitive: sensitivePatterns,
	}
	if len(insensitivePatterns) > 0 {
		matcher.caseInsensitive = regexp.MustCompile(`(?i)\b(?:` + strings.Join(insensitivePatterns, "|") + `)\b`)
	}

	return matcher
}

// NewBrandCitationAttributor creates a reusable attributor for a brand's search terms.
func NewBrandCitationAttributor(terms []BrandSearchTerm) *BrandCitationAttributor {
	return &BrandCitationAttributor{
		matcher: NewBrandMentionMatcher(terms),
	}
}

// Attribute maps brand mentions in text to the nearest inline citation.
func (a *BrandCitationAttributor) Attribute(text string, searchURLs []SearchURLInput) []BrandCitationHit {
	if a == nil || a.matcher == nil || text == "" || len(searchURLs) == 0 {
		return nil
	}

	data := buildResponseCitationData(text, searchURLs)
	if data == nil || len(data.anchors) == 0 {
		return nil
	}

	mentionPositions := a.matcher.FindPositions(text)
	if len(mentionPositions) == 0 {
		return nil
	}

	hits := make([]BrandCitationHit, 0, len(mentionPositions))
	for _, pos := range mentionPositions {
		anchor := nearestCitationAnchor(data, pos)
		if anchor == nil {
			continue
		}
		hits = append(hits, BrandCitationHit{
			Domain: anchor.domain,
			URL:    anchor.url,
		})
	}

	return hits
}

// AttributeBrandCitationDomains maps brand mentions in text to the nearest inline citation.
func AttributeBrandCitationDomains(text string, searchURLs []SearchURLInput, terms []BrandSearchTerm) []BrandCitationHit {
	return NewBrandCitationAttributor(terms).Attribute(text, searchURLs)
}

func (m *BrandMentionMatcher) FindPositions(text string) []int {
	if m == nil {
		return nil
	}

	excludedRanges := findExcludedMentionRanges(text)
	seen := make(map[int]struct{})
	positions := make([]int, 0)

	collectMatches := func(re *regexp.Regexp) {
		if re == nil {
			return
		}
		for _, match := range re.FindAllStringIndex(text, -1) {
			start := match[0]
			if isPositionExcluded(start, excludedRanges) {
				continue
			}
			if _, exists := seen[start]; exists {
				continue
			}
			seen[start] = struct{}{}
			positions = append(positions, start)
		}
	}

	collectMatches(m.caseInsensitive)
	for _, re := range m.caseSensitive {
		collectMatches(re)
	}

	if len(positions) > 1 {
		sort.Ints(positions)
	}

	return positions
}

func findExcludedMentionRanges(text string) [][2]int {
	ranges := make([][2]int, 0)

	for _, match := range markdownLinkRe.FindAllStringSubmatchIndex(text, -1) {
		if len(match) < 6 {
			continue
		}
		ranges = append(ranges, [2]int{match[4], match[5]})
	}

	for _, match := range numericCitationRe.FindAllStringSubmatchIndex(text, -1) {
		if len(match) < 2 {
			continue
		}
		ranges = append(ranges, [2]int{match[0], match[1]})
	}

	return mergeIntRanges(ranges)
}

func mergeIntRanges(ranges [][2]int) [][2]int {
	if len(ranges) == 0 {
		return nil
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i][0] < ranges[j][0]
	})

	merged := make([][2]int, 0, len(ranges))
	current := ranges[0]
	for i := 1; i < len(ranges); i++ {
		next := ranges[i]
		if next[0] <= current[1] {
			if next[1] > current[1] {
				current[1] = next[1]
			}
			continue
		}
		merged = append(merged, current)
		current = next
	}
	merged = append(merged, current)
	return merged
}

func isPositionExcluded(position int, ranges [][2]int) bool {
	for _, r := range ranges {
		if position >= r[0] && position < r[1] {
			return true
		}
	}
	return false
}

func buildResponseCitationData(text string, searchURLs []SearchURLInput) *responseCitationData {
	urlIndex := buildCitationURLIndex(searchURLs)
	anchors := findCitationAnchors(text, urlIndex)
	if len(anchors) == 0 {
		return nil
	}

	return &responseCitationData{
		text:       text,
		anchors:    anchors,
		lineStarts: computeLineStarts(text),
		urlIndex:   urlIndex,
	}
}

func buildCitationURLIndex(searchURLs []SearchURLInput) citationURLIndex {
	index := citationURLIndex{
		byCitationIndex: make(map[int]SearchURLInput, len(searchURLs)),
		sources:         searchURLs,
	}
	for _, source := range searchURLs {
		index.byCitationIndex[source.CitationIndex] = source
	}
	return index
}

func findCitationAnchors(text string, urlIndex citationURLIndex) []citationAnchor {
	seen := make(map[int]struct{})
	anchors := make([]citationAnchor, 0, len(urlIndex.sources)+4)

	for _, match := range numericCitationRe.FindAllStringSubmatchIndex(text, -1) {
		if len(match) < 4 {
			continue
		}
		position := match[0]
		if _, exists := seen[position]; exists {
			continue
		}

		rawIndex := text[match[2]:match[3]]
		index, err := strconv.Atoi(rawIndex)
		if err != nil || index < 1 {
			continue
		}

		domain, resolvedURL := resolveCitationFromIndex(index, urlIndex)
		if domain == "" {
			continue
		}

		seen[position] = struct{}{}
		anchors = append(anchors, citationAnchor{
			position: position,
			domain:   domain,
			url:      resolvedURL,
		})
	}

	for _, match := range markdownLinkRe.FindAllStringSubmatchIndex(text, -1) {
		if len(match) < 6 {
			continue
		}
		position := match[0]
		if _, exists := seen[position]; exists {
			continue
		}

		rawURL := strings.TrimSpace(text[match[4]:match[5]])
		domain := extractHostDomain(rawURL)
		if domain == "" {
			continue
		}

		seen[position] = struct{}{}
		anchors = append(anchors, citationAnchor{
			position: position,
			domain:   domain,
			url:      rawURL,
		})
	}

	if len(anchors) > 1 {
		sort.Slice(anchors, func(i, j int) bool {
			return anchors[i].position < anchors[j].position
		})
	}

	return anchors
}

func resolveCitationFromIndex(index int, urlIndex citationURLIndex) (string, string) {
	zeroBased := index - 1

	if source, ok := urlIndex.byCitationIndex[zeroBased]; ok && source.URL != "" {
		return extractHostDomain(source.URL), source.URL
	}
	if source, ok := urlIndex.byCitationIndex[index]; ok && source.URL != "" {
		return extractHostDomain(source.URL), source.URL
	}
	if zeroBased >= 0 && zeroBased < len(urlIndex.sources) && urlIndex.sources[zeroBased].URL != "" {
		source := urlIndex.sources[zeroBased]
		return extractHostDomain(source.URL), source.URL
	}

	return "", ""
}

func nearestCitationAnchor(data *responseCitationData, mentionPos int) *citationAnchor {
	lineStart, lineEnd := lineBoundsAtCached(data.lineStarts, len(data.text), mentionPos)

	if anchor := nearestAnchorInRange(data.anchors, mentionPos, lineStart, lineEnd, false); anchor != nil {
		return anchor
	}

	return nearestAnchorInRange(data.anchors, mentionPos, 0, len(data.text), true)
}

func nearestAnchorInRange(
	anchors []citationAnchor,
	mentionPos, rangeStart, rangeEnd int,
	enforceMaxDistance bool,
) *citationAnchor {
	var best *citationAnchor
	bestDistance := BrandCitationProximityMaxDistance + 1

	for i := range anchors {
		anchor := &anchors[i]
		if anchor.position < rangeStart || anchor.position > rangeEnd {
			continue
		}

		distance := absInt(anchor.position - mentionPos)
		if enforceMaxDistance && distance > BrandCitationProximityMaxDistance {
			continue
		}

		if distance < bestDistance || (distance == bestDistance && preferAnchor(anchor, best, mentionPos)) {
			best = anchor
			bestDistance = distance
		}
	}

	if best == nil {
		return nil
	}
	if enforceMaxDistance && bestDistance > BrandCitationProximityMaxDistance {
		return nil
	}

	return best
}

func preferAnchor(candidate, current *citationAnchor, mentionPos int) bool {
	if current == nil {
		return true
	}

	candidateAfter := candidate.position >= mentionPos
	currentAfter := current.position >= mentionPos
	if candidateAfter != currentAfter {
		return candidateAfter
	}

	return candidate.position < current.position
}

func computeLineStarts(text string) []int {
	lineStarts := []int{0}
	for index := 0; index < len(text); index++ {
		if text[index] == '\n' {
			lineStarts = append(lineStarts, index+1)
		}
	}
	return lineStarts
}

func lineBoundsAtCached(lineStarts []int, textLen, position int) (int, int) {
	if position < 0 {
		position = 0
	}
	if position > textLen {
		position = textLen
	}

	lineStart := 0
	for _, start := range lineStarts {
		if start > position {
			break
		}
		lineStart = start
	}

	lineEnd := textLen
	for _, start := range lineStarts {
		if start > position {
			lineEnd = start - 1
			break
		}
	}

	return lineStart, lineEnd
}

func extractHostDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	host := strings.ToLower(parsed.Host)
	if host == "" {
		return ""
	}
	return strings.TrimPrefix(host, "www.")
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
