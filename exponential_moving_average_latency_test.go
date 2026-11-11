package exponential_moving_average_latency

import (
	"math"
	"testing"
	"time"
)

const eps = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < eps
}

func TestNewSetsHalfLife(t *testing.T) {
	e := New(5)
	if got := e.HalfLife(); got != 5 {
		t.Fatalf("HalfLife() = %v, want 5", got)
	}
}

func TestNewPanicsOnInvalidHalfLife(t *testing.T) {
	for _, hl := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("New(%v) did not panic", hl)
				}
			}()
			New(hl)
		}()
	}
}

func TestZeroValueIsNotReady(t *testing.T) {
	var e EMA
	if e.Ready() {
		t.Fatal("zero EMA reports Ready")
	}
	if _, ok := e.Value(); ok {
		t.Fatal("zero EMA returned a value before any sample")
	}
	if err := e.Add(10 * time.Millisecond); err != ErrNotReady {
		t.Fatalf("Add on zero EMA = %v, want ErrNotReady", err)
	}
}

func TestFirstSampleBecomesValue(t *testing.T) {
	e := New(10)
	if err := e.Add(250 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	got, ok := e.Value()
	if !ok {
		t.Fatal("Value not ready after first sample")
	}
	if got != 250*time.Millisecond {
		t.Fatalf("Value() = %v, want 250ms", got)
	}
}

func TestAlphaMatchesHalfLifeFormula(t *testing.T) {
	// After one half-life worth of samples, the weight of the initial value
	// should have halved. With a constant input this is hard to observe, so
	// instead we check the alpha used internally by observing two steps.
	e := New(1)
	// alpha = 1 - exp(-ln2/1) = 0.5
	if err := e.Add(0); err != nil {
		t.Fatal(err)
	}
	if err := e.Add(1000 * time.Nanosecond); err != nil {
		t.Fatal(err)
	}
	// value = 0.5*1000 + 0.5*0 = 500
	got, _ := e.Value()
	if got != 500*time.Nanosecond {
		t.Fatalf("Value() = %v, want 500ns", got)
	}
}

func TestEMASmoothsSpike(t *testing.T) {
	e := New(2)
	// Feed a steady 100ns baseline, then one 1000ns spike.
	for i := 0; i < 10; i++ {
		if err := e.Add(100 * time.Nanosecond); err != nil {
			t.Fatal(err)
		}
	}
	before, _ := e.Value()
	if err := e.Add(1000 * time.Nanosecond); err != nil {
		t.Fatal(err)
	}
	after, _ := e.Value()
	if after <= before {
		t.Fatalf("spike did not raise the average: before=%v after=%v", before, after)
	}
	if after >= 1000*time.Nanosecond {
		t.Fatalf("spike was not smoothed: after=%v", after)
	}
}

func TestResetKeepsHalfLife(t *testing.T) {
	e := New(3)
	if err := e.Add(42 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	e.Reset()
	if e.Ready() {
		t.Fatal("EMA still ready after Reset")
	}
	if _, ok := e.Value(); ok {
		t.Fatal("Value returned a value after Reset")
	}
	if got := e.HalfLife(); got != 3 {
		t.Fatalf("HalfLife() = %v after Reset, want 3", got)
	}
	if err := e.Add(7 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	got, _ := e.Value()
	if got != 7*time.Millisecond {
		t.Fatalf("Value() = %v after Reset+Add, want 7ms", got)
	}
}

func TestFloatTolerance(t *testing.T) {
	// The EMA stores nanoseconds as a float64; exact equality is not
	// guaranteed for arbitrary values, so compare with a tolerance.
	e := New(4.5)
	if err := e.Add(123456789 * time.Nanosecond); err != nil {
		t.Fatal(err)
	}
	got, _ := e.Value()
	want := 123456789.0
	if !almostEqual(float64(got), want) {
		t.Fatalf("Value() = %v, want ~%v", got, want)
	}
}
