package tui

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
)

type sourceAddedMsg struct {
	result SourceActionResult
	err    error
}

type sourceRemovedMsg struct {
	result SourceActionResult
	err    error
}

func sanitizeSourceInput(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")

	return value
}

func trimLastRune(value string) string {
	if value == "" {
		return value
	}

	_, size := utf8.DecodeLastRuneInString(value)
	if size <= 0 {
		return value
	}

	return value[:len(value)-size]
}

func (m *Model) handleSourceInput(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	if !m.sourceInputActive {
		return *m, nil, false
	}

	switch msg.String() {
	case "ctrl+c":
		return *m, nil, false
	case "esc":
		m.sourceInputActive = false
		m.sourceInputValue = ""
		m.statusMessage = "Source add cancelled"
		return *m, nil, true
	case "enter":
		cmd := m.startAddSource()
		return *m, cmd, true
	case "backspace":
		m.sourceInputValue = trimLastRune(m.sourceInputValue)
		return *m, nil, true
	}

	if msg.Text != "" {
		m.sourceInputValue += sanitizeSourceInput(msg.Text)
		return *m, nil, true
	}

	return *m, nil, true
}

func (m *Model) handleSourcePaste(msg tea.PasteMsg) (Model, tea.Cmd, bool) {
	if !m.sourceInputActive {
		return *m, nil, false
	}

	m.sourceInputValue += sanitizeSourceInput(msg.String())

	return *m, nil, true
}

func (m *Model) beginAddSource() {
	if m.section != sectionSources {
		return
	}

	m.sourceInputActive = true
	m.sourceInputValue = ""
	m.statusMessage = "Paste a GitHub repo or tree URL and press enter • tree URLs are best for precision"
}

func (m *Model) startAddSource() tea.Cmd {
	if !m.sourceInputActive {
		return nil
	}

	locator := strings.TrimSpace(m.sourceInputValue)
	if locator == "" {
		m.statusMessage = "Source URL required"
		return nil
	}
	if m.workflow == nil || m.refreshing || m.syncing {
		m.statusMessage = "Source add is unavailable"
		return nil
	}

	m.sourceInputActive = false
	m.refreshing = true
	m.statusMessage = "Adding source"

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.workflow.AddSource(context.Background(), locator)
			return sourceAddedMsg{result: result, err: err}
		},
	)
}

func (m *Model) startRemoveSource() tea.Cmd {
	if !m.sourceRemoveConfirmActive {
		return nil
	}
	if m.workflow == nil || m.refreshing || m.syncing {
		m.statusMessage = "Source remove is unavailable"
		return nil
	}

	configuredSource := m.sourceToRemove
	m.sourceRemoveConfirmActive = false
	m.sourceToRemove = source.Source{}
	m.refreshing = true
	m.statusMessage = "Removing source"

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.workflow.RemoveSource(context.Background(), configuredSource.Locator())
			return sourceRemovedMsg{result: result, err: err}
		},
	)
}

func (m *Model) finishAddSource(msg sourceAddedMsg) {
	m.refreshing = false
	m.sourceInputValue = ""

	if msg.result.Snapshot != nil {
		m.applySnapshot(*msg.result.Snapshot)
		m.setSourceCursor(msg.result.Source.ID())
		m.ensureOffset()
	}

	if msg.err != nil {
		if msg.result.Snapshot == nil {
			m.sourceInputActive = true
		}
		if msg.result.Summary != "" {
			m.statusMessage = fmt.Sprintf("%s • error: %v", msg.result.Summary, msg.err)
			return
		}

		m.statusMessage = fmt.Sprintf("Add source failed: %v", msg.err)
		return
	}

	if msg.result.Summary != "" {
		m.statusMessage = msg.result.Summary
		return
	}

	m.statusMessage = "Source added"
}

func (m *Model) finishRemoveSource(msg sourceRemovedMsg) {
	m.refreshing = false

	if msg.result.Snapshot != nil {
		m.applySnapshot(*msg.result.Snapshot)
	}
	if _, ok := m.currentSource(); !ok {
		m.activeSourceID = ""
	}
	if len(m.snapshot.Sources) == 0 {
		m.cursor = 0
		m.offset = 0
	} else {
		if m.cursor >= len(m.snapshot.Sources) {
			m.cursor = len(m.snapshot.Sources) - 1
		}
		m.ensureOffset()
	}

	if msg.err != nil {
		if msg.result.Summary != "" {
			m.statusMessage = fmt.Sprintf("%s • error: %v", msg.result.Summary, msg.err)
			return
		}

		m.statusMessage = fmt.Sprintf("Remove source failed: %v", msg.err)
		return
	}

	if msg.result.Summary != "" {
		m.statusMessage = msg.result.Summary
		return
	}

	m.statusMessage = "Source removed"
}

func (m *Model) setSourceCursor(sourceID string) {
	if strings.TrimSpace(sourceID) == "" {
		return
	}

	for index, configuredSource := range m.snapshot.Sources {
		if configuredSource.ID() == sourceID {
			m.cursor = index
			return
		}
	}
}
