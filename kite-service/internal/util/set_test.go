package util

import "testing"

func TestIDSet(t *testing.T) {
	set := IDSet([]string{"a", "b", "a"})

	if len(set) != 2 {
		t.Errorf("IDSet produced %d entries, want 2 after dedup", len(set))
	}
	for _, want := range []string{"a", "b"} {
		if _, ok := set[want]; !ok {
			t.Errorf("IDSet is missing %q", want)
		}
	}
	if _, ok := set["c"]; ok {
		t.Error("IDSet contains an id that was not passed in")
	}
}
