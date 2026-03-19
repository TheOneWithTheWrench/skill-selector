package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Run builds the Cobra command tree, executes it, and routes output to the provided writers.
func Run(args []string, stdout io.Writer, stderr io.Writer, application Application) error {
	if application == nil {
		return fmt.Errorf("cli application required")
	}

	rootCommand := newRootCommand(stdout, stderr, application)
	if len(args) > 1 {
		rootCommand.SetArgs(args[1:])
	} else {
		rootCommand.SetArgs(nil)
	}

	return rootCommand.Execute()
}

func helpRunE(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
