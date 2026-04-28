package search

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// Keywords returns unique IPTC/XMP keywords from photo details, ordered alphabetically.
// It reads from the details table rather than the keywords full-text search index,
// which also contains words extracted from titles, captions, and AI labels.
func Keywords() (results []entity.Keyword, err error) {
	var rows []struct{ Keywords string }

	if result := UnscopedDb().Table("details").
		Select("DISTINCT keywords").
		Where("keywords <> ''").
		Scan(&rows); result.Error != nil {
		return results, result.Error
	}

	seen := make(map[string]struct{}, len(rows))

	for _, row := range rows {
		for _, kw := range strings.Split(row.Keywords, ",") {
			kw = strings.TrimSpace(kw)
			if kw != "" {
				seen[kw] = struct{}{}
			}
		}
	}

	for kw := range seen {
		results = append(results, entity.Keyword{Keyword: kw})
	}

	sort.Slice(results, func(i, j int) bool {
		return strings.ToLower(results[i].Keyword) < strings.ToLower(results[j].Keyword)
	})

	return results, nil
}
