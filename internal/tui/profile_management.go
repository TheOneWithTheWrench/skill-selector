package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
)

type profileActionMsg struct {
	result ProfilesActionResult
	err    error
}

type profileEditMsg struct {
	mode   profileInputMode
	name   string
	result ProfilesActionResult
	err    error
}

type profileRemovedMsg struct {
	name   string
	result ProfilesActionResult
	err    error
}

func (m *Model) handleProfileInput(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	if !m.profileInputActive {
		return *m, nil, false
	}

	switch msg.String() {
	case "ctrl+c":
		return *m, nil, false
	case "esc":
		m.cancelProfileInput("Profile edit cancelled")
		return *m, nil, true
	case "enter":
		cmd := m.startProfileEdit()
		return *m, cmd, true
	case "backspace":
		m.profileInputValue = trimLastRune(m.profileInputValue)
		return *m, nil, true
	}

	if msg.Text != "" {
		m.profileInputValue += sanitizeSourceInput(msg.Text)
		return *m, nil, true
	}

	return *m, nil, true
}

func (m *Model) handleProfilePaste(msg tea.PasteMsg) (Model, tea.Cmd, bool) {
	if !m.profileInputActive {
		return *m, nil, false
	}

	m.profileInputValue += sanitizeSourceInput(msg.String())

	return *m, nil, true
}

func (m *Model) beginCreateProfile() {
	if m.section != sectionProfiles {
		return
	}

	m.profileInputActive = true
	m.profileInputMode = profileInputCreate
	m.profileInputValue = ""
	m.profileToRename = ""
	m.statusMessage = "Enter profile name and press enter"
}

func (m *Model) beginRenameCurrentProfile() {
	if m.section != sectionProfiles {
		return
	}

	currentProfile, ok := m.currentProfileItem()
	if !ok {
		return
	}
	if currentProfile.Name() == profile.DefaultName {
		m.statusMessage = "Default profile cannot be renamed"
		return
	}
	if m.workflow == nil {
		m.statusMessage = "Profile rename is unavailable"
		return
	}

	m.profileInputActive = true
	m.profileInputMode = profileInputRename
	m.profileInputValue = currentProfile.Name()
	m.profileToRename = currentProfile.Name()
	m.statusMessage = "Edit profile name and press enter"
}

func (m *Model) startProfileEdit() tea.Cmd {
	name := strings.TrimSpace(m.profileInputValue)
	if name == "" {
		m.statusMessage = "Profile name required"
		return nil
	}

	mode := m.profileInputMode
	oldName := m.profileToRename

	switch mode {
	case profileInputCreate:
		if m.workflow == nil || m.syncing || m.refreshing {
			m.statusMessage = "Profile create is unavailable"
			return nil
		}

		m.profileInputActive = false
		m.syncing = true
		m.statusMessage = "Creating profile"

		return tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				result, err := m.workflow.CreateProfile(context.Background(), name)
				return profileEditMsg{mode: mode, name: name, result: result, err: err}
			},
		)
	case profileInputRename:
		if m.workflow == nil || m.syncing || m.refreshing {
			m.statusMessage = "Profile rename is unavailable"
			return nil
		}

		m.profileInputActive = false
		m.syncing = true
		m.statusMessage = "Renaming profile"

		return tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				result, err := m.workflow.RenameProfile(context.Background(), oldName, name)
				return profileEditMsg{mode: mode, name: name, result: result, err: err}
			},
		)
	}

	return nil
}

func (m *Model) finishProfileEdit(msg profileEditMsg) {
	m.syncing = false
	m.profileInputValue = ""
	m.profileToRename = ""
	m.profileInputMode = profileInputCreate

	if msg.result.Snapshot != nil {
		m.applySnapshot(*msg.result.Snapshot)
		m.setProfileCursor(msg.name)
		m.ensureOffset()
	}

	if msg.err != nil {
		actionName := "Create"
		if msg.mode == profileInputRename {
			actionName = "Rename"
		}
		if msg.result.Summary != "" {
			m.statusMessage = fmt.Sprintf("%s • error: %v", msg.result.Summary, msg.err)
			return
		}

		m.statusMessage = fmt.Sprintf("%s profile failed: %v", actionName, msg.err)
		return
	}

	if msg.result.Summary != "" {
		m.statusMessage = msg.result.Summary
		return
	}

	m.statusMessage = "Profile updated"
}

func (m *Model) cancelProfileInput(message string) {
	m.profileInputActive = false
	m.profileInputValue = ""
	m.profileToRename = ""
	m.profileInputMode = profileInputCreate
	m.statusMessage = message
}

func (m *Model) startSwitchProfile() tea.Cmd {
	if m.section != sectionProfiles {
		return nil
	}

	currentProfile, ok := m.currentProfileItem()
	if !ok {
		return nil
	}
	if currentProfile.Name() == m.snapshot.Profiles.ActiveName() {
		m.statusMessage = "Profile already active"
		return nil
	}
	if m.workflow == nil || m.syncing || m.refreshing {
		m.statusMessage = "Profile switch is unavailable"
		return nil
	}

	m.syncing = true
	m.statusMessage = "Switching profile"

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.workflow.SwitchProfile(context.Background(), currentProfile.Name())
			return profileActionMsg{result: result, err: err}
		},
	)
}

func (m *Model) finishProfileSwitch(msg profileActionMsg) {
	m.syncing = false
	if msg.result.Snapshot != nil {
		m.resetSelection(*msg.result.Snapshot)
		m.setProfileCursor(m.snapshot.Profiles.ActiveName())
		m.ensureOffset()
	}

	if msg.err != nil {
		if msg.result.Summary != "" {
			m.statusMessage = fmt.Sprintf("%s • error: %v", msg.result.Summary, msg.err)
			return
		}

		m.statusMessage = fmt.Sprintf("Switch profile failed: %v", msg.err)
		return
	}

	m.statusMessage = msg.result.Summary
}

func (m *Model) beginRemoveProfileConfirm() {
	if m.section != sectionProfiles {
		return
	}

	currentProfile, ok := m.currentProfileItem()
	if !ok {
		return
	}
	if currentProfile.Name() == profile.DefaultName {
		m.statusMessage = "Default profile cannot be removed"
		return
	}
	if currentProfile.Name() == m.snapshot.Profiles.ActiveName() {
		m.statusMessage = "Switch away from the active profile before removing it"
		return
	}
	if m.workflow == nil {
		m.statusMessage = "Profile remove is unavailable"
		return
	}

	m.profileRemoveConfirmActive = true
	m.profileToRemove = currentProfile
	m.statusMessage = "Confirm profile removal"
}

func (m *Model) handleProfileRemoveConfirm(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	if !m.profileRemoveConfirmActive {
		return *m, nil, false
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return *m, nil, false
	case "y", "enter":
		cmd := m.startRemoveProfile()
		return *m, cmd, true
	case "n", "esc", "backspace":
		m.cancelRemoveProfile()
		return *m, nil, true
	default:
		return *m, nil, true
	}
}

func (m *Model) startRemoveProfile() tea.Cmd {
	if !m.profileRemoveConfirmActive {
		return nil
	}
	if m.workflow == nil || m.syncing || m.refreshing {
		m.statusMessage = "Profile remove is unavailable"
		return nil
	}

	currentProfile := m.profileToRemove
	m.profileRemoveConfirmActive = false
	m.profileToRemove = profile.Profile{}
	m.syncing = true
	m.statusMessage = "Removing profile"

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.workflow.RemoveProfile(context.Background(), currentProfile.Name())
			return profileRemovedMsg{name: currentProfile.Name(), result: result, err: err}
		},
	)
}

func (m *Model) finishRemoveProfile(msg profileRemovedMsg) {
	m.syncing = false
	if msg.result.Snapshot != nil {
		m.applySnapshot(*msg.result.Snapshot)
		if m.cursor >= len(m.snapshot.Profiles.All()) {
			m.cursor = max(0, len(m.snapshot.Profiles.All())-1)
		}
		m.ensureOffset()
	}

	if msg.err != nil {
		if msg.result.Summary != "" {
			m.statusMessage = fmt.Sprintf("%s • error: %v", msg.result.Summary, msg.err)
			return
		}

		m.statusMessage = fmt.Sprintf("Remove profile failed: %v", msg.err)
		return
	}

	m.statusMessage = msg.result.Summary
}

func (m *Model) cancelRemoveProfile() {
	m.profileRemoveConfirmActive = false
	m.profileToRemove = profile.Profile{}
	m.statusMessage = "Profile removal cancelled"
}

func (m *Model) currentProfileItem() (profile.Profile, bool) {
	profiles := m.snapshot.Profiles.All()
	if m.section != sectionProfiles || m.cursor < 0 || m.cursor >= len(profiles) {
		return profile.Profile{}, false
	}

	return profiles[m.cursor], true
}

func (m *Model) setProfileCursor(name string) {
	for index, currentProfile := range m.snapshot.Profiles.All() {
		if currentProfile.Name() == normalizeProfileName(name) {
			m.cursor = index
			return
		}
	}
	for index, currentProfile := range m.snapshot.Profiles.All() {
		if currentProfile.Name() == m.snapshot.Profiles.ActiveName() {
			m.cursor = index
			return
		}
	}
}

func renderProfileDetail(currentProfile profile.Profile, active bool) string {
	lines := []string{fmt.Sprintf("Saved skills: %d", currentProfile.SelectedCount())}
	if active {
		lines = append(lines, "", "This is the active profile.", "Draft changes stay in the TUI until you sync.")
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "", "Press space to switch to this profile.")
	return strings.Join(lines, "\n")
}
