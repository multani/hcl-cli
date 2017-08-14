package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/spf13/cobra"

	"github.com/multani/hcl-cli/hcl/misc"
	"github.com/multani/hcl-cli/hcl/query"
)

type GetWalker struct {
	query  query.Query
	writer io.Writer
}

func (w GetWalker) walk(n ast.Node) (ast.Node, bool) {
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
				w.write(i.Val)
				return i, false
			}
		}
	}
	return n, true
}

func (w GetWalker) write(n ast.Node) error {
	switch node := n.(type) {
	case *ast.LiteralType:

		switch value := node.Token.Value().(type) {
		case string:
			w.writer.Write([]byte(value))
		case int64:
			w.writer.Write([]byte(fmt.Sprintf("%d", value)))
		case float64:
			w.writer.Write([]byte(fmt.Sprintf("%d", value)))
		case bool:
			w.writer.Write([]byte(fmt.Sprintf("%t", value)))
		default:
			panic(fmt.Sprintf("unknown type value: %s", value))
		}

	default:
		panic(fmt.Sprintf("unknown node type: %s", node))
	}
	w.writer.Write([]byte("\n"))
	return nil
}

func getCommand(cmd *cobra.Command, args []string) {
	if len(args) != 1 && len(args) != 2 {
		fmt.Println("not enough args")
		cmd.Usage()
		os.Exit(1)
	}

	queryPath := args[0]
	data, err := misc.FileOrStdinContent(args, 1)
	if err != nil {
		fmt.Printf("error while reading data: %v\n", err)
		os.Exit(255)
	}

	p, err := parser.Parse(data)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	tokens := query.Lex(queryPath)

	w := &GetWalker{
		query:  tokens.Parse(),
		writer: os.Stdout,
	}

	ast.Walk(p, w.walk)
}

func GetCommandFactory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get QUERY_PATH [HCL_FILE]",
		Short: "Get the value of a HCL attribute",
		Long: `Get the content of a HCL attribute.

This allows to print out a specific attribute in a HCL file.
Only simple values are supported (you can't get the content of an attribute
with a structure for example).

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

  $ cat test.hcl
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

  $ hcl-cli get "foo" test.hcl
  true
  $ hcl-cli get "obj.val" test.hcl
  56
  $ hcl-cli get "some[map].item" test.hcl
  42
  $ hcl-cli get "some[map].obj.val"
  abc
`,
		Run: getCommand,
	}

	return cmd
}
