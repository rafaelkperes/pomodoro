package pomodoro

import (
	"fmt"
	"log"
	"time"
)

type State string

const (
	stateNew      State = "new"
	stateFocusing State = "focusing"
	stateOnBreak  State = "breaking"
	stateInvalid  State = "invalid"
)

type Hook func(_ State) error

type Session struct {
	focusFor       time.Duration
	breakFor       time.Duration
	state          State
	stateTimestamp time.Time
	timer          *time.Timer
	hooks          []Hook
}

func NewSession(focusFor, breakFor time.Duration) *Session {
	s := &Session{
		focusFor:       focusFor,
		breakFor:       breakFor,
		state:          stateNew,
		stateTimestamp: time.Now(),
		hooks:          []Hook{},
	}
	return s
}

func (s *Session) State() State {
	return s.state
}

func (s *Session) AddHook(hook Hook) {
	s.hooks = append(s.hooks, hook)
}

func (s *Session) Start() error {
	if s.state != stateNew {
		return fmt.Errorf("from %q to %q: %w", s.state, stateFocusing, ErrInvalidStateTransition)
	}
	// Start session
	s.focus()
	return nil
}

// focus starts a focus subsession.
// The previous timer from the session is discarded.
func (s *Session) focus() {
	s.state = stateFocusing
	s.stateTimestamp = time.Now()
	s.fireHooks()
	// Add next transition
	s.timer = time.AfterFunc(s.focusFor, s.breakk)
}

// breakk starts a break subsession.
// The previous timer from the session is discarded.
func (s *Session) breakk() {
	s.state = stateOnBreak
	s.stateTimestamp = time.Now()
	s.fireHooks()
	// Add next transition
	s.timer = time.AfterFunc(s.breakFor, s.focus)
}

// teardown clears session resources.
func (s *Session) Teardown() {
	// Stop and drain channel
	if !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
	s.state = stateInvalid
}

// fireHooks calls as hooks for the current state.
// Errors from hooks are simply logged.
// A separate function with parameters is used to avoid race conditions with
// either the current state or the current slice of hooks.
func (s *Session) fireHooks() {
	go fireHooks(s.hooks, s.state)
}

func fireHooks(hooks []Hook, state State) {
	for _, hook := range hooks {
		if err := hook(state); err != nil {
			log.Printf("hook returned error for state %q: %v\n", state, err)
		}
	}
}
