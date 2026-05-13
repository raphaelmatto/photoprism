package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchKeywords finds and returns keywords as JSON.
//
//	@Summary	finds and returns keywords as JSON
//	@Id			SearchKeywords
//	@Tags		Keywords
//	@Produce	json
//	@Success	200				{array}	search.KeywordResult
//	@Failure	401,403			{object}	i18n.Response
//	@Router		/api/v1/keywords [get]
func SearchKeywords(router *gin.RouterGroup) {
	router.GET("/keywords", func(c *gin.Context) {
		s := Auth(c, acl.ResourceKeywords, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		result, err := search.Keywords()

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, result)
	})
}
