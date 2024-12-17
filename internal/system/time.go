package system

import "time"

// Time provides an interface for retrieving the current time.
type Time interface {
	Now() time.Time
}

// Clock is a struct representing a mechanism to retrieve the current time.
type Clock struct{}

// Now retrieves the current local time as a time.Time instance.
func (Clock) Now() time.Time { return time.Now() }
