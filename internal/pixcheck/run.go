package pixcheck

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// The Ebitengine session shared by all calls to [Run]. Each call hands its
// function over through jobs and waits for it to be done.
var (
	startOnce      sync.Once
	jobs           = make(chan *job)
	sessionStopped = make(chan struct{})
	sessionErr     error
)

// Run calls fn inside Ebitengine, where images can be read back, and waits
// until it's done.
//
// Go runs a package's tests in a single process, but Ebitengine can only be
// started once per process. So the first call starts Ebitengine in the
// background, every call hands its function over to it, and Ebitengine stops
// when the process exits after the last test. As a result:
//   - A display is needed: a small window stays open while the tests run.
//   - Tests that start Ebitengine themselves (e.g. interactive tests) can't
//     run in the same process.
//   - Inside fn, use t.Error rather than t.Fatal, which would stop Ebitengine
//     instead of the test. Panics are reported as failures.
func Run(t testing.TB, fn func()) {
	t.Helper()
	startOnce.Do(startSession)
	work := &job{fn: fn, done: make(chan struct{})}
	select {
	case jobs <- work:
	case <-sessionStopped:
		t.Fatalf("pixcheck: session stopped: %v", sessionErr)
	}
	select {
	case <-work.done:
	case <-sessionStopped:
		t.Fatalf("pixcheck: session stopped: %v", sessionErr)
	}
	if work.panicked != "" {
		t.Fatal(work.panicked)
	}
}

func startSession() {
	// TODO: after updating to Ebitengine v2.10, call
	// ebiten.SetWindowVisible(false) here to run without showing the window.
	ebiten.SetWindowSize(128, 128)
	ebiten.SetWindowTitle("pixcheck")
	go func() {
		sessionErr = ebiten.RunGameWithOptions(session{}, &ebiten.RunGameOptions{InitUnfocused: true})
		if sessionErr == nil {
			sessionErr = errors.New("window closed")
		}
		close(sessionStopped)
	}()
}

// job is a function passed by [Run] to the session.
type job struct {
	fn       func()
	panicked string
	done     chan struct{}
}

func (j *job) run() {
	defer close(j.done)
	defer func() {
		if err := recover(); err != nil {
			j.panicked = fmt.Sprintf("panic: %v\n%s", err, debug.Stack())
		}
	}()
	j.fn()
}

// session implements [ebiten.Game] to run the jobs passed by [Run].
type session struct{}

func (session) Update() error {
	select {
	case work := <-jobs:
		work.run()
	default:
	}
	return nil
}

func (session) Draw(*ebiten.Image)         {}
func (session) Layout(w, h int) (int, int) { return w, h }
