package session

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionManager(t *testing.T) {
	t.Run("test normal", func(t *testing.T) {
		f := func() (SessionIO, error) {
			return newMockSessionIO(), nil
		}
		mgr := NewSessionManager()
		sess1, err := mgr.GetSession("sess1", f)
		assert.NoError(t, err)
		assert.Equal(t, "sess1", sess1.GetId())

		exist := mgr.HasSession("sess1")
		assert.True(t, exist)

		temp, err := mgr.GetSession("sess1", f)
		assert.NoError(t, err)
		assert.Equal(t, sess1, temp)

		mgr.RemoveSession("sess1")
		exist = mgr.HasSession("sess1")
		assert.False(t, exist)
	})

	t.Run("test new session error", func(t *testing.T) {
		f := func() (SessionIO, error) {
			return nil, assert.AnError
		}

		mgr := NewSessionManager()
		_, err := mgr.GetSession("sess1", f)
		assert.Error(t, err)

		exist := mgr.HasSession("sess1")
		assert.False(t, exist)
	})
}
