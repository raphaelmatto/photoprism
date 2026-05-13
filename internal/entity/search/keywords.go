package search

import (
	"sort"
	"strings"
)

// KeywordResult is the API-friendly view of a keyword in the Keywords page
// listing, capturing both the keyword string and the number of photos it
// appears on. It keeps the response self-contained instead of leaking the
// entity.Keyword DB model and its other fields onto the wire.
type KeywordResult struct {
	Keyword string `json:"Keyword"`
	Count   int    `json:"Count"`
}

// Keywords returns unique IPTC/XMP keywords from photo details together with
// the number of photos each one appears on, ordered alphabetically. It reads
// from the details table rather than the keywords full-text search index,
// which also contains words extracted from titles, captions, and AI labels.
func Keywords() (results []KeywordResult, err error) {
	var rows []struct{ Keywords string }

	if result := UnscopedDb().Table("details").
		Select("keywords").
		Where("keywords <> ''").
		Scan(&rows); result.Error != nil {
		return results, result.Error
	}

	counts := make(map[string]int, len(rows))

	for _, row := range rows {
		seen := make(map[string]bool)
		for _, kw := range strings.Split(row.Keywords, ",") {
			kw = strings.TrimSpace(kw)
			if kw == "" || seen[kw] {
				continue
			}
			counts[kw]++
			seen[kw] = true
		}
	}

	results = make([]KeywordResult, 0, len(counts))
	for kw, c := range counts {
		results = append(results, KeywordResult{Keyword: kw, Count: c})
	}

	sort.Slice(results, func(i, j int) bool {
		return strings.ToLower(results[i].Keyword) < strings.ToLower(results[j].Keyword)
	})

	return results, nil
}
