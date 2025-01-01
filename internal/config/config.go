package config

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Leds              map[int]string `yaml:"leds,omitempty"`
	Stations          map[string]int
	LedCount          int     `yaml:"led_count,omitempty"`
	Brightness        int     `yaml:"brightness,omitempty"`
	MetarRefreshRateS int     `yaml:"metar_refresh_rate_s,omitempty"` // seconds
	LedRefreshRateMS  int     `yaml:"led_refresh_rate_ms,omitempty"`  // milliseconds
	Latitude          float64 `yaml:"latitude,omitempty"`
	Longitude         float64 `yaml:"longitude,omitempty"`
	Locale            string  `yaml:"locale,omitempty"`
}

func GetConfig(file *string) Config {
	c := Config{}

	data, err := os.ReadFile(*file)
	if err != nil {
		log.
			Fatal().
			Err(err).
			Caller().
			Str("configFile", *file).
			Msg("Could not read config file")
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.
			Fatal().
			Err(err).
			Caller().
			Msg("Cound not unmarshal configuration file")
	}

	c.Stations = reverseLeds(c.Leds)
	return c
}

func reverseLeds(m map[int]string) map[string]int {
	n := make(map[string]int)
	for k, v := range m {
		n[strings.ToUpper(v)] = k
	}
	return n
}
