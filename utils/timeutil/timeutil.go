package timeutil

import "time"

type TimeUtil interface {
	Now() time.Time
}

type timeUtil struct{}

func (t timeUtil) Now() time.Time {
	return time.Now()
}

func NewTimeUtil() TimeUtil {
	return &timeUtil{}
}
