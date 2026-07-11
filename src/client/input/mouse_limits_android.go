//go:build android

package input

// Touchscreens stream motion events the whole time a finger is down
// (120Hz sampling plus micro-jitter), so a tap easily exceeds the desktop
// budgets and would always be classified as a drag. Real GUI drags still
// trip the distance check: 30 draw-space px is well beyond finger jitter.
const (
	clickMaxSeqDelta = 1200
	clickMaxDist2    = 900 // 30px in draw space
)
