package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
)

type section int

type item struct {
	Title      string
	Subtitle   string
	Detail     string
	Active     bool
	Selectable bool
	Selected   bool
}

type itemRange struct {
	StartLine int
	EndLine   int
}

type profileInputMode int

// Model is the Bubble Tea model for the v2 TUI.
type Model struct {
	snapshot                   Snapshot
	workflow                   Workflow
	section                    section
	cursor                     int
	offset                     int
	activeSourceID             string
	sourceListCursor           int
	sourceListOffset           int
	sourceCatalogCursor        int
	sourceCatalogOffset        int
	profileInputActive         bool
	profileInputMode           profileInputMode
	profileInputValue          string
	profileToRename            string
	profileRemoveConfirmActive bool
	profileToRemove            profile.Profile
	width                      int
	height                     int
	ready                      bool
	selection                  draftSelection
	spinner                    spinner.Model
	syncing                    bool
	refreshing                 bool
	sourceInputActive          bool
	sourceInputValue           string
	sourceRemoveConfirmActive  bool
	sourceToRemove             source.Source
	statusMessage              string
}

const (
	profileInputCreate profileInputMode = iota
	profileInputRename
)

const (
	sectionSources section = iota
	sectionCatalog
	sectionProfiles
	sectionStatus
	listItemHeight          = 3
	detailPanelChromeHeight = 8
	contentSectionGapHeight = 4
	minBodyHeight           = 1
)

var sectionTitles = map[section]string{
	sectionSources:  "Sources",
	sectionCatalog:  "Catalog",
	sectionProfiles: "Profiles",
	sectionStatus:   "Status",
}

var topLevelSections = []section{sectionSources, sectionProfiles, sectionStatus}

// New constructs the TUI model from one initial snapshot and workflow adapter.
func New(initial Snapshot, workflow Workflow) Model {
	return Model{
		snapshot:  initial,
		workflow:  workflow,
		selection: newDraftSelection(initial.ActiveSelection(), initialDesiredSelection(initial)),
		spinner:   newSpinner(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.RequestWindowSize
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.ensureOffset()
		return m, nil
	case spinner.TickMsg:
		if m.syncing || m.refreshing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case syncCompletedMsg:
		m.finishSync(msg)
		return m, nil
	case profileActionMsg:
		m.finishProfileSwitch(msg)
		return m, nil
	case profileEditMsg:
		m.finishProfileEdit(msg)
		return m, nil
	case profileRemovedMsg:
		m.finishRemoveProfile(msg)
		return m, nil
	case sourceRemovedMsg:
		m.finishRemoveSource(msg)
		return m, nil
	case sourceAddedMsg:
		m.finishAddSource(msg)
		return m, nil
	case refreshCompletedMsg:
		m.finishRefresh(msg)
		return m, nil
	case tea.PasteMsg:
		if updatedModel, cmd, handled := m.handleProfilePaste(msg); handled {
			return updatedModel, cmd
		}
		if updatedModel, cmd, handled := m.handleSourcePaste(msg); handled {
			return updatedModel, cmd
		}
	case tea.KeyPressMsg:
		if updatedModel, cmd, handled := m.handleProfileRemoveConfirm(msg); handled {
			return updatedModel, cmd
		}
		if updatedModel, cmd, handled := m.handleProfileInput(msg); handled {
			return updatedModel, cmd
		}
		if updatedModel, cmd, handled := m.handleSourceRemoveConfirm(msg); handled {
			return updatedModel, cmd
		}
		if updatedModel, cmd, handled := m.handleSourceInput(msg); handled {
			return updatedModel, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			cmd := m.handleEnter()
			return m, cmd
		case "esc", "backspace":
			m.exitDetailMode()
		case " ", "space":
			if m.section == sectionProfiles {
				cmd := m.startSwitchProfile()
				return m, cmd
			}
			m.toggleCurrentCatalogSkill()
		case "a":
			if m.section == sectionCatalog {
				m.addCurrentSourceSkills()
				return m, nil
			}
			m.beginCreateInput()
		case "c":
			if m.section == sectionCatalog {
				m.clearCurrentSourceSkills()
			}
		case "d":
			m.beginRemoveCurrentItem()
		case "e":
			m.beginRenameCurrentProfile()
		case "r", "R":
			cmd := m.startRefreshSources()
			return m, cmd
		case "tab", "l", "right":
			m.switchTopLevelSection(1)
		case "shift+tab", "h", "left":
			m.switchTopLevelSection(-1)
		case "j", "down":
			items := m.currentItems()
			if len(items) > 0 && m.cursor < len(items)-1 {
				m.cursor++
				m.ensureOffset()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureOffset()
			}
		case "ctrl+d", "pgdown":
			items := m.currentItems()
			if len(items) > 0 {
				m.cursor = min(len(items)-1, m.cursor+max(1, m.visibleCount()/2))
				m.ensureOffset()
			}
		case "ctrl+u", "pgup":
			if m.cursor > 0 {
				m.cursor = max(0, m.cursor-max(1, m.visibleCount()/2))
				m.ensureOffset()
			}
		case "s":
			cmd := m.startSync()
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) bodyHeight() int {
	availableHeight := m.height - appStyle.GetVerticalPadding() - m.headerHeight() - m.footerHeight() - contentSectionGapHeight - panelStyle.GetVerticalPadding() - panelStyle.GetVerticalBorderSize()
	if availableHeight < minBodyHeight {
		return minBodyHeight
	}

	return availableHeight
}

func (m Model) currentItems() []item {
	switch m.section {
	case sectionSources:
		if len(m.snapshot.Sources) == 0 {
			return []item{{
				Title:    "No sources yet",
				Subtitle: "Press a to add a GitHub tree URL",
				Detail:   "Example:\n\nhttps://github.com/anthropics/skills/tree/main/skills",
			}}
		}

		items := make([]item, 0, len(m.snapshot.Sources))
		for _, configuredSource := range m.snapshot.Sources {
			discoveredSkillCount := m.skillCountForSource(configuredSource.ID())
			selectedCount := m.selection.DesiredCountForSource(configuredSource.ID())
			items = append(items, item{
				Title:    configuredSource.ID(),
				Subtitle: fmt.Sprintf("%d skills • %d selected", discoveredSkillCount, selectedCount),
				Detail:   renderSourceDetail(configuredSource, discoveredSkillCount, selectedCount),
			})
		}

		return items
	case sectionCatalog:
		if m.activeSourceID == "" {
			return []item{{
				Title:    "No source selected",
				Subtitle: "Press enter on a source to browse its skills",
				Detail:   "Sources are the entry point to skill browsing.",
			}}
		}

		skills := m.sourceSkills(m.activeSourceID)
		if len(skills) == 0 {
			return []item{{
				Title:    "No skills in source",
				Subtitle: "Run refresh after adding or changing the source",
				Detail:   "This source currently has no discovered skills in the catalog.",
			}}
		}

		items := make([]item, 0, len(skills))
		for _, discoveredSkill := range skills {
			items = append(items, item{
				Title:      discoveredSkill.Name(),
				Subtitle:   displayRelativePath(discoveredSkill.RelativePath()),
				Detail:     renderSkillDetail(discoveredSkill),
				Selectable: true,
				Selected:   m.isSelectedSkill(discoveredSkill),
			})
		}

		return items
	case sectionProfiles:
		items := make([]item, 0, len(m.snapshot.Profiles.All()))
		for _, currentProfile := range m.snapshot.Profiles.All() {
			selectedCount := currentProfile.SelectedCount()
			if currentProfile.Name() == m.snapshot.Profiles.ActiveName() {
				selectedCount = m.selectionSummary().SelectedCount
			}

			subtitle := fmt.Sprintf("%d selected skills", selectedCount)
			if currentProfile.Name() == m.snapshot.Profiles.ActiveName() {
				subtitle = "active • " + subtitle
			}

			items = append(items, item{
				Title:    currentProfile.Name(),
				Subtitle: subtitle,
				Detail:   renderProfileDetail(currentProfile, currentProfile.Name() == m.snapshot.Profiles.ActiveName()),
				Active:   currentProfile.Name() == m.snapshot.Profiles.ActiveName(),
			})
		}

		return items
	default:
		summary := m.selectionSummary()
		items := []item{{
			Title:    "Paths",
			Subtitle: "runtime + selection state",
			Detail:   renderStatusDetail(m.snapshot, summary),
		}}

		for _, syncLocation := range m.syncLocationStates() {
			title := syncLocation.RootPath
			if title == "" {
				title = strings.Join(syncLocation.Adapters, ", ")
			}
			items = append(items, item{
				Title:    title,
				Subtitle: fmt.Sprintf("%s • %d synced skills", strings.Join(syncLocation.Adapters, ", "), len(syncLocation.Identities)),
				Detail:   renderSyncLocationDetail(syncLocation),
			})
		}

		return items
	}
}

func (m *Model) enterCurrentSource() {
	if m.section != sectionSources || m.sourceInputActive || len(m.snapshot.Sources) == 0 || m.cursor >= len(m.snapshot.Sources) {
		return
	}

	m.activeSourceID = m.snapshot.Sources[m.cursor].ID()
	m.sourceListCursor = m.cursor
	m.sourceListOffset = m.offset
	m.section = sectionCatalog
	m.cursor = m.sourceCatalogCursor
	m.offset = m.sourceCatalogOffset
	m.ensureOffset()
}

func (m *Model) handleEnter() tea.Cmd {
	if m.section == sectionSources {
		m.enterCurrentSource()
	}

	return nil
}

func (m *Model) exitSourceDetail() {
	if m.section != sectionCatalog {
		return
	}

	m.sourceCatalogCursor = m.cursor
	m.sourceCatalogOffset = m.offset
	m.activeSourceID = ""
	m.section = sectionSources
	m.cursor = m.sourceListCursor
	m.offset = m.sourceListOffset
	m.ensureOffset()
}

func (m *Model) exitDetailMode() {
	if m.section == sectionCatalog {
		m.exitSourceDetail()
	}
}

func (m *Model) beginCreateInput() {
	if m.section == sectionSources {
		m.beginAddSource()
		return
	}

	if m.section == sectionProfiles {
		m.beginCreateProfile()
	}
}

func (m *Model) beginRemoveCurrentItem() {
	if m.section == sectionSources {
		m.beginRemoveSourceConfirm()
		return
	}

	if m.section == sectionProfiles {
		m.beginRemoveProfileConfirm()
	}
}

func (m *Model) switchTopLevelSection(direction int) {
	if len(topLevelSections) == 0 {
		return
	}

	if m.section == sectionCatalog {
		m.sourceCatalogCursor = m.cursor
		m.sourceCatalogOffset = m.offset
	}

	current := m.currentTopLevelSection()
	index := 0
	for candidateIndex, candidateSection := range topLevelSections {
		if candidateSection == current {
			index = candidateIndex
			break
		}
	}

	index = (index + direction + len(topLevelSections)) % len(topLevelSections)
	next := topLevelSections[index]
	if next == sectionSources && m.activeSourceID != "" {
		if _, ok := m.currentSource(); !ok {
			m.activeSourceID = ""
		} else {
			m.section = sectionCatalog
			m.cursor = m.sourceCatalogCursor
			m.offset = m.sourceCatalogOffset
			m.ensureOffset()
			return
		}
	}

	m.section = next
	if next == sectionSources {
		m.cursor = m.sourceListCursor
		m.offset = m.sourceListOffset
		m.ensureOffset()
		return
	}

	m.cursor = 0
	m.offset = 0
}

func (m Model) currentTopLevelSection() section {
	if m.section == sectionCatalog {
		return sectionSources
	}

	return m.section
}

func (m Model) currentSource() (source.Source, bool) {
	for _, configuredSource := range m.snapshot.Sources {
		if configuredSource.ID() == m.activeSourceID {
			return configuredSource, true
		}
	}

	return source.Source{}, false
}

func (m Model) sourceSkills(sourceID string) catalog.Skills {
	var discoveredSkills catalog.Skills
	for _, discoveredSkill := range m.snapshot.Catalog.Skills() {
		if discoveredSkill.SourceID() == sourceID {
			discoveredSkills = append(discoveredSkills, discoveredSkill)
		}
	}

	return discoveredSkills
}

func (m Model) skillCountForSource(sourceID string) int {
	count := 0
	for _, discoveredSkill := range m.snapshot.Catalog.Skills() {
		if discoveredSkill.SourceID() == sourceID {
			count++
		}
	}

	return count
}

func (m Model) currentCatalogSkill() (catalog.Skill, bool) {
	if m.section != sectionCatalog {
		return catalog.Skill{}, false
	}

	skills := m.sourceSkills(m.activeSourceID)
	if m.cursor < 0 || m.cursor >= len(skills) {
		return catalog.Skill{}, false
	}

	return skills[m.cursor], true
}

func (m *Model) ensureOffset() {
	items := m.currentItems()
	if len(items) == 0 {
		m.offset = 0
		return
	}

	listWidth := m.listViewportWidth()
	contentWidth := m.listContentWidth(listWidth)
	ranges := renderedItemRanges(items, contentWidth)
	if m.cursor >= len(ranges) {
		m.cursor = len(ranges) - 1
	}

	viewportLines := m.listViewportLineCount(listWidth)
	selected := ranges[m.cursor]
	visibleStart := max(m.offset, 0)
	visibleEnd := visibleStart + viewportLines - 1

	if selected.StartLine < visibleStart {
		m.offset = selected.StartLine
		return
	}

	if selected.EndLine > visibleEnd {
		m.offset = max(selected.EndLine-viewportLines+1, 0)
	}
}

func (m Model) visibleRange(items []item) (int, int) {
	return m.visibleRangeWithWidth(items, m.listViewportWidth())
}

func (m Model) visibleRangeWithWidth(items []item, width int) (int, int) {
	visibleLines := m.listViewportLineCount(width)
	if len(items) == 0 {
		return 0, 0
	}

	contentWidth := m.listContentWidth(width)
	ranges := renderedItemRanges(items, contentWidth)
	if len(ranges) == 0 {
		return 0, 0
	}

	visibleStart := max(m.offset, 0)
	visibleEnd := visibleStart + visibleLines - 1
	start := len(ranges)
	end := 0

	for index, lineRange := range ranges {
		if lineRange.EndLine < visibleStart {
			continue
		}
		if lineRange.StartLine > visibleEnd {
			break
		}
		if start == len(ranges) {
			start = index
		}
		end = index + 1
	}

	if start == len(ranges) {
		start = min(len(ranges)-1, max(0, m.cursor))
		end = start + 1
	}

	return start, end
}

func (m Model) visibleCount() int {
	items := m.currentItems()
	start, end := m.visibleRange(items)
	if end <= start {
		return 1
	}

	return end - start
}

func renderSkillDetail(discoveredSkill catalog.Skill) string {
	lines := make([]string, 0, 5)

	if discoveredSkill.Description() != "" {
		lines = append(lines, discoveredSkill.Description(), "")
	}

	lines = append(lines, "Source: "+discoveredSkill.SourceID())
	lines = append(lines, "Path: "+displayRelativePath(discoveredSkill.RelativePath()))
	lines = append(lines, "Entry: "+discoveredSkill.FilePath())

	return strings.Join(lines, "\n")
}

func renderSourceDetail(configuredSource source.Source, discoveredSkillCount int, selectedCount int) string {
	return strings.Join([]string{
		"URL: " + configuredSource.Locator(),
		fmt.Sprintf("Skills discovered: %d", discoveredSkillCount),
		fmt.Sprintf("Selected skills: %d", selectedCount),
		"",
		"Press enter to browse this source.",
	}, "\n")
}

func renderStatusDetail(snapshot Snapshot, summary selectionSummary) string {
	lastIndexed := "not indexed yet"
	if !snapshot.Catalog.IndexedAt().IsZero() {
		lastIndexed = snapshot.Catalog.IndexedAt().Local().Format("Mon Jan 2 15:04:05 MST 2006")
	}

	lines := []string{
		"Sources: " + snapshot.Runtime.SourcesFile,
		"Catalog: " + snapshot.Runtime.CatalogFile,
		"Profiles: " + snapshot.Runtime.ProfilesFile,
		"Sync state: " + snapshot.Runtime.SyncStateDir,
		"Active profile: " + snapshot.Profiles.ActiveName(),
		"Saved skills: " + fmt.Sprintf("%d", snapshot.ActiveProfile().SelectedCount()),
		"Draft skills: " + fmt.Sprintf("%d", summary.SelectedCount),
		"Draft changes: " + summary.PendingLabel(),
		"Synced skills: " + fmt.Sprintf("%d", len(snapshot.SyncedSelection)),
		"Last indexed: " + lastIndexed,
	}

	if len(snapshot.Warnings) > 0 {
		lines = append(lines, "", "Warnings:")
		for _, warning := range snapshot.Warnings {
			lines = append(lines, "- "+warning)
		}
	}

	return strings.Join(lines, "\n")
}

func displayRelativePath(relativePath string) string {
	if relativePath == "" {
		return "."
	}

	return relativePath
}

func (m Model) totalWidth() int {
	availableWidth := m.width - appStyle.GetHorizontalPadding()
	if availableWidth < 1 {
		return 1
	}

	return availableWidth
}

func (m Model) leftPaneWidth() int {
	totalWidth := m.totalWidth()
	if totalWidth <= 1 {
		return 1
	}

	leftWidth := max(totalWidth/3, 1)
	if leftWidth >= totalWidth {
		leftWidth = totalWidth - 1
	}
	if leftWidth < 1 {
		return totalWidth
	}

	return leftWidth
}

func (m Model) rightPaneWidth() int {
	totalWidth := m.totalWidth()
	rightWidth := totalWidth - m.leftPaneWidth() - 1
	if rightWidth < 1 {
		return 1
	}

	return rightWidth
}

func (m Model) listViewportWidth() int {
	viewportWidth := m.leftPaneWidth() - panelStyle.GetHorizontalPadding() - panelStyle.GetHorizontalBorderSize()
	if viewportWidth < 1 {
		return 1
	}

	return viewportWidth
}

func (m Model) listContentWidth(width int) int {
	if width <= 0 {
		width = m.listViewportWidth()
	}

	return max(1, width-2)
}

func (m Model) listViewportLineCount(width int) int {
	if m.bodyHeight() <= 0 {
		return listItemHeight
	}

	if width <= 0 {
		width = m.listViewportWidth()
	}

	availableLines := m.bodyHeight() - m.listChromeHeight(width)
	if availableLines < listItemHeight {
		return listItemHeight
	}

	return availableLines
}

func (m Model) listChromeHeight(width int) int {
	items := m.currentItems()
	control := controlBoxStyle.Width(max(1, width-2)).Render(strings.Join([]string{
		controlLabelStyle.Render(m.renderPaneLabel()),
		controlValueStyle.Render(m.renderSectionSummary(items)),
	}, "\n"))

	chrome := strings.Join([]string{
		panelTitleStyle.Render(strings.ToLower(m.paneTitle())),
		panelMetaStyle.Render(m.renderPaneMeta(items)),
		"",
		control,
		"",
	}, "\n")

	return lipgloss.Height(chrome)
}

func renderedItemHeight(item item, width int) int {
	return len(renderListItemRows(item, false, width))
}

func renderedItemRanges(items []item, width int) []itemRange {
	ranges := make([]itemRange, 0, len(items))
	currentLine := 0

	for index, entry := range items {
		itemHeight := renderedItemHeight(entry, width)
		if itemHeight <= 0 {
			itemHeight = 1
		}

		startLine := currentLine
		endLine := currentLine + itemHeight - 1
		ranges = append(ranges, itemRange{StartLine: startLine, EndLine: endLine})

		currentLine = endLine + 1
		if index < len(items)-1 {
			currentLine++
		}
	}

	return ranges
}

func newListViewport(width int, height int) viewport.Model {
	vp := viewport.New(
		viewport.WithWidth(max(1, width)),
		viewport.WithHeight(max(1, height)),
	)
	vp.FillHeight = true
	vp.MouseWheelEnabled = true

	return vp
}

func (m Model) headerHeight() int {
	return lipgloss.Height(headerPanelStyle.Width(m.totalWidth()).Render(m.renderHeader(max(1, m.totalWidth()-headerPanelStyle.GetHorizontalPadding()-headerPanelStyle.GetHorizontalBorderSize()))))
}

func (m Model) footerHeight() int {
	return lipgloss.Height(m.renderFooter())
}

func (m Model) panelInnerHeight(height int) int {
	if height < 1 {
		return 1
	}

	return height
}
