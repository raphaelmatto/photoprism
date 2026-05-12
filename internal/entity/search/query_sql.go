package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// querySQL holds a SQL fragment and its bind arguments.
type querySQL struct {
	sql  string
	args []interface{}
}

// likeEscaper rewrites the LIKE wildcards and escape character in a phrase so
// substring matching only triggers on the literal user input.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// normalizeTerm prepares user input for case-insensitive comparisons. The
// result is sanitized for SQL string interpolation by clean.SqlString.
func normalizeTerm(s string) string {
	return strings.ToLower(clean.SqlString(s))
}

// wordClause matches a bare word against the keyword index using a prefix-LIKE
// and also against the photo title and caption with a prefix wildcard. Words
// under four characters require an exact keyword match so the index stays
// useful and short queries do not return the entire library; title and
// caption are always prefix-matched so single letters still find titles like
// "Neckarbrücke" without a separate description fallback.
func wordClause(word string) querySQL {
	w := normalizeTerm(word)
	if w == "" {
		return querySQL{}
	}

	escaped := likeEscaper.Replace(w)
	keywordPattern := escaped
	if len([]rune(w)) >= 4 {
		keywordPattern = keywordPattern + "%"
	}
	titlePattern := escaped + "%"

	return querySQL{
		sql: "(photos.id IN (SELECT pk.photo_id FROM photos_keywords pk JOIN keywords k ON pk.keyword_id = k.id" +
			" WHERE LOWER(k.keyword) LIKE ?)" +
			" OR LOWER(photos.photo_title) LIKE ?" +
			" OR LOWER(photos.photo_caption) LIKE ?)",
		args: []interface{}{keywordPattern, titlePattern, titlePattern},
	}
}

// phraseClause matches a quoted phrase. It checks for an exact keyword row and
// a case-insensitive substring in the photo title and caption.
func phraseClause(phrase string) querySQL {
	p := normalizeTerm(phrase)
	if p == "" {
		return querySQL{}
	}

	substr := "%" + likeEscaper.Replace(p) + "%"

	return querySQL{
		sql: "(photos.id IN (SELECT pk.photo_id FROM photos_keywords pk JOIN keywords k ON pk.keyword_id = k.id" +
			" WHERE LOWER(k.keyword) = ?)" +
			" OR LOWER(photos.photo_title) LIKE ?" +
			" OR LOWER(photos.photo_caption) LIKE ?)",
		args: []interface{}{p, substr, substr},
	}
}

// joinClauses combines child clauses with the given boolean operator,
// dropping empty children.
func joinClauses(children []queryNode, op string) querySQL {
	var parts []string
	var args []interface{}

	for _, c := range children {
		cl := buildQuerySQL(c)
		if cl.sql == "" {
			continue
		}
		parts = append(parts, cl.sql)
		args = append(args, cl.args...)
	}

	switch len(parts) {
	case 0:
		return querySQL{}
	case 1:
		return querySQL{sql: parts[0], args: args}
	default:
		return querySQL{sql: "(" + strings.Join(parts, fmt.Sprintf(" %s ", op)) + ")", args: args}
	}
}

// buildQuerySQL walks the AST and returns a single SQL fragment plus bind
// arguments. The returned clause is suitable for gorm.DB.Where.
func buildQuerySQL(node queryNode) querySQL {
	switch n := node.(type) {
	case nil:
		return querySQL{}
	case queryTerm:
		if n.Phrase {
			return phraseClause(n.Text)
		}
		return wordClause(n.Text)
	case queryAnd:
		return joinClauses(n.Children, "AND")
	case queryOr:
		return joinClauses(n.Children, "OR")
	}

	return querySQL{}
}

// collectWords walks the AST and returns the deduplicated set of bare word
// terminals. Phrases are skipped because they are intended as literal matches.
func collectWords(node queryNode) []string {
	seen := map[string]struct{}{}
	var out []string

	var walk func(queryNode)
	walk = func(n queryNode) {
		switch v := n.(type) {
		case queryTerm:
			if v.Phrase {
				return
			}
			w := strings.ToLower(strings.TrimSpace(v.Text))
			if w == "" {
				return
			}
			if _, ok := seen[w]; ok {
				return
			}
			seen[w] = struct{}{}
			out = append(out, w)
		case queryAnd:
			for _, c := range v.Children {
				walk(c)
			}
		case queryOr:
			for _, c := range v.Children {
				walk(c)
			}
		}
	}
	walk(node)
	return out
}
