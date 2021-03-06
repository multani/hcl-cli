package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/spf13/cobra"

	"github.com/multani/hcl-cli/hcl/misc"
	"github.com/multani/hcl-cli/hcl/query"
)

type Walker struct {
	query     query.Query
	value     string
	valueType token.Type
}

func (w *Walker) FormatValue() string {
	switch w.valueType {

	case token.STRING:
		return fmt.Sprintf(`"%s"`, w.value)

	case token.BOOL:
		v, err := strconv.ParseBool(w.value)
		if err != nil {
			panic(err)
		}
		if v {
			return "true"
		} else {
			return "false"
		}

	case token.NUMBER:
		_, err := strconv.ParseInt(w.value, 10, 64)
		if err != nil {
			panic(err)
		}
		return w.value

	case token.FLOAT:
		_, err := strconv.ParseFloat(w.value, 64)
		if err != nil {
			panic(err)
		}
		return w.value

	default:
		panic("unknown formatting")
	}
}

func (w Walker) walk(n ast.Node) (ast.Node, bool) {
	switch i := n.(type) {
	case *ast.ObjectItem:
		switch m := w.query.(type) {
		case *query.KeyValue:
			if len(i.Keys) == 2 && i.Keys[0].Token.Text == m.Key && i.Keys[1].Token.Text == m.Value {
				w.query = m.In
				i.Val = ast.Walk(i.Val, w.walk)
				return i, false
			}

		case *query.Obj:
			if len(i.Keys) == 1 && i.Keys[0].Token.Text == m.Name {
				w.query = m.In
				i.Val = ast.Walk(i.Val, w.walk)
				return i, false
			}

		case *query.Single:
			if len(i.Keys) == 1 && i.Keys[0].Token.Text == m.Name {
				i.Val = &ast.LiteralType{
					Token: token.Token{
						Type: w.valueType,
						Text: w.FormatValue(),
					},
				}
				return i, false
			}
		}
	}
	return n, true
}

var ValueType string

func setCommand(cmd *cobra.Command, args []string) {
	if len(args) != 2 && len(args) != 3 {
		fmt.Println("not enough args")
		cmd.Usage()
		os.Exit(1)
	}

	queryPath := args[0]
	value := args[1]

	data, err := misc.FileOrStdinContent(args, 2)
	if err != nil {
		fmt.Printf("error while reading data: %v\n", err)
		os.Exit(255)
	}

	p, err := parser.Parse(data)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	typeToken := map[string]token.Type{
		"string": token.STRING,
		"bool":   token.BOOL,
		"int":    token.NUMBER,
		"float":  token.FLOAT,
	}

	t, present := typeToken[ValueType]
	if !present {
		keys := make([]string, 0)
		for key, _ := range typeToken {
			keys = append(keys, key)
		}
		msg := fmt.Sprintf("type should be one of: %s", strings.Join(keys, ", "))
		fmt.Println(msg)
		cmd.Usage()
		os.Exit(1)
	}

	tokens := query.Lex(queryPath)

	w := &Walker{
		query:     tokens.Parse(),
		value:     value,
		valueType: t,
	}

	rewritten := ast.Walk(p, w.walk)
	printer.Fprint(os.Stdout, rewritten)
	os.Stdout.WriteString("\n")
}

func SetCommandFactory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set QUERY_PATH VALUE [HCL_FILE]",
		Short: "Set the value of a HCL attribute",
		Long: `Edit the content of a HCL attribute.

This allows to access specific attribute in a HCL file and set it to a new
value.  Only simple values are supported (you can't set the content of an
attribute with a structure for example).

Syntax for accessing an attribute:
----------------------------------

  * IDENT
    access a single field
  * (IDENT ".")+ IDENT
    access sub item of a simple structure
  * (IDENT "[" IDENT "]" .) IDENT
    access items from a named structure


Example:
--------

  Considering the following HCL data:

    foo = true
    obj {
      val = 56
    }
    some "map" {
      item = 42
      obj {
        val = "abc"
      }
    }

  - "foo"
        query the value: true
  - "obj.val"
        query the value: 56
  - "some[map].item"
        query the value: 42
  - "some[map].obj.val"
        query the value: "abc"
`,
		Run: setCommand,
	}

	cmd.PersistentFlags().StringVarP(
		&ValueType, "type", "t", "string", "Type of the value to set")

	return cmd
}
