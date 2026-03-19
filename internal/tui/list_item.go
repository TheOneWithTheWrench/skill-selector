package tui

import "charm.land/lipgloss/v2"

type listItemVisualState int

const (
	listItemNormal listItemVisualState = iota
	listItemSelected
	listItemActive
	listItemFocused
	listItemFocusedSelected
	listItemFocusedActive
)

func listItemStyles(item item, focused bool, width int) (lipgloss.Style, lipgloss.Style) {
	titleStyle := itemStyle.Width(width)
	subtitleStyle := metaStyle.Width(width)

	switch resolveListItemVisualState(item, focused) {
	case listItemSelected:
		titleStyle = chosenItemStyle.Width(width)
		subtitleStyle = chosenMetaStyle.Width(width)
	case listItemActive:
		titleStyle = activeProfileItemStyle.Width(width)
		subtitleStyle = activeProfileMetaStyle.Width(width)
	case listItemFocused:
		titleStyle = selectedItemStyle.Width(width)
		subtitleStyle = selectedMetaStyle.Width(width)
	case listItemFocusedSelected:
		titleStyle = chosenFocusedItemStyle.Width(width)
		subtitleStyle = chosenFocusedMetaStyle.Width(width)
	case listItemFocusedActive:
		titleStyle = activeProfileFocusedItemStyle.Width(width)
		subtitleStyle = activeProfileFocusedMetaStyle.Width(width)
	}

	return titleStyle, subtitleStyle
}

func resolveListItemVisualState(item item, focused bool) listItemVisualState {
	if item.Active && focused {
		return listItemFocusedActive
	}

	if item.Selectable && item.Selected && focused {
		return listItemFocusedSelected
	}

	if focused {
		return listItemFocused
	}

	if item.Active {
		return listItemActive
	}

	if item.Selectable && item.Selected {
		return listItemSelected
	}

	return listItemNormal
}
