package appstate

import "testing"

func TestImpersonatesDiscord(t *testing.T) {
	for name, want := range map[string]bool{
		"Kite":            false,
		"Server Helper":   false,
		"Discord":         true,
		"discord support": true,
		"D.i.s.c.o.r.d":   true,
		"Dis cord Staff":  true,
	} {
		if got := impersonatesDiscord(name); got != want {
			t.Errorf("%q: %v, want %v", name, got, want)
		}
	}
}
