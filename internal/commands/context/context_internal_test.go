package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sumup "github.com/sumup/sumup-go"
)

func TestModelUpdate(t *testing.T) {
	t.Run("ignores stale search results", func(t *testing.T) {
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

		assert.Nil(t, cmd)

		got, ok := updated.(model)
		require.True(t, ok)
		assert.True(t, got.loading)
		require.NotEmpty(t, got.displayed)
		assert.Equal(t, "current", got.displayed[0].Resource.ID)
	})

	t.Run("applies active search results", func(t *testing.T) {
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

		assert.Nil(t, cmd)

		got, ok := updated.(model)
		require.True(t, ok)
		assert.False(t, got.loading)
		require.Len(t, got.displayed, 1)
		assert.Equal(t, "fresh", got.displayed[0].Resource.ID)
	})
}
