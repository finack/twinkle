package display

import (
	"testing"
	"time"
)

func TestCalcRiseSet(t *testing.T) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().In(loc)

	rise, set, err := calcRiseSet(now, -122.0578, 37.9884)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	zero := time.Time{}
	if rise == zero {
		t.Error("rise is zero time")
	}
	if set == zero {
		t.Error("set is zero time")
	}
	if !rise.Before(set) {
		t.Errorf("expected rise (%v) to be before set (%v)", rise, set)
	}
	if rise.Hour() >= 12 {
		t.Errorf("sunrise hour %d looks wrong (expected AM)", rise.Hour())
	}
	if set.Hour() < 12 {
		t.Errorf("sunset hour %d looks wrong (expected PM)", set.Hour())
	}
}

func TestCalcBrightness(t *testing.T) {
	base := time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC)
	rise := base.Add(6 * time.Hour) // 06:00
	set := base.Add(20 * time.Hour) // 20:00
	day := 200
	night := 50

	tests := []struct {
		name string
		now  time.Time
		want int
	}{
		{
			name: "deep night before sunrise window",
			now:  rise.Add(-2 * time.Hour), // 04:00, well before transition
			want: night,
		},
		{
			name: "start of sunrise transition",
			now:  rise.Add(-30 * time.Minute), // 05:30
			want: night,
		},
		{
			name: "halfway through sunrise transition",
			now:  rise, // 06:00 — 30min into 60min window
			want: (day + night) / 2,
		},
		{
			name: "end of sunrise transition",
			now:  rise.Add(30 * time.Minute), // 06:30
			want: day,
		},
		{
			name: "midday",
			now:  base.Add(12 * time.Hour), // 12:00
			want: day,
		},
		{
			name: "start of sunset transition",
			now:  set.Add(-30 * time.Minute), // 19:30
			want: day,
		},
		{
			name: "halfway through sunset transition",
			now:  set, // 20:00 — 30min into 60min window
			want: (day + night) / 2,
		},
		{
			name: "end of sunset transition",
			now:  set.Add(30 * time.Minute), // 20:30
			want: night,
		},
		{
			name: "deep night after sunset",
			now:  set.Add(2 * time.Hour), // 22:00
			want: night,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcBrightness(tt.now, rise, set, day, night)
			// Allow ±1 for integer rounding
			if got < tt.want-1 || got > tt.want+1 {
				t.Errorf("calcBrightness at %v: got %d, want %d (±1)", tt.now.Format("15:04"), got, tt.want)
			}
		})
	}
}
