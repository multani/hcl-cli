package main

import (
	"fmt"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/hashicorp/hcl/hcl/token"
	"os"
)

// update job.nomad "job[hydra].group[hydra].task[hydra].config.image" xxx

type Match interface {
	match()
}

type Multiple struct {
	Key   string
	Value string
	In    Match
}

type Obj struct {
	Name string
	In   Match
}

type Single struct {
	Name string
}

func (Multiple) match() {}
func (Obj) match()      {}
func (Single) match()   {}

type Walker struct {
	match Match
	value string
}

func newWalker(query string, value string) *Walker {
	return &Walker{
		match: parseQuery(query),
		value: value,
	}
}

func walkDebug(n ast.Node) (ast.Node, bool) {
	switch i := n.(type) {

	case *ast.Comment:
		fmt.Println("comment")
	case *ast.CommentGroup:
		fmt.Println("comment group")
	case *ast.File:
		fmt.Println("file")
	case *ast.ListType:
		fmt.Printf("list: %d items\n", len(i.List))
	case *ast.LiteralType:
		fmt.Printf("literal: %s (type=%v)\n", i.Token.Text, i.Token.Type)
	case *ast.ObjectItem:
		fmt.Printf("object item: %d keys\n", len(i.Keys))
	case *ast.ObjectKey:
		fmt.Printf("object key: %s (type=%v)\n", i.Token.Text, i.Token.Type)
	case *ast.ObjectList:
		fmt.Printf("object list: %d items\n", len(i.Items))
	case *ast.ObjectType:
		fmt.Printf("object type: %v\n", i.List)
	}
	return n, true
}

func (w Walker) walk(n ast.Node) (ast.Node, bool) {
	switch i := n.(type) {
	case *ast.ObjectItem:
		switch m := w.match.(type) {
		case *Multiple:
			if len(i.Keys) == 2 && i.Keys[0].Token.Text == m.Key && i.Keys[1].Token.Text == m.Value {
				w.match = m.In
				i.Val = ast.Walk(i.Val, w.walk)
				return i, false
			}

		case *Obj:
			if len(i.Keys) == 1 && i.Keys[0].Token.Text == m.Name {
				w.match = m.In
				i.Val = ast.Walk(i.Val, w.walk)
				return i, false
			}

		case *Single:
			if len(i.Keys) == 1 && i.Keys[0].Token.Text == m.Name {
				i.Val = &ast.LiteralType{
					Token: token.Token{
						Type: token.STRING,
						Text: fmt.Sprintf(`"%s"`, w.value),
					},
				}
				return i, false
			}
		}
	}
	return n, true
}

func parseQuery(query string) Match {

	v := make([]Match, 0)

	acc := ""
	key := ""
	for _, c := range query {
		if c == '[' {
			key = acc
			acc = ""
		} else if c == ']' {
			x := &Multiple{
				Key:   key,
				Value: fmt.Sprintf(`"%s"`, acc),
				In:    nil,
			}
			acc = ""
			key = ""
			v = append(v, x)
		} else if c == '.' {
			if acc != "" {
				x := &Obj{
					Name: acc,
					In:   nil,
				}
				acc = ""
				v = append(v, x)
			}
		} else {
			acc += string(c)
		}
	}

	x := &Single{
		Name: acc,
	}
	v = append(v, x)

	var m Match
	for i := len(v) - 1; i >= 0; i-- {
		e := v[i]
		switch j := e.(type) {
		case *Multiple:
			j.In = m
		case *Obj:
			j.In = m
		}
		m = e
	}

	return m
}

type Type int

const (
	LBRACK = 0
	RBRACK = 1
	DOT    = 2
	IDENT  = 3
)

type Token struct {
	Type  Type
	Value string
}

func lexer(query string) Tokens {
	tokens := make([]Token, 0)

	acc := ""

	for _, c := range query {
		if c == '[' {
			if acc != "" {
				tokens = append(tokens, Token{Type: IDENT, Value: acc})
				acc = ""
			}
			tokens = append(tokens, Token{Type: LBRACK})
		} else if c == ']' {
			if acc != "" {
				tokens = append(tokens, Token{Type: IDENT, Value: acc})
				acc = ""
			}
			tokens = append(tokens, Token{Type: RBRACK})
		} else if c == '.' {
			if acc != "" {
				tokens = append(tokens, Token{Type: IDENT, Value: acc})
				acc = ""
			}
			tokens = append(tokens, Token{Type: DOT})
		} else {
			acc += string(c)
			continue
		}
	}
	if acc != "" {
		tokens = append(tokens, Token{Type: IDENT, Value: acc})
		acc = ""
	}

	fmt.Println(tokens)
	t := Tokens{
		tokens: tokens,
		pos:    0,
	}
	return t
}

//func parseBracket(key string, tokens []Token) Match {
//for i = 0; i < len(tokens); i++ {
//token := tokens[i]
//if token.Type == IDENT && tokens[i+1].Type == RBRACK {

//return Multiple{
//Key:   key,
//Value: token.Value,
//In:    parse(tokens[i:]),
//}
//} else {
//fmt.Println("fail to parse")
//os.Exit(1)
//}
//}
//}

type Tokens struct {
	tokens []Token
	pos    int
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

func parse(tokens []Token) Match {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch token.Type {
		case IDENT:
			if i == len(tokens)-1 {
				return &Single{Name: token.Value}
			} else if tokens, m := tokMatch(tokens, DOT); m {
				return &Obj{
					Name: token.Value,
					In:   parse(tokens[1:]),
				}
			} else if tokens, m := tokMatch(tokens, LBRACK); m {
				if tokens, m := tokMatch(tokens, IDENT); m {
					v := tokens[0].Value
					if tokens, m := tokMatch(tokens, RBRACK); m {
						return &Multiple{
							Key:   token.Value,
							Value: fmt.Sprintf(`"%s"`, v),
							In:    parse(tokens[1:]),
						}
					}
				} else {
					fmt.Println("parse error expecting [IDENT]", tokens)
					os.Exit(1)
				}
			} else {
				if len(tokens[i+1:]) > 0 {
					fmt.Println("err")
					os.Exit(1)
				}
				return &Single{Name: token.Value}
			}

		case DOT:
			return parse(tokens[i+1:])

		default:
			fmt.Println("unhandled token", token, tokens)
			os.Exit(1)
		}
	}
	return nil
}

// expr = term
// expr = (term.)+ expr
// expr = (term[term].)+ expr

func parse2(tokens Tokens) Match {

}

func parseSingle(tokens []Token) ([]Token, Match) {
	tok
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("not enough args")
		os.Exit(1)
	}

	jobFile := os.Args[1]
	queryPath := os.Args[2]
	value := os.Args[3]

	f, err := os.Open(jobFile)
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
	data := make([]byte, 10000)
	count, err := f.Read(data)

	p, err := parser.Parse(data[:count])
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}

	//ast.Walk(p, walkDebug)

	w := newWalker(queryPath, value)

	tokens := lexer(queryPath)
	m := parse(tokens)

	fmt.Println(w.match)
	fmt.Println(m)

	w.match = m

	rewritten := ast.Walk(p, w.walk)

	printer.Fprint(os.Stdout, rewritten)
}
