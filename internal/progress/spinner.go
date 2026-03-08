// Package progress provides animated progress indicators for CLI applications.
package progress

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Spinner represents an animated spinner
type Spinner struct {
	frames   []string
	interval time.Duration
	message  string
	stop     chan struct{}
	wg       sync.WaitGroup
	active   bool
	mu       sync.Mutex
}

// DefaultSpinnerFrames contains the default spinner animation frames
var DefaultSpinnerFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

// NewSpinner creates a new spinner with default settings
func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:   DefaultSpinnerFrames,
		interval: 100 * time.Millisecond,
		message:  message,
		stop:     make(chan struct{}),
	}
}

// Start starts the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s %s ", s.frames[i%len(s.frames)], s.message)
				i++
				time.Sleep(s.interval)
			}
		}
	}()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	close(s.stop)
	s.wg.Wait()
	fmt.Fprintf(os.Stderr, "\r✓ %s\n", s.message)
}

// StopWithMessage stops the spinner with a custom message
func (s *Spinner) StopWithMessage(message string) {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		fmt.Fprintf(os.Stderr, "\r%s\n", message)
		return
	}
	s.active = false
	s.mu.Unlock()

	close(s.stop)
	s.wg.Wait()
	fmt.Fprintf(os.Stderr, "\r✓ %s\n", message)
}

// StopWithError stops the spinner with an error message
func (s *Spinner) StopWithError(message string) {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		fmt.Fprintf(os.Stderr, "\r✗ %s\n", message)
		return
	}
	s.active = false
	s.mu.Unlock()

	close(s.stop)
	s.wg.Wait()
	fmt.Fprintf(os.Stderr, "\r✗ %s\n", message)
}

// SetMessage updates the spinner message
func (s *Spinner) SetMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// SpinnerFunc is a convenience function for running code with a spinner
func SpinnerFunc(message string, fn func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()
	defer spinner.Stop()

	if err := fn(); err != nil {
		spinner.StopWithError(message)
		return err
	}
	return nil
}
