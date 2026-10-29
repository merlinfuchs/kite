package notify

import "testing"

func TestEscape(t *testing.T) {
	got := escape("x** [Approve](https://evil.example) _a_")
	want := `x\*\* \[Approve\]\(https://evil.example\) \_a\_`
	if got != want {
		t.Errorf("escape() = %q, want %q", got, want)
	}
}
