package context

import (
	"testing"

	sumup "github.com/sumup/sumup-go"
)

func TestModelUpdateIgnoresStaleSearchResult(t *testing.T) {
	currentItems := []sumup.Membership{{Resource: sumup.MembershipResource{ID: "current"}}}
	m := model{
		currentLevel:   navigationLevel{memberships: currentItems},
		displayed:      currentItems,
		loading:        true,
		activeSearchID: 2,
	}

	updated, cmd := m.Update(searchResultMsg{
		requestID:   1,
		memberships: []sumup.Membership{{Resource: sumup.MembershipResource{ID: "stale"}}},
	})
	if cmd != nil {
		t.Fatalf("Update() cmd = %v, want nil", cmd)
	}

	got := updated.(model)
	if !got.loading {
		t.Fatal("Update() loading = false, want true for ignored stale result")
	}
	if got.displayed[0].Resource.ID != "current" {
		t.Fatalf("Update() displayed = %q, want current result to remain", got.displayed[0].Resource.ID)
	}
}

func TestModelUpdateAppliesActiveSearchResult(t *testing.T) {
	m := model{
		currentLevel:   navigationLevel{},
		displayed:      nil,
		loading:        true,
		activeSearchID: 3,
	}

	updated, cmd := m.Update(searchResultMsg{
		requestID:   3,
		memberships: []sumup.Membership{{Resource: sumup.MembershipResource{ID: "fresh"}}},
	})
	if cmd != nil {
		t.Fatalf("Update() cmd = %v, want nil", cmd)
	}

	got := updated.(model)
	if got.loading {
		t.Fatal("Update() loading = true, want false")
	}
	if len(got.displayed) != 1 || got.displayed[0].Resource.ID != "fresh" {
		t.Fatalf("Update() displayed = %+v, want active search result", got.displayed)
	}
}
