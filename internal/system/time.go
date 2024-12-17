package system

import "time"

type Time interface {
	Now() time.Time
}

type Clock struct{}

func (Clock) Now() time.Time { return time.Now() }
