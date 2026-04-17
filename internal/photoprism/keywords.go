package photoprism

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/txt"
)

// rebuildDetailsKeywords recomputes persisted photo keywords from trusted metadata and generated terms.
func rebuildDetailsKeywords(details *entity.Details, explicitKeywords, generatedKeywords []string, metadataKeywordsLoaded bool) {
	if details == nil {
		return
	}

	current := txt.UniqueWordsPreservingCase(txt.Words(details.Keywords))
	generated := txt.UniqueWords(generatedKeywords)

	switch details.KeywordsSrc {
	case string(entity.SrcManual), string(entity.SrcBatch):
		if !entity.AddAIKeywordsEnabled() && len(generated) > 0 {
			current = txt.RemoveFromWordsPreservingCase(current, strings.Join(generated, " "))
		}

		details.Keywords = strings.Join(txt.UniqueWordsPreservingCase(current), ", ")
		return
	}

	if metadataKeywordsLoaded {
		current = txt.UniqueWordsPreservingCase(explicitKeywords)
	} else if !entity.AddAIKeywordsEnabled() && len(generated) > 0 {
		current = txt.RemoveFromWordsPreservingCase(current, strings.Join(generated, " "))
	}

	if entity.AddAIKeywordsEnabled() {
		current = append(current, generated...)
	}

	details.Keywords = strings.Join(txt.UniqueWordsPreservingCase(current), ", ")
}
