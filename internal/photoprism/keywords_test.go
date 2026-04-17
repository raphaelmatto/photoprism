package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestRebuildDetailsKeywords(t *testing.T) {
	t.Run("RemovesGeneratedKeywordsFromManualDetailsWhenDisabled", func(t *testing.T) {
		entity.SetAddAIKeywords(false)
		t.Cleanup(func() {
			entity.SetAddAIKeywords(true)
		})

		details := &entity.Details{
			Keywords:    "road, thornton, wedding",
			KeywordsSrc: string(entity.SrcManual),
		}

		rebuildDetailsKeywords(details, nil, []string{"road", "thornton"}, false)

		assert.Equal(t, "wedding", details.Keywords)
	})

	t.Run("RebuildsMetadataKeywordsWithoutGeneratedTermsWhenDisabled", func(t *testing.T) {
		entity.SetAddAIKeywords(false)
		t.Cleanup(func() {
			entity.SetAddAIKeywords(true)
		})

		details := &entity.Details{
			Keywords:    "grey, road, thornton",
			KeywordsSrc: string(entity.SrcMeta),
		}

		rebuildDetailsKeywords(details, []string{"Wedding", "Raphe"}, []string{"road", "thornton", "grey"}, true)

		assert.Equal(t, "Raphe, Wedding", details.Keywords)
	})

	t.Run("AppendsGeneratedKeywordsWhenEnabled", func(t *testing.T) {
		entity.SetAddAIKeywords(true)
		t.Cleanup(func() {
			entity.SetAddAIKeywords(true)
		})

		details := &entity.Details{
			Keywords:    "Wedding",
			KeywordsSrc: string(entity.SrcMeta),
		}

		rebuildDetailsKeywords(details, []string{"Wedding"}, []string{"road", "thornton"}, true)

		assert.Equal(t, "road, thornton, Wedding", details.Keywords)
	})
}
