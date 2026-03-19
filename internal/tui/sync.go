package tui

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

type syncCompletedMsg struct {
	result SyncActionResult
	err    error
}

func newSpinner() spinner.Model {
	return spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(chosenItemStyle),
	)
}

func (m *Model) startSync() tea.Cmd {
	if m.syncing {
		return nil
	}
	if m.workflow == nil {
		m.statusMessage = "Sync is unavailable"
		return nil
	}

	desiredSelection := m.selection.Desired()
	m.syncing = true
	m.statusMessage = "Syncing selection"

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.workflow.Sync(context.Background(), desiredSelection)
			return syncCompletedMsg{result: result, err: err}
		},
	)
}

func (m *Model) finishSync(msg syncCompletedMsg) {
	m.syncing = false
	if msg.result.HasState {
		m.applySyncAction(msg.result)
	}

	if msg.err != nil {
		if summary := msg.result.Result.Summary(); summary != "" {
			m.statusMessage = fmt.Sprintf("%s • error: %v", summary, msg.err)
			return
		}

		m.statusMessage = fmt.Sprintf("Sync failed: %v", msg.err)
		return
	}

	if summary := msg.result.Result.Summary(); summary != "" {
		m.statusMessage = summary
		return
	}

	m.statusMessage = "Sync completed"
}

func (m *Model) applySyncAction(result SyncActionResult) {
	m.snapshot.Manifests = manifestsFromResult(result.Manifests)
	m.snapshot.Warnings = append([]string(nil), result.Warnings...)
	m.selection.ReplaceActive(result.ActiveSelection)
}
