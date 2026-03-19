package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
)

type refreshCompletedMsg struct {
	result RefreshActionResult
	err    error
}

func (m *Model) startRefreshSources() tea.Cmd {
	if m.refreshing || m.syncing {
		return nil
	}
	if m.workflow == nil {
		m.statusMessage = "Source refresh is unavailable"
		return nil
	}

	m.refreshing = true
	m.statusMessage = "Refreshing sources"

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.workflow.Refresh(context.Background())
			return refreshCompletedMsg{result: result, err: err}
		},
	)
}

func (m *Model) finishRefresh(msg refreshCompletedMsg) {
	m.refreshing = false
	if msg.result.Snapshot != nil {
		m.applySnapshot(*msg.result.Snapshot)
		if _, ok := m.currentSource(); !ok {
			m.activeSourceID = ""
		}
		m.ensureOffset()
	}

	if msg.err != nil {
		if msg.result.Summary != "" {
			m.statusMessage = fmt.Sprintf("%s • error: %v", msg.result.Summary, msg.err)
			return
		}

		m.statusMessage = fmt.Sprintf("Refresh failed: %v", msg.err)
		return
	}

	if msg.result.Summary != "" {
		m.statusMessage = msg.result.Summary
		return
	}

	m.statusMessage = "Sources refreshed"
}
