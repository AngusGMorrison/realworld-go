package testutil

import "time"

// TimeSource provides the current time, allowing time to be mocked in tests.
type TimeSource interface {
	Now() time.Time
}

// StdTimeSource keeps time using the standard library.
type StdTimeSource struct{}

func (StdTimeSource) Now() time.Time {
	return time.Now()
}

// FixedTimeSource returns a fixed time.
type FixedTimeSource struct {
	Time time.Time
}

func (f FixedTimeSource) Now() time.Time {
	return f.Time
}
