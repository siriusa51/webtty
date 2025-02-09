package session

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
)

type mockSessionIO struct {
	buff   bytes.Buffer
	close  bool
	width  int
	height int
	cancel context.CancelFunc
	ctx    context.Context
}

func (m *mockSessionIO) Done() <-chan struct{} {
	return m.ctx.Done()
}

func newMockSessionIO() *mockSessionIO {
	ctx, cancel := context.WithCancel(context.Background())
	return &mockSessionIO{
		buff:   bytes.Buffer{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (m *mockSessionIO) Read(p []byte) (n int, err error) {
	return m.buff.Read(p)
}

func (m *mockSessionIO) Write(p []byte) (n int, err error) {
	return m.buff.Write(p)
}

func (m *mockSessionIO) Close() error {
	m.close = true
	m.cancel()
	return nil
}

func (m *mockSessionIO) ResizeWindow(width, height int) error {
	m.width = width
	m.height = height
	return nil
}

func newMockSession(id string, sio SessionIO) *Session {
	return NewSession(id, sio, slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
}

func TestSession_Close(t *testing.T) {
	t.Run("test Close()", func(t *testing.T) {
		sess := newMockSession("test", newMockSessionIO())
		assert.NoError(t, sess.Close())
		assert.True(t, sess.sio.(*mockSessionIO).close)
	})
}

func TestSession_GetId(t *testing.T) {
	t.Run("test GetId()", func(t *testing.T) {
		sess := newMockSession("test", newMockSessionIO())
		assert.Equal(t, "test", sess.GetId())
	})
}

func TestSession_ResizeWindow(t *testing.T) {
	t.Run("test ResizeWindow()", func(t *testing.T) {
		sess := newMockSession("test", newMockSessionIO())
		assert.NoError(t, sess.ResizeWindow(80, 24))
		assert.Equal(t, 80, sess.sio.(*mockSessionIO).width)
		assert.Equal(t, 24, sess.sio.(*mockSessionIO).height)
	})
}

func TestSession_Occupied(t *testing.T) {
	t.Run("test Occupied()", func(t *testing.T) {
		sess := newMockSession("test", newMockSessionIO())
		assert.False(t, sess.Occupied())

		err := sess.Occupy()
		assert.NoError(t, err)
		assert.True(t, sess.Occupied())

		err = sess.Occupy()
		assert.Error(t, err)

	})
}

func TestSession_Release(t *testing.T) {
	t.Run("test Release()", func(t *testing.T) {
		sess := newMockSession("test", newMockSessionIO())
		assert.False(t, sess.Occupied())

		sess.Release()
		assert.False(t, sess.Occupied())

		err := sess.Occupy()
		assert.NoError(t, err)

		sess.Release()
		assert.False(t, sess.Occupied())

		sess.Release()
		assert.False(t, sess.Occupied())
	})
}

func TestSession_ReadWrite(t *testing.T) {
	t.Run("test Read()/Write()", func(t *testing.T) {
		sess := newMockSession("test", newMockSessionIO())
		defer sess.Close()
		n, err := sess.Write([]byte("hello world"))
		assert.NoError(t, err)
		assert.Equal(t, 11, n)

		buff := make([]byte, 1024)
		n, err = sess.Read(buff)
		assert.NoError(t, err)
		assert.Equal(t, 11, n)

		sess.Write([]byte("hello world"))
		buff = make([]byte, 1)
		n, err = sess.Read(buff)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Equal(t, "h", string(buff[:n]))

		n, err = sess.Read(buff)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Equal(t, "e", string(buff[:n]))

		buff = make([]byte, 1024)
		n, err = sess.Read(buff)
		assert.NoError(t, err)
		assert.Equal(t, 9, n)
		assert.Equal(t, "llo world", string(buff[:n]))
	})
}
