# exponential-moving-average-latency

Computes an exponential moving average of request durations from a stream of individual measurements.

```go
package main

import (
	"fmt"
	"time"

	ema "exponential_moving_average_latency"
)

func main() {
	// Half-life of 10 samples: an old measurement matters half as much
	// after 10 further samples have been added.
	e := ema.New(10)

	_ = e.Add(120 * time.Millisecond)
	_ = e.Add(95 * time.Millisecond)
	_ = e.Add(300 * time.Millisecond)

	avg, ok := e.Value()
	fmt.Println(avg, ok) // e.g. 143.75ms true
}
```

## Why this exists

Latency is bursty. A plain mean over a long window reacts slowly to a real
regression; a plain mean over a short window is noisy and a single outlier can
move it for a long time. An exponential moving average sits in between: it
weights recent samples more heavily and decays older ones geometrically.

The trade-off is that the EMA has a memory that never fully reaches zero. An
old measurement keeps a tiny influence forever. The half-life parameter makes
that decay explicit: after `halfLife` samples, a given sample's weight has
halved. Pick a half-life in samples, not seconds, because the EMA advances once
per measurement, not once per unit of wall-clock time.

## API

- `New(halfLife float64) *EMA` — construct an EMA. `halfLife` is in samples
  and must be positive and finite; it panics otherwise.
- `(*EMA) Add(d time.Duration) error` — incorporate one sample.
- `(*EMA) Value() (time.Duration, bool)` — current average, and whether any
  sample has been added.
- `(*EMA) Ready() bool` — whether at least one sample has been added.
- `(*EMA) Reset()` — clear the state, keeping the half-life.
- `(*EMA) HalfLife() float64` — the configured half-life.
- `ErrNotReady` — returned by `Add` on a zero-value `EMA`.

## The awkward edge

The zero value of `EMA` is not usable for adding samples: without a half-life
it cannot compute an alpha, so `Add` returns `ErrNotReady`. Use `New` to get a
working value. `Value` on an unready EMA returns `0, false` rather than
inventing a number.
