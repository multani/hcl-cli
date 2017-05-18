package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/spf13/cobra"
)

func formatCommand(cmd *cobra.Command, args []string) {
	if len(args) != 0 && len(args) != 1 {
		fmt.Println("not enough args")
		cmd.Usage()
		os.Exit(1)
	}

	fp, closeFunc := func() (*os.File, func() error) {
		if len(args) == 1 {
			fp, err := os.Open(args[0])
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

	output, err := printer.Format(data)
	if err != nil {
		panic(fmt.Sprintf("error while formatting: %v", err))
	}
	os.Stdout.Write(output)
}

func FormatCommandFactory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt [HCL_FILE]",
		Short: "Format HCL",
		Long:  `long help`,
		Run:   formatCommand,
	}

	return cmd
}
