package tui

import (
	"fmt"
	"maps"
	"strings"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
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
		return "in sync"
	}

	return fmt.Sprintf("+%d -%d pending sync", s.PendingAddCount, s.PendingDelCount)
}

func (s selectionSummary) SelectionLabel() string {
	return fmt.Sprintf("%d selected", s.SelectedCount)
}

type draftSelection struct {
	active  map[string]skillidentity.Identity
	desired map[string]skillidentity.Identity
}

func newDraftSelection(active skillidentity.Identities, desired skillidentity.Identities) draftSelection {
	return draftSelection{
		active:  identityMap(active),
		desired: identityMap(desired),
	}
}

func initialDesiredSelection(snapshot Snapshot) skillidentity.Identities {
	configuredSourceIDs := make(map[string]struct{}, len(snapshot.Sources))
	for _, configuredSource := range snapshot.Sources {
		configuredSourceIDs[configuredSource.ID()] = struct{}{}
	}

	var desired skillidentity.Identities
	for _, identity := range snapshot.ActiveSelection {
		if _, ok := configuredSourceIDs[identity.SourceID()]; !ok {
			continue
		}

		desired = append(desired, identity)
	}

	return skillidentity.NewIdentities(desired...)
}

func (s draftSelection) Summary() selectionSummary {
	var summary selectionSummary

	summary.SelectedCount = len(s.desired)

	for key := range s.desired {
		if _, ok := s.active[key]; !ok {
			summary.PendingAddCount++
		}
	}

	for key := range s.active {
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

func (s *draftSelection) ReplaceActive(active skillidentity.Identities) {
	s.active = identityMap(active)
}

func (s *draftSelection) RemoveSource(sourceID string) {
	for key, identity := range s.desired {
		if identity.SourceID() == strings.TrimSpace(sourceID) {
			delete(s.desired, key)
		}
	}
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

func cloneIdentityMap(identities map[string]skillidentity.Identity) map[string]skillidentity.Identity {
	cloned := make(map[string]skillidentity.Identity, len(identities))
	maps.Copy(cloned, identities)

	return cloned
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
	m.selection.ReplaceActive(snapshot.ActiveSelection)
}

func configuredSourceIDs(configuredSources source.Sources) map[string]struct{} {
	result := make(map[string]struct{}, len(configuredSources))
	for _, configuredSource := range configuredSources {
		result[configuredSource.ID()] = struct{}{}
	}

	return result
}
