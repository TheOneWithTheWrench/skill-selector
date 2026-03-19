package main

import (
	"fmt"
	"io"
	"os"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/cli"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
)

func main() {
	if err := run(os.Args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	_ = stderr

	runtimePaths, err := paths.Default()
	if err != nil {
		return err
	}

	application, err := app.New(runtimePaths)
	if err != nil {
		return err
	}

	return cli.Run(args, stdout, stderr, application)
}
