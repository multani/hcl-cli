package main

import (
	"fmt"
	"os"

	"github.com/multani/hcl-cli/commands"
	"github.com/spf13/cobra"
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "hcl",
		Short: "HCL command-line tool",
		Long:  `blabla`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	setCmd := commands.SetCommandFactory()
	rootCmd.AddCommand(setCmd)

	formatCmd := commands.FormatCommandFactory()
	rootCmd.AddCommand(formatCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
