package search

import "github.com/photoprism/photoprism/internal/entity"

// Keywords returns all active keywords ordered alphabetically.
func Keywords() (results []entity.Keyword, err error) {
	s := UnscopedDb().Table("keywords").
		Select("id, keyword, skip").
		Where("skip = 0").
		Where("keyword <> ''").
		Order("keyword ASC")

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
