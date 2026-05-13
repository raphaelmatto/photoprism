package search

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// formatNode renders a queryNode to a compact S-expression so test assertions
// can describe expected ASTs without referring to private struct types.
func formatNode(n queryNode) string {
	switch v := n.(type) {
	case nil:
		return "nil"
	case queryTerm:
		if v.Phrase {
			return "phrase(" + v.Text + ")"
		}
		return "word(" + v.Text + ")"
	case queryAnd:
		var parts []string
		for _, c := range v.Children {
			parts = append(parts, formatNode(c))
		}
		return "and(" + strings.Join(parts, ",") + ")"
	case queryOr:
		var parts []string
		for _, c := range v.Children {
			parts = append(parts, formatNode(c))
		}
		return "or(" + strings.Join(parts, ",") + ")"
	}
	return "?"
}

func TestParseQuery(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"Empty", "", "nil"},
		{"Whitespace", "   ", "nil"},
		{"SingleWord", "family", "word(family)"},
		{"BareTwoWords", "family reunion", "or(word(family),word(reunion))"},
		{"BareThreeWords", "family reunion raphe", "or(word(family),word(reunion),word(raphe))"},
		{"QuotedPhrase", `"Family reunion"`, "phrase(Family reunion)"},
		{"PhrasePlusWord", `"Family reunion" raphe`, "or(phrase(Family reunion),word(raphe))"},
		{"PhraseAndWord", `"Dan Nixon" & raphe`, "and(phrase(Dan Nixon),word(raphe))"},
		{"PhraseOrWord", `"Dan Nixon" | raphe`, "or(phrase(Dan Nixon),word(raphe))"},
		{"AndOverOr", "family & reunion | raphe", "or(and(word(family),word(reunion)),word(raphe))"},
		{"OrThenAnd", "raphe | family & reunion", "or(word(raphe),and(word(family),word(reunion)))"},
		{"ParensOverridePrecedence", "family & (reunion | raphe)", "and(word(family),or(word(reunion),word(raphe)))"},
		{"NestedParens", "((family))", "word(family)"},
		{"WhitespaceBetweenAndChunks", "a & b c", "or(and(word(a),word(b)),word(c))"},
		{"LeadingOperatorDropped", "& family", "word(family)"},
		{"TrailingOperatorDropped", "family &", "word(family)"},
		{"EmptyQuotesDropped", `"" raphe`, "word(raphe)"},
		{"UnmatchedOpenParen", "(family", "word(family)"},
		{"GluedOperators", "family&reunion|raphe", "or(and(word(family),word(reunion)),word(raphe))"},
		{"CaseIsPreservedInPhrase", `"Dan Nixon"`, "phrase(Dan Nixon)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatNode(parseQuery(tc.in))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBuildQuerySQL(t *testing.T) {
	t.Run("EmptyAst", func(t *testing.T) {
		assert.Empty(t, buildQuerySQL(nil).sql)
	})

	t.Run("SingleWord", func(t *testing.T) {
		c := buildQuerySQL(parseQuery("family"))
		assert.Contains(t, c.sql, "photos_keywords")
		assert.Contains(t, c.sql, "LOWER(k.keyword) LIKE ?")
		assert.Contains(t, c.sql, "LOWER(photos.photo_title) LIKE ?")
		assert.Equal(t, []interface{}{"family%", "family%", "family%"}, c.args)
	})

	t.Run("ShortWordExactMatch", func(t *testing.T) {
		c := buildQuerySQL(parseQuery("dog"))
		// Short words match the keyword index exactly but still prefix-match
		// titles and captions.
		assert.Equal(t, []interface{}{"dog", "dog%", "dog%"}, c.args)
	})

	t.Run("MultiWordPhrase", func(t *testing.T) {
		c := buildQuerySQL(parseQuery(`"Family reunion"`))
		// Multi-word phrases match the exact keyword tag plus a substring in
		// title and caption.
		assert.Contains(t, c.sql, "LOWER(k.keyword) = ?")
		assert.Contains(t, c.sql, "LOWER(photos.photo_title) LIKE ?")
		assert.Contains(t, c.sql, "LOWER(photos.photo_caption) LIKE ?")
		assert.Equal(t, []interface{}{"family reunion", "%family reunion%", "%family reunion%"}, c.args)
	})

	t.Run("SingleTokenPhraseExactTagOnly", func(t *testing.T) {
		c := buildQuerySQL(parseQuery(`"Ai"`))
		// Single-token phrases must not pull in title/caption substring matches:
		// "Ai" should not match captions containing "Gail" or "Paris".
		assert.Contains(t, c.sql, "LOWER(k.keyword) = ?")
		assert.NotContains(t, c.sql, "photo_title")
		assert.NotContains(t, c.sql, "photo_caption")
		assert.Equal(t, []interface{}{"ai"}, c.args)
	})

	t.Run("AndBindsTighterThanOr", func(t *testing.T) {
		c := buildQuerySQL(parseQuery("family & reunion | raphe"))
		// The outer combinator should be OR, not AND.
		assert.True(t, strings.HasPrefix(c.sql, "("))
		assert.Contains(t, c.sql, " OR ")
		// The AND group sits as a parenthesized sub-expression.
		assert.Contains(t, c.sql, " AND ")
	})

	t.Run("ParensFlipPrecedence", func(t *testing.T) {
		c := buildQuerySQL(parseQuery("family & (reunion | raphe)"))
		// Three word clauses, each contributing one keyword subquery.
		assert.Equal(t, 3, strings.Count(c.sql, "photos_keywords"))
		assert.Contains(t, c.sql, " AND ")
		assert.Contains(t, c.sql, " OR ")
	})

	t.Run("LikeWildcardsEscaped", func(t *testing.T) {
		c := buildQuerySQL(parseQuery(`"50% cotton"`))
		// The phrase contains a literal % that should be escaped.
		for _, a := range c.args {
			if s, ok := a.(string); ok {
				assert.NotContains(t, s, "%50%")
			}
		}
	})
}
