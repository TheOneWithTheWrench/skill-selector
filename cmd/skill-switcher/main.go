package main

import (
	"fmt"
	"io"
	"os"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
)

const usage = "usage: skill-switcher"

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

	_, err = app.New(runtimePaths)
	if err != nil {
		return err
	}

	if len(args) > 1 {
		switch args[1] {
		case "-h", "--help", "help":
			_, err := fmt.Fprintln(stdout, usage)
			return err
		}
	}

	_, err = fmt.Fprintln(stdout, "skill-switcher v2 is under construction")
	return err
}
