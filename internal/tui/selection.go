package tui

import (
	"fmt"

	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
)

type selectionSummary struct {
	SelectedCount   int
	PendingAddCount int
	PendingDelCount int
}

func (s selectionSummary) Dirty() bool {
	return s.PendingAddCount > 0 || s.PendingDelCount > 0
}

func (s selectionSummary) PendingLabel() string {
	if !s.Dirty() {
		return "saved"
	}

	return fmt.Sprintf("+%d -%d unsaved", s.PendingAddCount, s.PendingDelCount)
}

func (s selectionSummary) SelectionLabel() string {
	return fmt.Sprintf("%d selected", s.SelectedCount)
}

type draftSelection struct {
	baseline map[string]skill_identity.Identity
	desired  map[string]skill_identity.Identity
}

func newDraftSelection(baseline skill_identity.Identities, desired skill_identity.Identities) draftSelection {
	return draftSelection{
		baseline: identityMap(baseline),
		desired:  identityMap(desired),
	}
}

func initialDesiredSelection(snapshot Snapshot) skill_identity.Identities {
	return snapshot.ActiveSelection()
}

func (s draftSelection) Summary() selectionSummary {
	var summary selectionSummary

	summary.SelectedCount = len(s.desired)

	for key := range s.desired {
		if _, ok := s.baseline[key]; !ok {
			summary.PendingAddCount++
		}
	}

	for key := range s.baseline {
		if _, ok := s.desired[key]; !ok {
			summary.PendingDelCount++
		}
	}

	return summary
}

func (s draftSelection) Wants(identity skill_identity.Identity) bool {
	_, ok := s.desired[identity.Key()]
	return ok
}

func (s draftSelection) Desired() skill_identity.Identities {
	return identitiesFromMap(s.desired)
}

func (s *draftSelection) Toggle(identity skill_identity.Identity) {
	key := identity.Key()
	if _, ok := s.desired[key]; ok {
		delete(s.desired, key)
		return
	}

	s.desired[key] = identity
}

func (s *draftSelection) ReplaceBaseline(baseline skill_identity.Identities) {
	s.baseline = identityMap(baseline)
}

func (s *draftSelection) Reset(baseline skill_identity.Identities) {
	s.baseline = identityMap(baseline)
	s.desired = identityMap(baseline)
}

func (s draftSelection) DesiredCountForSource(sourceID string) int {
	count := 0
	for _, identity := range s.desired {
		if identity.SourceID() == sourceID {
			count++
		}
	}

	return count
}

func identityMap(identities skill_identity.Identities) map[string]skill_identity.Identity {
	result := make(map[string]skill_identity.Identity, len(identities))
	for _, identity := range skill_identity.NewIdentities(identities...) {
		result[identity.Key()] = identity
	}

	return result
}

func identitiesFromMap(identities map[string]skill_identity.Identity) skill_identity.Identities {
	result := make(skill_identity.Identities, 0, len(identities))
	for _, identity := range identities {
		result = append(result, identity)
	}

	return skill_identity.NewIdentities(result...)
}

func (m Model) selectionSummary() selectionSummary {
	return m.selection.Summary()
}

func (m Model) isSelectedSkill(discoveredSkill catalog.Skill) bool {
	return m.selection.Wants(discoveredSkill.Identity())
}

func (m *Model) toggleCurrentCatalogSkill() {
	discoveredSkill, ok := m.currentCatalogSkill()
	if !ok {
		return
	}

	m.selection.Toggle(discoveredSkill.Identity())

	m.statusMessage = "Selection updated"
	if summary := m.selectionSummary(); summary.Dirty() {
		m.statusMessage += " • " + summary.PendingLabel()
	}
}

func (m *Model) applySnapshot(snapshot Snapshot) {
	m.snapshot = snapshot
	m.selection.ReplaceBaseline(snapshot.ActiveSelection())
}

func (m *Model) resetSelection(snapshot Snapshot) {
	m.snapshot = snapshot
	m.selection.Reset(snapshot.ActiveSelection())
}
