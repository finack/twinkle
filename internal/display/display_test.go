package display

import (
	"image/color"
	"testing"
)

// mockWsEngine satisfies the wsEngine interface for tests.
type mockWsEngine struct {
	leds        []uint32
	renderErr   error
	renderCalls int
}

func newMock(size int) *mockWsEngine {
	return &mockWsEngine{leds: make([]uint32, size)}
}

func (m *mockWsEngine) Init() error                           { return nil }
func (m *mockWsEngine) Fini()                                 {}
func (m *mockWsEngine) Wait() error                           { return nil }
func (m *mockWsEngine) SetBrightness(channel, brightness int) {}
func (m *mockWsEngine) Leds(channel int) []uint32             { return m.leds }
func (m *mockWsEngine) Render() error {
	m.renderCalls++
	return m.renderErr
}

// ParseHexColor tests

func TestParseHexColor_RRGGBB(t *testing.T) {
	c, err := ParseHexColor("#ff8800")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.R != 0xff || c.G != 0x88 || c.B != 0x00 || c.A != 0xff {
		t.Errorf("got R=%x G=%x B=%x A=%x", c.R, c.G, c.B, c.A)
	}
}

func TestParseHexColor_RGB(t *testing.T) {
	c, err := ParseHexColor("#f80")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// #f80 expands to #ff8800 via *17
	if c.R != 0xff || c.G != 0x88 || c.B != 0x00 {
		t.Errorf("got R=%x G=%x B=%x", c.R, c.G, c.B)
	}
}

func TestParseHexColor_Uppercase(t *testing.T) {
	c, err := ParseHexColor("#FF0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.R != 0xff || c.G != 0x00 || c.B != 0x00 {
		t.Errorf("got %+v", c)
	}
}

func TestParseHexColor_MissingHash(t *testing.T) {
	_, err := ParseHexColor("ff0000")
	if err == nil {
		t.Error("expected error for missing #, got nil")
	}
}

func TestParseHexColor_WrongLength(t *testing.T) {
	_, err := ParseHexColor("#ff00")
	if err == nil {
		t.Error("expected error for wrong length, got nil")
	}
}

func TestParseHexColor_InvalidChars(t *testing.T) {
	_, err := ParseHexColor("#zz0000")
	if err == nil {
		t.Error("expected error for invalid hex chars, got nil")
	}
}

// ParseRGBAtoUint32 tests

func TestParseRGBAtoUint32(t *testing.T) {
	tests := []struct {
		name string
		c    color.RGBA
		want uint32
	}{
		{"red", color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}, 0xff0000},
		{"green", color.RGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff}, 0x00ff00},
		{"blue", color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}, 0x0000ff},
		{"white", color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, 0xffffff},
		{"black", color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}, 0x000000},
		{"mixed", color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}, 0x123456},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRGBAtoUint32(tt.c)
			if got != tt.want {
				t.Errorf("got %#x, want %#x", got, tt.want)
			}
		})
	}
}

// Leds.Display tests

func TestLedsDisplay(t *testing.T) {
	mock := newMock(5)
	l := newWithEngine(mock)

	red := color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
	l.Display(2, red)

	want := ParseRGBAtoUint32(red)
	if mock.leds[2] != want {
		t.Errorf("leds[2] = %#x, want %#x", mock.leds[2], want)
	}
	// Other slots should be untouched
	if mock.leds[0] != 0 || mock.leds[1] != 0 {
		t.Errorf("expected other LEDs to be 0")
	}
}

func TestLedsClear(t *testing.T) {
	mock := newMock(4)
	mock.leds[0] = 0xff0000
	mock.leds[3] = 0x00ff00
	l := newWithEngine(mock)

	if err := l.Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}

	for i, v := range mock.leds {
		if v != 0 {
			t.Errorf("leds[%d] = %#x after Clear, want 0", i, v)
		}
	}

	if mock.renderCalls != len(mock.leds) {
		t.Errorf("Render called %d times, want %d (once per LED)", mock.renderCalls, len(mock.leds))
	}
}

func TestParseHexColor_Empty(t *testing.T) {
	_, err := ParseHexColor("")
	if err == nil {
		t.Error("expected error for empty string, got nil")
	}
}
