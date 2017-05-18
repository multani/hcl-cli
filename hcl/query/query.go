package query

import (
	"fmt"
	"strings"
)

// expr = ident
// expr = ident "." expr
// expr = ident "[" ident "]" "." expr

type Query interface{}

type KeyValue struct {
	Key   string
	Value string
	In    Query
}

type Obj struct {
	Name string
	In   Query
}

type Single struct {
	Name string
}

type Type string

const (
	LBRACK = "["
	RBRACK = "]"
	DOT    = "."
	IDENT  = "IDENT"
	EOF    = "EOF"
)

type Token struct {
	Type  Type
	Value string
	Start int
	End   int
}

func Lex(query string) Tokens {
	tokens := make([]Token, 0)

	acc := ""
	tokStart := 0

	for index, c := range query {
		if c == '[' {
			if acc != "" {
				tokens = append(tokens, Token{Type: IDENT, Value: acc, Start: tokStart, End: index})
				tokStart = index + 1
				acc = ""
			}
			tokens = append(tokens, Token{Type: LBRACK, Start: tokStart, End: index})
			tokStart = index + 1
		} else if c == ']' {
			if acc != "" {
				tokens = append(tokens, Token{Type: IDENT, Value: acc, Start: tokStart, End: index})
				tokStart = index + 1
				acc = ""
			}
			tokens = append(tokens, Token{Type: RBRACK, Start: tokStart, End: index})
			tokStart = index + 1
		} else if c == '.' {
			if acc != "" {
				tokens = append(tokens, Token{Type: IDENT, Value: acc, Start: tokStart, End: index})
				tokStart = index + 1
				acc = ""
			}
			tokens = append(tokens, Token{Type: DOT, Start: tokStart, End: index})
			tokStart = index + 1
		} else {
			acc += string(c)
			continue
		}
	}
	if acc != "" {
		tokens = append(tokens, Token{Type: IDENT, Value: acc, Start: tokStart, End: tokStart + len(acc)})
		acc = ""
	}

	tokens = append(tokens, Token{Type: EOF, Start: -1, End: -1})

	t := Tokens{
		query:  query,
		tokens: tokens,
		pos:    0,
	}
	return t
}

type Tokens struct {
	tokens []Token
	pos    int
	query  string
}

func (t *Tokens) match(typ Type) bool {
	n := t.pos + 1
	if n < len(t.tokens) && t.tokens[n].Type == typ {
		t.pos = n
		return true
	} else {
		return false
	}
}

func (t *Tokens) check(typ Type) bool {
	n := t.pos
	return n < len(t.tokens) && t.tokens[n].Type == typ
}

func (t *Tokens) advance() Token {
	if t.pos < len(t.tokens)-1 {
		t.pos++
	}
	return t.tokens[t.pos-1]
}

func (t *Tokens) isAtEnd() bool {
	return t.tokens[t.pos].Type == EOF
}

func (tokens *Tokens) Parse() Query {
	token := tokens.advance()
	lastToken := token

	if token.Type == IDENT {
		if tokens.isAtEnd() {
			return &Single{Name: token.Value}
		}

		if tokens.check(DOT) {
			tokens.advance()
			return &Obj{
				Name: token.Value,
				In:   tokens.Parse(),
			}
		}

		if tokens.check(LBRACK) {
			lastToken = tokens.advance()
			if tokens.check(IDENT) {
				value := tokens.advance()
				if tokens.check(RBRACK) {
					lastToken = tokens.advance()
					if tokens.check(DOT) {
						lastToken = tokens.advance()
						return &KeyValue{
							Key:   token.Value,
							Value: fmt.Sprintf(`"%s"`, value.Value),
							In:    tokens.Parse(),
						}
					} else {
						fmt.Println(tokens.query)
						fmt.Printf("%s^^\n", strings.Repeat(" ", lastToken.End))
						panic(fmt.Sprintf("parse error: expecting '.' at character %d", lastToken.End))
					}
				} else {
					panic("parse error: expecting ']'")
				}
			}
		}
	}

	panic(fmt.Sprintf("parse error 2: %s", token))
}
