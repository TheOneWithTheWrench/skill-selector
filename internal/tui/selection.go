package tui

import (
	"fmt"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
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
	baseline map[string]skillidentity.Identity
	desired  map[string]skillidentity.Identity
}

func newDraftSelection(baseline skillidentity.Identities, desired skillidentity.Identities) draftSelection {
	return draftSelection{
		baseline: identityMap(baseline),
		desired:  identityMap(desired),
	}
}

func initialDesiredSelection(snapshot Snapshot) skillidentity.Identities {
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

func (s draftSelection) Wants(identity skillidentity.Identity) bool {
	_, ok := s.desired[identity.Key()]
	return ok
}

func (s draftSelection) Desired() skillidentity.Identities {
	return identitiesFromMap(s.desired)
}

func (s *draftSelection) Toggle(identity skillidentity.Identity) {
	key := identity.Key()
	if _, ok := s.desired[key]; ok {
		delete(s.desired, key)
		return
	}

	s.desired[key] = identity
}

func (s *draftSelection) ReplaceBaseline(baseline skillidentity.Identities) {
	s.baseline = identityMap(baseline)
}

func (s *draftSelection) Reset(baseline skillidentity.Identities) {
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

func identityMap(identities skillidentity.Identities) map[string]skillidentity.Identity {
	result := make(map[string]skillidentity.Identity, len(identities))
	for _, identity := range skillidentity.NewIdentities(identities...) {
		result[identity.Key()] = identity
	}

	return result
}

func identitiesFromMap(identities map[string]skillidentity.Identity) skillidentity.Identities {
	result := make(skillidentity.Identities, 0, len(identities))
	for _, identity := range identities {
		result = append(result, identity)
	}

	return skillidentity.NewIdentities(result...)
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
