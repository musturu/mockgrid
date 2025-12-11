// Package clock provides a time abstraction for testable code.
package clock

import "time"

// Clock is an interface for time operations, allowing tests to control time.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// RealClock implements Clock using the real system time.
type RealClock struct{}

// Now returns the current system time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// MockClock implements Clock with a fixed, controllable time.
type MockClock struct {
	current time.Time
}

// NewMockClock creates a MockClock set to the given time.
func NewMockClock(t time.Time) *MockClock {
	return &MockClock{current: t}
}

// Now returns the mock's current time.
func (m *MockClock) Now() time.Time {
	return m.current
}

// Set updates the mock's current time.
func (m *MockClock) Set(t time.Time) {
	m.current = t
}

// Add advances the mock's current time by the given duration.
func (m *MockClock) Add(d time.Duration) {
	m.current = m.current.Add(d)
}
