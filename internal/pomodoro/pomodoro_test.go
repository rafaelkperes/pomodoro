package pomodoro

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewSession(t *testing.T) {
	t.Parallel()
	rq := require.New(t)
	sess := NewSession(time.Minute, time.Second)
	rq.Equal(stateNew, sess.State())
	rq.Equal(time.Minute, sess.focusFor)
	rq.Equal(time.Second, sess.breakFor)
}

func TestStartAndTeardownSession(t *testing.T) {
	t.Parallel()
	t.Run("FromNewSession", func(t *testing.T) {
		t.Parallel()
		rq := require.New(t)
		sess := NewSession(time.Minute, time.Minute)
		rq.NoError(sess.Start())
		rq.Equal(stateFocusing, sess.State())
		sess.Teardown()
		rq.Equal(stateInvalid, sess.State())
	})
	t.Run("AlreadyStartedSession", func(t *testing.T) {
		t.Parallel()
		rq := require.New(t)
		sess := NewSession(time.Minute, time.Minute)
		rq.NoError(sess.Start())
		rq.Error(sess.Start())
		sess.Teardown()
	})
	t.Run("FromInvalidSession", func(t *testing.T) {
		t.Parallel()
		rq := require.New(t)
		sess := NewSession(time.Minute, time.Minute)
		rq.NoError(sess.Start())
		sess.Teardown()
		rq.Error(sess.Start())
		rq.Equal(stateInvalid, sess.State())
		sess.Teardown()
		rq.Equal(stateInvalid, sess.State())
	})
}

func TestSessionStateTransition(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		focusDuration   time.Duration
		onBreakDuration time.Duration
	}{
		"Focus=Break": {
			focusDuration:   time.Second / 4,
			onBreakDuration: time.Second / 4,
		},
		"Focus>Break": {
			focusDuration:   time.Second / 2,
			onBreakDuration: time.Second / 4,
		},
		"Focus<Break": {
			focusDuration:   time.Second / 4,
			onBreakDuration: time.Second / 2,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rq := require.New(t)
			sess := NewSession(test.focusDuration, test.onBreakDuration)
			rq.NoError(sess.Start())
			rq.Equal(stateFocusing, sess.State())
			// Use 100ms as a threshold for checks
			time.Sleep(100 * time.Millisecond)
			// Still focusing
			time.Sleep(test.focusDuration / 2)
			rq.Equal(stateFocusing, sess.State())
			// Transition to on-break
			time.Sleep(test.focusDuration / 2)
			rq.Equal(stateOnBreak, sess.State())
			// Still on break
			time.Sleep(test.onBreakDuration / 2)
			rq.Equal(stateOnBreak, sess.State())
			// Transition back to focus
			time.Sleep(test.onBreakDuration / 2)
			rq.Equal(stateFocusing, sess.State())
			sess.Teardown()
		})
	}
}

func TestSessionHooks(t *testing.T) {
	t.Parallel()
	rq := require.New(t)
	duration := time.Second / 8
	sess := NewSession(duration, duration)
	// Add hooks
	hookCh := make(chan State)
	hook := func(state State) error {
		hookCh <- state
		return nil
	}
	sess.AddHook(Hook(hook))
	sess.AddHook(Hook(hook))
	rq.NoError(sess.Start())
	defer sess.Teardown()
	rq.Equal(stateFocusing, <-hookCh)
	rq.Equal(stateFocusing, <-hookCh)
	// Use 100ms as a threshold for checks
	time.Sleep(100 * time.Millisecond)
	select {
	case <-hookCh:
		t.Fatal("unexpected hook triggered")
	default:
	}
	// Wait for break
	time.Sleep(duration)
	rq.Equal(stateOnBreak, <-hookCh)
	rq.Equal(stateOnBreak, <-hookCh)
	// Wait for focus session
	time.Sleep(duration)
	rq.Equal(stateFocusing, <-hookCh)
	rq.Equal(stateFocusing, <-hookCh)
	// Wait for break
	time.Sleep(duration)
	rq.Equal(stateOnBreak, <-hookCh)
	rq.Equal(stateOnBreak, <-hookCh)
}
