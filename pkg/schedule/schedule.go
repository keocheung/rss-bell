package schedule

import (
	"math/rand"
	"time"

	"github.com/robfig/cron/v3"
)

type randomDelaySchedule struct {
	maxDelay int32
	schedule cron.Schedule
}

func (s *randomDelaySchedule) Next(from time.Time) time.Time {
	if s.maxDelay <= 0 {
		return s.schedule.Next(from)
	}
	delay := rand.Int31n(s.maxDelay)
	return s.schedule.Next(from).Add(time.Duration(delay) * time.Second)
}

func NewRandomDelaySchedule(spec string, maxDelayInSecond int32) (cron.Schedule, error) {
	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return nil, err
	}
	return &randomDelaySchedule{
		maxDelay: maxDelayInSecond,
		schedule: schedule,
	}, nil
}
