package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestGetSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetSettings(router)

		r := PerformRequest(app, "GET", "/api/v1/settings")
		val := gjson.Get(r.Body.String(), "ui.theme")
		assert.NotEmpty(t, val.String())
		val2 := gjson.Get(r.Body.String(), "ui.language")
		assert.NotEmpty(t, val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestSaveSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetSettings(router)

		r := PerformRequest(app, "GET", "/api/v1/settings")
		val := gjson.Get(r.Body.String(), "ui.language")
		assert.Equal(t, "en", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
		SaveSettings(router)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language": "de"}}`)
		assert.Equal(t, http.StatusOK, r2.Code)
		r4 := PerformRequest(app, "GET", "/api/v1/settings")
		val2 := gjson.Get(r4.Body.String(), "ui.language")
		assert.Equal(t, "de", val2.String())
		r3 := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language": "en"}}`)
		assert.Equal(t, http.StatusOK, r3.Code)
	})
	t.Run("MetadataLayout", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetSettings(router)
		SaveSettings(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"display":{"metadata":{"cards":["date","caption","keywords"],"list":["filename","date"],"lightbox":["date","caption","fileInfo"]}}}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "date", gjson.Get(r.Body.String(), "display.metadata.cards.0").String())
		assert.Equal(t, "filename", gjson.Get(r.Body.String(), "display.metadata.list.0").String())
		assert.Equal(t, "fileInfo", gjson.Get(r.Body.String(), "display.metadata.lightbox.2").String())

		r2 := PerformRequest(app, "GET", "/api/v1/settings")
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.Equal(t, "date", gjson.Get(r2.Body.String(), "display.metadata.cards.0").String())
		assert.Equal(t, "filename", gjson.Get(r2.Body.String(), "display.metadata.list.0").String())
		assert.Equal(t, "fileInfo", gjson.Get(r2.Body.String(), "display.metadata.lightbox.2").String())
	})
	t.Run("LightboxBorder", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetSettings(router)
		SaveSettings(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"display":{"lightboxBorder":2.5}}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, 2.5, gjson.Get(r.Body.String(), "display.lightboxBorder").Float())

		r2 := PerformRequest(app, "GET", "/api/v1/settings")
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.Equal(t, 2.5, gjson.Get(r2.Body.String(), "display.lightboxBorder").Float())
	})
	t.Run("AddAIKeywords", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetSettings(router)
		SaveSettings(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"index":{"addAIKeywords":false}}`)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.False(t, gjson.Get(r.Body.String(), "index.addAIKeywords").Bool())

		r2 := PerformRequest(app, "GET", "/api/v1/settings")
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.False(t, gjson.Get(r2.Body.String(), "index.addAIKeywords").Bool())
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()

		SaveSettings(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language":123}}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
