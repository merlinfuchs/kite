// Package schedule parses the cron expressions of scheduled event listeners and
// defines the event that triggers their flows.
package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// The seconds field is optional so standard 5-field expressions keep their
// usual meaning. Descriptors like @hourly and @every are allowed too.
var parser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// Schedules are always evaluated in UTC.
type Schedule struct {
	cron cron.Schedule
}

func Parse(expr string) (*Schedule, error) {
	s, err := parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}
	return &Schedule{cron: s}, nil
}

// Next returns the first occurrence strictly after t, or the zero time if
// there is none within the next five years.
func (s *Schedule) Next(t time.Time) time.Time {
	return s.cron.Next(t.UTC())
}

// gapSamples bounds how many upcoming occurrences MinGap looks at. Expressions
// like "0,1 * * * *" only show their shortest gap once per period, so a single
// pair of runs isn't enough.
const gapSamples = 100

// MinGap returns the shortest time between consecutive upcoming occurrences
// after from, or 0 if the schedule never runs.
func (s *Schedule) MinGap(from time.Time) time.Duration {
	prev := s.Next(from)
	if prev.IsZero() {
		return 0
	}

	var minGap time.Duration
	for range gapSamples {
		next := s.Next(prev)
		if next.IsZero() {
			break
		}
		if gap := next.Sub(prev); minGap == 0 || gap < minGap {
			minGap = gap
		}
		prev = next
	}

	return minGap
}
