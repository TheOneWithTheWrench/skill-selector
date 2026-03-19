package tui

import (
	"fmt"
	"strings"
)

type helpItem struct {
	Key  string
	Desc string
}

func (m Model) renderFooter() string {
	return footerRowStyle.Render(helpBoxStyle.Width(max(1, m.totalWidth())).Render(m.renderHelpBox()))
}

func (m Model) renderHelpBox() string {
	controlLine := strings.Join([]string{
		helpTitleStyle.Render("controls"),
		renderHelpItems([]helpItem{{Key: "tab/h/l/<-/->", Desc: "sections"}, {Key: "j/k/up/down", Desc: "move"}, {Key: "ctrl+d/u", Desc: "page"}, {Key: "q", Desc: "quit"}}),
		helpDividerStyle.Render("|"),
		helpTitleStyle.Render(strings.ToLower(m.helpContextLabel())),
		renderHelpItems(m.contextHelpItems()),
	}, " ")

	lines := []string{controlLine}
	if m.shouldRenderFooterStatus() {
		lines = append(lines, footerStatusLabelStyle.Render("status")+" "+footerStatusValueStyle.Render(m.footerStatusValue()))
	}

	return strings.Join(lines, "\n")
}

func (m Model) shouldRenderFooterStatus() bool {
	if m.syncing || m.refreshing || m.statusMessage != "" || m.sourceInputActive || m.sourceRemoveConfirmActive {
		return true
	}
	if len(m.snapshot.Warnings) > 0 {
		return true
	}

	return m.selectionSummary().Dirty()
}

func (m Model) footerStatusValue() string {
	if m.refreshing {
		if m.statusMessage != "" {
			return m.spinner.View() + " " + strings.ToLower(m.statusMessage)
		}

		return m.spinner.View() + " refreshing sources"
	}

	if m.syncing {
		if m.statusMessage != "" {
			return m.spinner.View() + " " + strings.ToLower(m.statusMessage)
		}

		return m.spinner.View() + " syncing selection"
	}

	if m.statusMessage != "" {
		return m.statusMessage
	}

	summary := m.selectionSummary()
	if len(m.snapshot.Warnings) > 0 && !summary.Dirty() {
		return m.snapshot.Warnings[0]
	}

	return fmt.Sprintf("%s • %s", summary.SelectionLabel(), summary.PendingLabel())
}

func (m Model) contextHelpItems() []helpItem {
	if m.sourceInputActive {
		return []helpItem{{Key: "enter", Desc: "save"}, {Key: "esc", Desc: "cancel"}, {Key: "backspace", Desc: "erase"}}
	}
	if m.sourceRemoveConfirmActive {
		return []helpItem{{Key: "y/enter", Desc: "confirm"}, {Key: "n/esc", Desc: "cancel"}}
	}

	switch m.section {
	case sectionSources:
		return []helpItem{{Key: "enter", Desc: "browse"}, {Key: "a", Desc: "add"}, {Key: "d", Desc: "remove"}, {Key: "r/R", Desc: "refresh"}}
	case sectionCatalog:
		return []helpItem{{Key: "space", Desc: "toggle"}, {Key: "s", Desc: "sync"}, {Key: "esc", Desc: "back"}}
	case sectionProfiles:
		return []helpItem{{Key: "planned", Desc: "next slice"}}
	default:
		return []helpItem{{Key: "s", Desc: "sync"}}
	}
}

func (m Model) helpContextLabel() string {
	if m.section == sectionCatalog {
		return "source"
	}
	if m.sourceRemoveConfirmActive {
		return "confirm"
	}

	return sectionTitles[m.section]
}

func renderHelpItems(items []helpItem) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, helpKeyStyle.Render(item.Key)+" "+helpDescStyle.Render(item.Desc))
	}

	return strings.Join(parts, " "+helpDividerStyle.Render("•")+" ")
}
