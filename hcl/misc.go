package misc

import (
	"fmt"
	"github.com/hashicorp/hcl/hcl/ast"
)

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
