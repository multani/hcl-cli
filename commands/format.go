package commands

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/spf13/cobra"

	"github.com/multani/hcl-cli/hcl/misc"
)

func formatCommand(cmd *cobra.Command, args []string) {
	if len(args) != 0 && len(args) != 1 {
		fmt.Println("not enough args")
		cmd.Usage()
		os.Exit(1)
	}

	data, err := misc.FileOrStdinContent(args, 0)
	if err != nil {
		fmt.Printf("error while reading data: %v\n", err)
		os.Exit(255)
	}

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
		Long: `Correctly format, using a canonical representation, HCL data.

If no file is specified, it will read HCL content from the standard input.`,
		Run: formatCommand,
	}

	return cmd
}
