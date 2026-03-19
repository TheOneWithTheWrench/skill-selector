package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// Run builds the Cobra command tree, executes it, and routes output to the provided writers.
func Run(args []string, stdout io.Writer, stderr io.Writer, application Application, openTUI func() error) error {
	if application == nil {
		return fmt.Errorf("cli application required")
	}
	if openTUI == nil {
		return fmt.Errorf("cli tui launcher required")
	}

	rootCommand := newRootCommand(stdout, stderr, application, openTUI)
	if len(args) > 1 {
		rootCommand.SetArgs(args[1:])
	} else {
		rootCommand.SetArgs(nil)
	}

	return fang.Execute(context.Background(), rootCommand)
}

func helpRunE(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}
