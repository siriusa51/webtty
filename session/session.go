package session

import (
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type SessionIO interface {
	io.ReadWriteCloser
	Done() <-chan struct{}
	ResizeWindow(width, height int) error
}

type Message struct {
	Data  []byte
	Error error
}

type Session struct {
	id     string
	occupy bool
	sio    SessionIO
	lock   sync.Mutex

	log *slog.Logger
}

func NewSession(id string, sio SessionIO, log *slog.Logger) *Session {
	sess := &Session{
		id:  id,
		sio: sio,
		log: log.With("sid", id),
	}

	return sess
}

// GetId returns the session id.
func (s *Session) GetId() string {
	return s.id
}

// Close closes the session.
func (s *Session) Close() error {
	return s.sio.Close()
}

// Read reads data from the session.
func (s *Session) Read(buff []byte) (int, error) {
	return s.sio.Read(buff)
}

// Write writes data to the session.
func (s *Session) Write(buff []byte) (int, error) {
	return s.sio.Write(buff)
}

// ResizeWindow resizes the window of the session.
func (s *Session) ResizeWindow(width, height int) error {
	return s.sio.ResizeWindow(width, height)
}

func (s *Session) Done() <-chan struct{} {
	return s.sio.Done()
}

// Occupy returns true if the session is occupied.
func (s *Session) Occupied() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.occupy
}

// Occupy occupies the session, preventing other users from using it.
func (s *Session) Occupy() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.occupy {
		return fmt.Errorf("session is occupied")
	}

	s.log.Info("occupy session")
	s.occupy = true
	return nil
}

// Release releases the session.
func (s *Session) Release() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.occupy {
		return
	}

	s.log.Info("release session")
	s.occupy = false
}
