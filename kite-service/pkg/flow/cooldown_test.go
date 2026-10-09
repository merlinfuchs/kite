package flow

import (
	"testing"
	"time"
)

func TestParseCooldownDuration(t *testing.T) {
	valid := map[string]time.Duration{
		"1":     time.Second,
		"60":    time.Minute,
		" 90 ":  90 * time.Second,
		"3600":  time.Hour,
		"00120": 2 * time.Minute,
	}
	for raw, want := range valid {
		got, err := parseCooldownDuration(raw)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", raw, err)
		} else if got != want {
			t.Errorf("%q: got %v, want %v", raw, got, want)
		}
	}

	invalid := []string{
		"",
		"   ",
		"abc",
		"1.5",
		"0",
		"-5",
		"3601",
		"9223372036854775807",
		"99999999999999999999",
	}
	for _, raw := range invalid {
		if got, err := parseCooldownDuration(raw); err == nil {
			t.Errorf("%q: expected an error, got %v", raw, got)
		}
	}
}

func TestCooldownDataValidation(t *testing.T) {
	valid := []FlowNodeData{
		{CooldownDurationSeconds: "30"},
		{CooldownDurationSeconds: "30", CooldownScope: CooldownScopeGuild},
		{CooldownDurationSeconds: "{{arg('seconds')}}"},
	}
	for _, d := range valid {
		if err := d.Validate(FlowNodeTypeOptionCommandCooldown); err != nil {
			t.Errorf("%+v: unexpected error: %v", d, err)
		}
	}

	invalid := []FlowNodeData{
		{},
		{CooldownDurationSeconds: "1.5"},
		{CooldownDurationSeconds: "7200"},
		{CooldownDurationSeconds: "30", CooldownScope: "channel"},
	}
	for _, d := range invalid {
		if err := d.Validate(FlowNodeTypeOptionCommandCooldown); err == nil {
			t.Errorf("%+v: expected an error", d)
		}
	}
}
