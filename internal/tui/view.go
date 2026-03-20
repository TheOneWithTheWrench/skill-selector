package tui

import (
	"fmt"
	"math"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	if !m.ready {
		view := tea.NewView("\n  Starting skill-selector...")
		view.AltScreen = true
		return view
	}

	totalWidth := m.totalWidth()
	leftWidth := m.leftPaneWidth()
	rightWidth := m.rightPaneWidth()
	bodyHeight := m.bodyHeight()
	headerContentWidth := max(1, totalWidth-headerPanelStyle.GetHorizontalPadding()-headerPanelStyle.GetHorizontalBorderSize())
	panelContentWidth := func(panelWidth int) int {
		return max(1, panelWidth-panelStyle.GetHorizontalPadding()-panelStyle.GetHorizontalBorderSize())
	}

	header := headerPanelStyle.Width(totalWidth).Render(m.renderHeader(headerContentWidth))
	leftPanel := panelStyle.Width(leftWidth).Height(bodyHeight).Render(m.renderLeftPane(panelContentWidth(leftWidth), bodyHeight))
	rightPanel := panelStyle.Width(rightWidth).Height(bodyHeight).Render(m.renderRightPane(panelContentWidth(rightWidth), bodyHeight))
	footer := m.renderFooter()

	content := strings.Join([]string{
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel),
		footer,
	}, "\n\n")
	if m.sourceRemoveConfirmActive || m.profileRemoveConfirmActive {
		content = m.renderConfirmView()
	}

	view := tea.NewView(appStyle.Render(content))
	view.AltScreen = true

	return view
}

func (m Model) renderHeader(width int) string {
	if width < 1 {
		return ""
	}

	if width < 80 || m.height < 32 {
		profileLine := lipgloss.JoinHorizontal(
			lipgloss.Left,
			logoStyle.Render("Skill Selector"),
			" ",
			badgeStyle.Render(m.selectionOwnerLabel()),
		)
		statsLine := strings.Join([]string{
			m.renderHeaderStat("sources", fmt.Sprintf("%d", len(m.snapshot.Sources))),
			m.renderHeaderStat("catalog", fmt.Sprintf("%d", len(m.snapshot.Catalog.Skills()))),
			m.renderHeaderStat("selected", fmt.Sprintf("%d", m.selectionSummary().SelectedCount)),
			m.renderHeaderStat("focus", strings.ToLower(m.focusLabel())),
		}, "  ")

		return strings.Join([]string{
			profileLine,
			m.renderHeaderTabs(),
			statsLine,
		}, "\n")
	}

	logoWidth := max(1, width*2/3)
	metaWidth := max(1, width-logoWidth-1)

	logo := lipgloss.NewStyle().Width(logoWidth).Render(
		logoStyle.Render(m.renderLogo(logoWidth)) + "\n" + logoTaglineStyle.Render("source-aware skill catalogs for AI agents"),
	)
	meta := lipgloss.NewStyle().Width(metaWidth).Align(lipgloss.Right).Render(m.renderHeaderMeta(metaWidth))

	return lipgloss.JoinHorizontal(lipgloss.Top, logo, " ", meta)
}

func (m Model) renderHeaderMeta(width int) string {
	summary := m.selectionSummary()

	return lipgloss.NewStyle().Width(max(1, width)).Align(lipgloss.Right).Render(strings.Join([]string{
		lipgloss.JoinHorizontal(lipgloss.Left, headerStyle.Render("PROFILE"), " ", badgeStyle.Render(m.selectionOwnerLabel())),
		"",
		m.renderHeaderTabs(),
		"",
		m.renderHeaderStat("sources", fmt.Sprintf("%d", len(m.snapshot.Sources))),
		m.renderHeaderStat("catalog", fmt.Sprintf("%d skills", len(m.snapshot.Catalog.Skills()))),
		m.renderHeaderStat("selected", fmt.Sprintf("%d", summary.SelectedCount)),
		m.renderHeaderStat("changes", summary.PendingLabel()),
		m.renderHeaderStat("focus", strings.ToLower(m.focusLabel())),
	}, "\n"))
}

func (m Model) renderLogo(width int) string {
	if width < 64 {
		return "Skill Selector"
	}

	return strings.Trim(logoRaw, "\n")
}

func (m Model) renderHeaderTabs() string {
	tabs := make([]string, len(topLevelSections))
	for index, currentSection := range topLevelSections {
		style := inactiveTabStyle
		if currentSection == m.currentTopLevelSection() {
			style = activeTabStyle
		}

		tabs[index] = style.Render(strings.ToLower(sectionTitles[currentSection]))
	}

	return strings.Join(tabs, " ")
}

func (m Model) renderHeaderStat(label string, value string) string {
	return headerMetaLabelStyle.Render(strings.ToUpper(label)+":") + " " + headerMetaValueStyle.Render(value)
}

func (m Model) paneTitle() string {
	if m.section == sectionCatalog {
		if configuredSource, ok := m.currentSource(); ok {
			return configuredSource.ID()
		}

		return "source skills"
	}

	return sectionTitles[m.section]
}

func (m Model) focusLabel() string {
	if m.section == sectionCatalog {
		if configuredSource, ok := m.currentSource(); ok {
			return "source / " + configuredSource.ID()
		}

		return "source"
	}

	return sectionTitles[m.currentTopLevelSection()]
}

func (m Model) selectionOwnerLabel() string {
	return m.snapshot.Profiles.Active().Name()
}

func (m Model) renderLeftPane(width int, height int) string {
	items := m.currentItems()
	control := controlBoxStyle.Width(max(1, width-2)).Render(m.renderControlBox(items))
	content := strings.Join([]string{
		panelTitleStyle.Render(strings.ToLower(m.paneTitle())),
		panelMetaStyle.Render(m.renderPaneMeta(items)),
		"",
		control,
		"",
		m.renderList(width, height),
	}, "\n")

	return lipgloss.NewStyle().MaxHeight(m.panelInnerHeight(height)).Render(content)
}

func (m Model) renderRightPane(width int, height int) string {
	content := strings.Join([]string{
		panelTitleStyle.Render("preview"),
		panelMetaStyle.Render(m.renderPreviewMeta()),
		"",
		m.renderDetail(width, height),
	}, "\n")

	return lipgloss.NewStyle().MaxHeight(m.panelInnerHeight(height)).Render(content)
}

func (m Model) renderPaneLabel() string {
	switch m.section {
	case sectionSources:
		if m.sourceInputActive {
			return "new source"
		}

		return "registered sources"
	case sectionCatalog:
		return "source skills"
	case sectionProfiles:
		if m.profileInputActive {
			if m.profileInputMode == profileInputRename {
				return "rename profile"
			}

			return "new profile"
		}

		return "profiles"
	default:
		return "sync status"
	}
}

func (m Model) renderPaneMeta(items []item) string {
	if len(items) == 0 {
		if m.section == sectionSources && m.sourceInputActive {
			return "enter a GitHub tree URL"
		}
		if m.section == sectionProfiles && m.profileInputActive {
			return "enter a profile name"
		}
		return "nothing to show yet"
	}

	if m.section == sectionCatalog {
		return fmt.Sprintf("%d skills • %d selected", len(items), m.selection.DesiredCountForSource(m.activeSourceID))
	}
	if m.section == sectionProfiles {
		return fmt.Sprintf("%d profiles • active %s", len(items), m.snapshot.Profiles.ActiveName())
	}

	return fmt.Sprintf("%d items • cursor %d", len(items), min(m.cursor+1, len(items)))
}

func (m Model) renderPreviewMeta() string {
	items := m.currentItems()
	if len(items) == 0 || m.cursor >= len(items) {
		return "select an item to preview"
	}

	return fmt.Sprintf("previewing %d of %d", m.cursor+1, len(items))
}

func (m Model) renderSectionSummary(items []item) string {
	if m.section == sectionSources && m.sourceInputActive {
		if strings.TrimSpace(m.sourceInputValue) == "" {
			return "paste URL and press enter"
		}

		return m.sourceInputValue
	}
	if m.section == sectionProfiles && m.profileInputActive {
		if strings.TrimSpace(m.profileInputValue) == "" {
			return "type profile name and press enter"
		}

		return m.profileInputValue
	}

	if len(items) == 0 {
		return "0 items"
	}

	if m.section == sectionCatalog {
		return fmt.Sprintf("%d of %d selected", m.selection.DesiredCountForSource(m.activeSourceID), len(items))
	}
	if m.section == sectionProfiles {
		return fmt.Sprintf("active %s", m.snapshot.Profiles.ActiveName())
	}

	return fmt.Sprintf("cursor %d of %d", min(m.cursor+1, len(items)), len(items))
}

func (m Model) renderControlBox(items []item) string {
	return strings.Join([]string{
		controlLabelStyle.Render(m.renderPaneLabel()),
		controlValueStyle.Render(m.renderSectionSummary(items)),
	}, "\n")
}

func (m Model) renderList(width int, height int) string {
	_ = height

	items := m.currentItems()
	contentWidth := m.listContentWidth(width)
	viewportHeight := m.listViewportLineCount(width)
	vp := newListViewport(contentWidth, viewportHeight)
	vp.SetContent(renderListContent(items, m.cursor, contentWidth))
	vp.SetYOffset(m.offset)

	content := vp.View()
	if vp.TotalLineCount() <= vp.VisibleLineCount() {
		return content
	}

	scrollbar := renderViewportScrollbar(vp)
	return lipgloss.JoinHorizontal(lipgloss.Top, content, " ", scrollbar)
}

func renderListContent(items []item, cursor int, width int) string {
	rows := make([]string, 0, max(1, len(items)*listItemHeight))

	for index, entry := range items {
		rows = append(rows, renderListItemRows(entry, index == cursor, width)...)
		if index < len(items)-1 {
			rows = append(rows, "")
		}
	}

	return strings.Join(rows, "\n")
}

func renderViewportScrollbar(vp viewport.Model) string {
	trackHeight := max(1, vp.Height())
	visibleLines := max(1, vp.VisibleLineCount())
	totalLines := max(1, vp.TotalLineCount())
	maxOffset := max(1, totalLines-visibleLines)
	thumbSize := min(max(1, int(math.Round(float64(trackHeight*visibleLines)/float64(totalLines)))), trackHeight)

	thumbTop := 0
	if trackHeight > thumbSize {
		thumbTop = int(math.Round(float64((trackHeight-thumbSize)*vp.YOffset()) / float64(maxOffset)))
	}

	rows := make([]string, 0, trackHeight)
	for index := range trackHeight {
		style := scrollTrackStyle
		char := "│"
		if index >= thumbTop && index < thumbTop+thumbSize {
			style = scrollThumbStyle
			char = "█"
		}

		rows = append(rows, style.Render(char))
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderDetail(width int, height int) string {
	items := m.currentItems()
	if len(items) == 0 {
		return warningStyle.Render("Nothing to show yet.")
	}

	if m.cursor >= len(items) {
		return warningStyle.Render("Selection is out of range.")
	}

	selected := items[m.cursor]
	body := m.renderDetailBody(selected, width, height)
	if body == "" {
		body = detailBodyStyle.Render("No details yet.")
	}

	sections := []string{detailTitleStyle.Render(selected.Title)}
	if selected.Subtitle != "" {
		sections = append(sections, detailMetaStyle.Render(selected.Subtitle))
	}

	chips := m.renderSkillTagChips()
	if chips != "" {
		sections = append(sections, chips)
	}

	sections = append(sections, renderDivider(width), body)

	return strings.Join(sections, "\n\n")
}

func renderListItemRows(item item, selected bool, width int) []string {
	title := item.Title
	if item.Selectable && item.Selected {
		title = "[x] " + title
	} else if item.Selectable {
		title = "[ ] " + title
	}

	titleStyle, subtitleStyle := listItemStyles(item, selected, width)

	subtitle := item.Subtitle
	if subtitle == "" {
		subtitle = " "
	}

	titleRows := strings.Split(titleStyle.Render(title), "\n")
	subtitleRows := strings.Split(subtitleStyle.Render(subtitle), "\n")

	return append(titleRows, subtitleRows...)
}

func (m Model) renderDetailBody(selected item, width int, height int) string {
	trimmed := strings.TrimSpace(selected.Detail)
	if trimmed == "" {
		return ""
	}

	parts := strings.Split(trimmed, "\n\n")
	blocks := []string{
		sectionHeadingStyle.Render("overview"),
		detailBodyStyle.Width(width).Render(parts[0]),
	}

	if len(parts) > 1 {
		blocks = append(blocks,
			renderDivider(width),
			sectionHeadingStyle.Render("details"),
			detailBodyStyle.Width(width).Render(strings.Join(parts[1:], "\n\n")),
		)
	}

	visibleLines := max(4, height-detailPanelChromeHeight)
	return clipLines(strings.Join(blocks, "\n\n"), visibleLines)
}

func (m Model) renderSkillTagChips() string {
	if m.section != sectionCatalog {
		return ""
	}

	discoveredSkill, ok := m.currentCatalogSkill()
	if !ok || len(discoveredSkill.Tags()) == 0 {
		return ""
	}

	labels := make([]string, 0, min(len(discoveredSkill.Tags()), 4))
	for index, tag := range discoveredSkill.Tags() {
		if index == 3 {
			labels = append(labels, fmt.Sprintf("+%d more", len(discoveredSkill.Tags())-index))
			break
		}
		labels = append(labels, tag)
	}

	chips := make([]string, 0, len(labels))
	for _, label := range labels {
		chips = append(chips, chipStyle.Render(label))
	}

	return strings.Join(chips, " ")
}

func renderDivider(width int) string {
	lineWidth := max(8, width-1)
	return dividerStyle.Width(lineWidth).Render(strings.Repeat("-", lineWidth))
}

func clipLines(value string, maxLines int) string {
	lines := strings.Split(value, "\n")
	if len(lines) <= maxLines {
		return value
	}
	if maxLines <= 1 {
		return warningStyle.Render("...")
	}

	clipped := append([]string{}, lines[:maxLines-1]...)
	clipped = append(clipped, warningStyle.Render("... more ..."))

	return strings.Join(clipped, "\n")
}
