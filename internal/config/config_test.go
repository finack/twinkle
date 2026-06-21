package config

import (
	"os"
	"testing"
)

func TestGetConfig(t *testing.T) {
	yaml := `
led_count: 10
brightness: 150
night_brightness: 40
metar_refresh_rate_s: 300
led_refresh_rate_ms: 100
latitude: 37.9884
longitude: -122.0578
locale: America/Los_Angeles
leds:
  0: KOAK
  1: ksfo
  2: KSQL
`
	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(yaml)
	f.Close()

	name := f.Name()
	c := GetConfig(&name)

	if c.LedCount != 10 {
		t.Errorf("LedCount: got %d, want 10", c.LedCount)
	}
	if c.Brightness != 150 {
		t.Errorf("Brightness: got %d, want 150", c.Brightness)
	}
	if c.NightBrightness != 40 {
		t.Errorf("NightBrightness: got %d, want 40", c.NightBrightness)
	}
	if c.MetarRefreshRateS != 300 {
		t.Errorf("MetarRefreshRateS: got %d, want 300", c.MetarRefreshRateS)
	}
	if c.Locale != "America/Los_Angeles" {
		t.Errorf("Locale: got %q, want America/Los_Angeles", c.Locale)
	}
	if c.WindLowKt != 0 {
		t.Errorf("WindLowKt: got %v, want 0 (omitted from yaml)", c.WindLowKt)
	}
	if len(c.Leds) != 3 {
		t.Errorf("Leds count: got %d, want 3", len(c.Leds))
	}
	if c.Stations["KOAK"] != 0 {
		t.Errorf("Stations[KOAK]: got %d, want 0", c.Stations["KOAK"])
	}
	// lowercase in YAML should be uppercased in Stations map
	if c.Stations["KSFO"] != 1 {
		t.Errorf("Stations[KSFO]: got %d, want 1", c.Stations["KSFO"])
	}
}

func TestReverseLeds(t *testing.T) {
	input := map[int]string{
		0:  "KOAK",
		5:  "ksfo",
		12: "KSQL",
	}

	result := reverseLeds(input)

	if result["KOAK"] != 0 {
		t.Errorf("KOAK: got %d, want 0", result["KOAK"])
	}
	if result["KSFO"] != 5 {
		t.Errorf("KSFO: got %d, want 5 (lowercase input should be uppercased)", result["KSFO"])
	}
	if result["KSQL"] != 12 {
		t.Errorf("KSQL: got %d, want 12", result["KSQL"])
	}
	if len(result) != 3 {
		t.Errorf("result length: got %d, want 3", len(result))
	}
}

func TestReverseLeds_Empty(t *testing.T) {
	result := reverseLeds(map[int]string{})
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}
