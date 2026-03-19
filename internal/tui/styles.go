package tui

import "charm.land/lipgloss/v2"

var (
	paletteBorder        = lipgloss.Color("141")
	paletteBorderMuted   = lipgloss.Color("237")
	paletteText          = lipgloss.Color("255")
	paletteMuted         = lipgloss.Color("240")
	paletteHeading       = lipgloss.Color("51")
	paletteHeadingAlt    = lipgloss.Color("183")
	paletteSelection     = lipgloss.Color("99")
	paletteSelectionText = lipgloss.Color("255")
	paletteChipBG        = lipgloss.Color("141")
	paletteChipText      = lipgloss.Color("183")
	paletteAccentPink    = lipgloss.Color("212")
	paletteWarning       = lipgloss.Color("226")

	frameBorder = lipgloss.NormalBorder()

	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	headerPanelStyle = lipgloss.NewStyle().
				Border(frameBorder).
				BorderForeground(paletteBorder).
				Padding(1, 1)

	logoStyle = lipgloss.NewStyle().
			Foreground(paletteHeading).
			Bold(true)

	logoTaglineStyle = lipgloss.NewStyle().
				Foreground(paletteAccentPink)

	headerStyle = lipgloss.NewStyle().
			Foreground(paletteText).
			Padding(0, 1).
			Bold(true)

	badgeStyle = lipgloss.NewStyle().
			Foreground(paletteText).
			Background(paletteAccentPink).
			Padding(0, 1)

	headerMetaLabelStyle = lipgloss.NewStyle().
				Foreground(paletteMuted)

	headerMetaValueStyle = lipgloss.NewStyle().
				Foreground(paletteHeadingAlt).
				Bold(true)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(paletteHeading).
			Background(paletteSelection).
			Padding(0, 1).
			Bold(true)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(paletteMuted).
				Padding(0, 1)

	panelStyle = lipgloss.NewStyle().
			Border(frameBorder).
			BorderForeground(paletteBorder).
			Padding(1, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(paletteHeading).
			Bold(true)

	panelMetaStyle = lipgloss.NewStyle().
			Foreground(paletteMuted)

	controlBoxStyle = lipgloss.NewStyle().
			Border(frameBorder).
			BorderForeground(paletteBorderMuted).
			Padding(0, 1)

	confirmBoxStyle = lipgloss.NewStyle().
			Border(frameBorder).
			BorderForeground(paletteAccentPink).
			Padding(1, 2).
			Width(72)

	confirmTitleStyle = lipgloss.NewStyle().
				Foreground(paletteAccentPink).
				Bold(true)

	confirmBodyStyle = lipgloss.NewStyle().
				Foreground(paletteText)

	confirmHelpStyle = lipgloss.NewStyle().
				Foreground(paletteMuted)

	controlLabelStyle = lipgloss.NewStyle().
				Foreground(paletteMuted)

	controlValueStyle = lipgloss.NewStyle().
				Foreground(paletteHeadingAlt).
				Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(paletteSelectionText).
				Background(paletteSelection).
				Padding(0, 1)

	selectedMetaStyle = lipgloss.NewStyle().
				Foreground(paletteSelectionText).
				Background(paletteSelection).
				Padding(0, 1)

	chosenItemStyle = lipgloss.NewStyle().
			Foreground(paletteAccentPink).
			Bold(true).
			Padding(0, 1)

	chosenMetaStyle = lipgloss.NewStyle().
			Foreground(paletteHeadingAlt).
			Padding(0, 1)

	chosenFocusedItemStyle = lipgloss.NewStyle().
				Foreground(paletteAccentPink).
				Background(paletteSelection).
				Bold(true).
				Padding(0, 1)

	chosenFocusedMetaStyle = lipgloss.NewStyle().
				Foreground(paletteText).
				Background(paletteSelection).
				Padding(0, 1)

	activeProfileItemStyle = lipgloss.NewStyle().
				Foreground(paletteHeading).
				Bold(true).
				Padding(0, 1)

	activeProfileMetaStyle = lipgloss.NewStyle().
				Foreground(paletteHeadingAlt).
				Padding(0, 1)

	activeProfileFocusedItemStyle = lipgloss.NewStyle().
					Foreground(paletteHeading).
					Background(paletteSelection).
					Bold(true).
					Padding(0, 1)

	activeProfileFocusedMetaStyle = lipgloss.NewStyle().
					Foreground(paletteText).
					Background(paletteSelection).
					Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(paletteText).
			Padding(0, 1)

	metaStyle = lipgloss.NewStyle().
			Foreground(paletteMuted).
			Padding(0, 1)

	detailTitleStyle = lipgloss.NewStyle().
				Foreground(paletteHeading).
				Bold(true)

	detailMetaStyle = lipgloss.NewStyle().
			Foreground(paletteMuted)

	sectionHeadingStyle = lipgloss.NewStyle().
				Foreground(paletteHeading).
				Bold(true)

	detailBodyStyle = lipgloss.NewStyle().
			Foreground(paletteText)

	chipStyle = lipgloss.NewStyle().
			Foreground(paletteChipText).
			Background(paletteChipBG).
			Padding(0, 1)

	dividerStyle = lipgloss.NewStyle().
			Foreground(paletteBorderMuted)

	footerRowStyle = lipgloss.NewStyle().
			PaddingTop(1)

	helpBoxStyle = lipgloss.NewStyle().
			Border(frameBorder).
			BorderForeground(paletteBorderMuted).
			Padding(0, 1)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(paletteHeading).
			Bold(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(paletteHeadingAlt).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(paletteMuted)

	helpDividerStyle = lipgloss.NewStyle().
				Foreground(paletteBorderMuted)

	footerStatusLabelStyle = lipgloss.NewStyle().
				Foreground(paletteMuted)

	footerStatusValueStyle = lipgloss.NewStyle().
				Foreground(paletteText).
				Bold(true)

	scrollTrackStyle = lipgloss.NewStyle().
				Foreground(paletteBorderMuted)

	scrollThumbStyle = lipgloss.NewStyle().
				Foreground(paletteBorder).
				Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(paletteWarning).
			Bold(true)
)
