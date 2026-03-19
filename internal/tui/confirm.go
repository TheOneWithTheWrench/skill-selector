package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
)

func (m *Model) handleSourceRemoveConfirm(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	if !m.sourceRemoveConfirmActive {
		return *m, nil, false
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return *m, nil, false
	case "y", "enter":
		cmd := m.startRemoveSource()
		return *m, cmd, true
	case "n", "esc", "backspace":
		m.cancelRemoveSource()
		return *m, nil, true
	default:
		return *m, nil, true
	}
}

func (m *Model) beginRemoveSourceConfirm() {
	if m.section != sectionSources || len(m.snapshot.Sources) == 0 || m.cursor >= len(m.snapshot.Sources) {
		return
	}
	if m.workflow == nil {
		m.statusMessage = "Source remove is unavailable"
		return
	}

	m.sourceRemoveConfirmActive = true
	m.sourceToRemove = m.snapshot.Sources[m.cursor]
	m.statusMessage = "Confirm source removal"
}

func (m *Model) cancelRemoveSource() {
	m.sourceRemoveConfirmActive = false
	m.sourceToRemove = source.Source{}
	m.statusMessage = "Source removal cancelled"
}

func (m Model) renderConfirmView() string {
	lines := []string{
		confirmTitleStyle.Render("Remove Source?"),
		"",
		confirmBodyStyle.Render(m.sourceToRemove.ID()),
		metaStyle.Render(m.sourceToRemove.Locator()),
		"",
		confirmBodyStyle.Render("This removes the source from the app. Any synced skills from this source stay active until you sync."),
		"",
		confirmHelpStyle.Render("y/enter confirm • n/esc cancel"),
	}

	boxWidth := min(72, max(32, m.totalWidth()-8))
	box := confirmBoxStyle.Width(boxWidth).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(max(1, m.totalWidth()), max(1, m.height-appStyle.GetVerticalPadding()), lipgloss.Center, lipgloss.Center, box)
}
