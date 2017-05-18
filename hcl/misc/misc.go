package misc

import (
	"fmt"
	"io"
	"os"

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

func FileOrStdinContent(args []string, index int) ([]byte, error) {
	fp, closeFunc := func() (*os.File, func() error) {
		if len(args) > index {
			fp, err := os.Open(args[index])
			if err != nil {
				fmt.Printf("error: %v", err)
				os.Exit(1)
			}
			return fp, fp.Close
		} else {
			return os.Stdin, func() error { return nil }
		}
	}()

	data := make([]byte, 0)

	for {
		buf := make([]byte, 100)
		count, err := fp.Read(buf)

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("error while reading: %v", err)
			os.Exit(1)
		}
		data = append(data, buf[:count]...)
	}
	defer closeFunc()

	return data, nil
}
