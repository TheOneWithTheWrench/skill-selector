package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
)

// Run opens the interactive terminal UI on top of the shared app surface.
func Run(runtime paths.Runtime, application Application) error {
	service, err := NewService(runtime, application)
	if err != nil {
		return err
	}

	initialSnapshot, err := service.Load(context.Background())
	if err != nil {
		return err
	}

	program := tea.NewProgram(New(initialSnapshot, service))
	_, err = program.Run()
	if err != nil {
		return fmt.Errorf("run tui: %w", err)
	}

	return nil
}
