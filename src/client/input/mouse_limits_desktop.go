//go:build !android

package input

// A release within these limits of the press counts as a click; beyond them
// it becomes a drag end. Tuned for mice: a stationary press generates no
// events, so the sequence budget stays small.
const (
	clickMaxSeqDelta = 15
	clickMaxDist2    = 100 // 10px in draw space
)
