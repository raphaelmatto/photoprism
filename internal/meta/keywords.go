package meta

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media/projection"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Built-in keyword slugs inferred from metadata.
const (
	KeywordFlash           = "flash"
	KeywordHdr             = "hdr"
	KeywordBurst           = "burst"
	KeywordPanorama        = "panorama"
	KeywordEquirectangular = string(projection.Equirectangular)
)

// Keywords represents a list of metadata keywords.
type Keywords []string

// String returns a string containing all keywords.
func (w Keywords) String() string {
	return strings.Join(w, ", ")
}

// AutoKeywords lists keywords we automatically infer from descriptions or EXIF flags.
var AutoKeywords = []string{KeywordHdr, KeywordBurst, KeywordPanorama, KeywordEquirectangular}

// autoKeywordsEnabled reports whether automatically generated indexing keywords are enabled.
func autoKeywordsEnabled() bool {
	return entity.AddAIKeywordsEnabled()
}

// AddKeywords appends keywords, preserving multi-word entries like "Dan Nixon".
// Input is split on commas and semicolons; each segment is stored as a whole keyword token.
func (data *Data) AddKeywords(w string) {
	w = SanitizeMeta(w)

	if len(w) < 1 {
		return
	}

	for _, kw := range strings.FieldsFunc(w, func(r rune) bool { return r == ',' || r == ';' }) {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		data.Keywords = txt.UniqueWordsPreservingCase(append(data.Keywords, kw))
	}
}

// AutoAddKeywords automatically appends relevant keywords from a string (e.g. description).
func (data *Data) AutoAddKeywords(s string) {
	s = strings.ToLower(SanitizeMeta(s))

	if len(s) < 1 {
		return
	}

	for _, w := range AutoKeywords {
		if strings.Contains(s, w) {
			data.AddKeywords(w)
			if w == KeywordHdr {
				data.ImageType = ImageTypeHDR
			}
		}
	}
}
