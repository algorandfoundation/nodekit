package mock

import "time"

type Clock struct{}

func (Clock) Now() time.Time { return time.Time{} }
