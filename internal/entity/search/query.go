package search

import (
	"strings"
	"unicode"
)

// Photo search query syntax.
//
// Terminals:
//
//	WORD      bare alphanumeric token, e.g. family
//	          matches keywords with a prefix-LIKE: keyword LIKE 'word%'.
//	PHRASE    quoted multi-word token, e.g. "Family reunion"
//	          matches an exact keyword row OR a substring in
//	          photos.photo_title / photos.photo_caption / photos.photo_description.
//
// Operators:
//
//	&         AND  (higher precedence)
//	|         OR   (lower precedence)
//	()        explicit grouping
//	<space>   implicit OR  (same precedence as |)
//
// Grammar:
//
//	expr   = orExpr
//	orExpr = andExpr ( ('|' | ws) andExpr )*
//	andExpr= unary  ( '&' unary )*
//	unary  = '(' expr ')' | term
//	term   = QUOTED | WORD
//
// The parser is forgiving: dangling/leading operators are dropped, unmatched
// parentheses fall through as if implicit groups, and empty input yields nil.

// queryNode is the AST root for a parsed search query.
type queryNode interface {
	isQueryNode()
}

// queryTerm is a single matchable token. Phrase distinguishes quoted phrases
// from bare words so the SQL builder can apply different matching rules.
type queryTerm struct {
	Text   string
	Phrase bool
}

// queryAnd represents an AND-joined sequence of children.
type queryAnd struct {
	Children []queryNode
}

// queryOr represents an OR-joined sequence of children.
type queryOr struct {
	Children []queryNode
}

func (queryTerm) isQueryNode() {}
func (queryAnd) isQueryNode()  {}
func (queryOr) isQueryNode()   {}

// queryToken is a single lexer output.
type queryToken struct {
	kind  queryTokenKind
	value string
}

type queryTokenKind int

const (
	tokWord queryTokenKind = iota
	tokPhrase
	tokAnd
	tokOr
	tokLParen
	tokRParen
)

// tokenizeQuery splits a normalized query string into tokens. Whitespace
// between two value tokens becomes an implicit OR token.
func tokenizeQuery(s string) []queryToken {
	var tokens []queryToken
	runes := []rune(s)
	i := 0

	flushImplicitOr := func() {
		if len(tokens) == 0 {
			return
		}
		switch tokens[len(tokens)-1].kind {
		case tokWord, tokPhrase, tokRParen:
			tokens = append(tokens, queryToken{kind: tokOr})
		}
	}

	for i < len(runes) {
		r := runes[i]

		switch {
		case unicode.IsSpace(r):
			// Collapse runs of whitespace; insert implicit OR if it sits
			// between two value tokens.
			j := i
			for j < len(runes) && unicode.IsSpace(runes[j]) {
				j++
			}
			i = j
			if i >= len(runes) {
				continue
			}
			// Only emit implicit OR if the next token is also a value.
			next := runes[i]
			if next != '&' && next != '|' && next != ')' {
				flushImplicitOr()
			}
		case r == '&':
			tokens = append(tokens, queryToken{kind: tokAnd})
			i++
		case r == '|':
			tokens = append(tokens, queryToken{kind: tokOr})
			i++
		case r == '(':
			tokens = append(tokens, queryToken{kind: tokLParen})
			i++
		case r == ')':
			tokens = append(tokens, queryToken{kind: tokRParen})
			i++
		case r == '"':
			j := i + 1
			for j < len(runes) && runes[j] != '"' {
				j++
			}
			phrase := strings.TrimSpace(string(runes[i+1 : j]))
			if phrase != "" {
				tokens = append(tokens, queryToken{kind: tokPhrase, value: phrase})
			}
			if j < len(runes) {
				j++ // consume closing quote
			}
			i = j
		default:
			j := i
			for j < len(runes) {
				c := runes[j]
				if unicode.IsSpace(c) || c == '&' || c == '|' || c == '(' || c == ')' || c == '"' {
					break
				}
				j++
			}
			word := strings.TrimSpace(string(runes[i:j]))
			if word != "" {
				tokens = append(tokens, queryToken{kind: tokWord, value: word})
			}
			i = j
		}
	}

	return tokens
}

// queryParser is a recursive-descent parser for the search query grammar.
type queryParser struct {
	tokens []queryToken
	pos    int
}

func (p *queryParser) peek() (queryToken, bool) {
	if p.pos >= len(p.tokens) {
		return queryToken{}, false
	}
	return p.tokens[p.pos], true
}

func (p *queryParser) consume() (queryToken, bool) {
	t, ok := p.peek()
	if ok {
		p.pos++
	}
	return t, ok
}

// parseExpr parses a top-level expression.
func (p *queryParser) parseExpr() queryNode {
	return p.parseOr()
}

// parseOr parses an OR-chain of AND expressions.
func (p *queryParser) parseOr() queryNode {
	var children []queryNode

	if first := p.parseAnd(); first != nil {
		children = append(children, first)
	}

	for {
		t, ok := p.peek()
		if !ok || t.kind != tokOr {
			break
		}
		p.consume()
		next := p.parseAnd()
		if next != nil {
			children = append(children, next)
		}
	}

	return wrapOr(children)
}

// parseAnd parses an AND-chain of unary terms.
func (p *queryParser) parseAnd() queryNode {
	var children []queryNode

	if first := p.parseUnary(); first != nil {
		children = append(children, first)
	}

	for {
		t, ok := p.peek()
		if !ok || t.kind != tokAnd {
			break
		}
		p.consume()
		next := p.parseUnary()
		if next != nil {
			children = append(children, next)
		}
	}

	return wrapAnd(children)
}

// parseUnary parses a single term or a parenthesized expression. Leading
// operators are discarded so malformed queries degrade gracefully.
func (p *queryParser) parseUnary() queryNode {
	for {
		t, ok := p.peek()
		if !ok {
			return nil
		}

		switch t.kind {
		case tokAnd, tokOr:
			// Drop dangling operator.
			p.consume()
			continue
		case tokLParen:
			p.consume()
			inner := p.parseExpr()
			if next, ok := p.peek(); ok && next.kind == tokRParen {
				p.consume()
			}
			if inner == nil {
				continue
			}
			return inner
		case tokRParen:
			// Unmatched close paren; stop here so the caller can recover.
			return nil
		case tokWord:
			p.consume()
			return queryTerm{Text: t.value, Phrase: false}
		case tokPhrase:
			p.consume()
			return queryTerm{Text: t.value, Phrase: true}
		}
	}
}

// wrapAnd returns a node that ANDs the given children, simplifying degenerate
// 0/1-child cases so downstream SQL generation is easier.
func wrapAnd(children []queryNode) queryNode {
	switch len(children) {
	case 0:
		return nil
	case 1:
		return children[0]
	default:
		return queryAnd{Children: children}
	}
}

// wrapOr returns a node that ORs the given children, simplifying degenerate
// 0/1-child cases.
func wrapOr(children []queryNode) queryNode {
	switch len(children) {
	case 0:
		return nil
	case 1:
		return children[0]
	default:
		return queryOr{Children: children}
	}
}

// parseQuery turns a normalized search query into an AST. It returns nil when
// the query contains no usable tokens.
func parseQuery(s string) queryNode {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	p := &queryParser{tokens: tokenizeQuery(s)}
	return p.parseExpr()
}
