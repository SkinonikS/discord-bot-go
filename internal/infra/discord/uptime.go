package discord

import "time"

type UpTime time.Time

func NewUpTime() UpTime {
	t := time.Now()
	return UpTime(t)
}

func (u UpTime) Time() time.Time {
	return time.Time(u)
}
