// Package exponential_moving_average_latency computes an exponential moving
// average (EMA) of request durations from a stream of individual measurements.
//
// The EMA gives more weight to recent samples while still smoothing over the
// whole history, which is useful for latency tracking: a single slow request
// nudges the estimate without dominating it forever.
//
// The smoothing factor alpha is chosen from a half-life: alpha = 1 -
// exp(-ln(2) / halfLife). This means that after halfLife samples, the weight
// of an old sample has halved. The half-life form is used instead of a raw
// alpha because it is easier to reason about: "how many samples until an old
// measurement matters half as much?"
package exponential_moving_average_latency

import (
	"errors"
	"math"
	"time"
)

// EMA computes an exponential moving average of time.Duration samples.
//
// A zero EMA is not ready: Ready reports false and Add returns an error until
// the first sample has been added. The zero value is otherwise usable; use New
// to set a half-life.
type EMA struct {
	halfLife float64 // half-life in samples; must be > 0
	alpha    float64 // smoothing factor derived from halfLife
	value    float64 // current EMA in nanoseconds
	ready    bool
}

// ErrNotReady is returned by Add before the first sample has been added.
var ErrNotReady = errors.New("ema: no samples yet")

// New returns an EMA with the given half-life expressed in samples. The
// half-life must be positive and finite; otherwise New panics, because a
// misconfigured half-life is a programming error, not a runtime condition.
func New(halfLife float64) *EMA {
	if !(halfLife > 0) || math.IsInf(halfLife, 0) || math.IsNaN(halfLife) {
		panic("ema: half-life must be a positive finite number")
	}
	alpha := 1 - math.Exp(-math.Ln2/halfLife)
	return &EMA{halfLife: halfLife, alpha: alpha}
}

// HalfLife returns the configured half-life in samples.
func (e *EMA) HalfLife() float64 {
	return e.halfLife
}

// Ready reports whether at least one sample has been added.
func (e *EMA) Ready() bool {
	return e.ready
}

// Add incorporates a single duration sample into the EMA. It returns
// ErrNotReady only when called on a zero-value EMA (created without New),
// because the zero value has no configured half-life and cannot compute a
// meaningful average.
func (e *EMA) Add(d time.Duration) error {
	if e.alpha == 0 {
		return ErrNotReady
	}
	x := float64(d)
	if !e.ready {
		e.value = x
		e.ready = true
		return nil
	}
	e.value = e.alpha*x + (1-e.alpha)*e.value
	return nil
}

// Value returns the current EMA as a time.Duration. The second return value
// is false when no sample has been added yet; in that case Value returns 0.
func (e *EMA) Value() (time.Duration, bool) {
	if !e.ready {
		return 0, false
	}
	return time.Duration(e.value), true
}

// Reset clears the EMA back to its initial unready state, keeping the
// configured half-life. This lets a caller reuse an EMA across measurement
// windows without allocating a new one.
func (e *EMA) Reset() {
	e.value = 0
	e.ready = false
}
