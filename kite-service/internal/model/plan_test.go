package model

import (
	"testing"
	"time"
)

func TestFeaturesMinScheduleInterval(t *testing.T) {
	if got := (Features{}).MinScheduleInterval(); got != DefaultMinScheduleInterval {
		t.Errorf("unset interval = %s, want default %s", got, DefaultMinScheduleInterval)
	}

	free := Features{MinScheduleIntervalSeconds: 300}
	premium := Features{MinScheduleIntervalSeconds: 1}
	unset := Features{}

	if got := free.Merge(premium).MinScheduleInterval(); got != time.Second {
		t.Errorf("free merged with premium = %s, want 1s", got)
	}
	if got := unset.Merge(free).MinScheduleInterval(); got != 5*time.Minute {
		t.Errorf("unset merged with free = %s, want 5m", got)
	}
	if got := premium.Merge(unset).MinScheduleInterval(); got != time.Second {
		t.Errorf("premium merged with unset = %s, want 1s", got)
	}
}
