package notify

import "testing"

func TestFormatCredits(t *testing.T) {
	for n, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 10000: "10,000", 1234567: "1,234,567"} {
		if got := formatCredits(n); got != want {
			t.Errorf("formatCredits(%d) = %q, want %q", n, got, want)
		}
	}
}
