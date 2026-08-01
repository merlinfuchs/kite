package util

// IDSet builds a membership set from a slice of IDs.
//
// Callers that test many candidates against the same slice should build this
// once and share it: the engine's dangling sweeps previously rebuilt an
// equivalent map per app, which made them O(apps x entities).
func IDSet[T comparable](ids []T) map[T]struct{} {
	set := make(map[T]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}
