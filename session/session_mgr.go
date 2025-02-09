package session

import (
	"fmt"
	"log/slog"
	"sync"
)

type NewSessionIOFunc func() (SessionIO, error)

type SessionManager struct {
	opt      *options
	log      *slog.Logger
	sessions map[string]*Session
	lock     sync.Mutex
}

func NewSessionManager(optfs ...OptionFunc) *SessionManager {
	opt := newOptions(optfs...)
	return &SessionManager{
		opt:      opt,
		sessions: make(map[string]*Session),
		log:      slog.New(opt.logHandler).With("module", "webtty/session"),
	}
}

// GetSession returns a session by id. If the session does not exist, it will create a new session
func (mgr *SessionManager) GetSession(id string, f NewSessionIOFunc) (*Session, error) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	if session, exist := mgr.sessions[id]; exist {
		mgr.log.With("sid", id).Info("session already exist")
		return session, nil
	}

	sio, err := f()
	if err != nil {
		return nil, fmt.Errorf("failed to create session io: %w", err)
	}

	session := NewSession(id, sio, mgr.log)

	mgr.sessions[id] = session
	mgr.log.With("sid", id).Info("session created")
	return session, nil
}

// HasSession checks if a session exists by id
func (mgr *SessionManager) HasSession(id string) bool {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	_, exist := mgr.sessions[id]
	return exist
}

// RemoveSession removes a session by id
func (mgr *SessionManager) RemoveSession(id string) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	if session, exist := mgr.sessions[id]; exist {
		session.Close()
		delete(mgr.sessions, id)
		mgr.log.With("sid", id).Info("session removed")
	} else {
		mgr.log.With("sid", id).Warn("session not found")
	}
}
