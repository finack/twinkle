package config

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Leds              map[int]string `yaml:"leds,omitempty"`
	Stations          map[string]int
	LedCount          int `yaml:"led_count,omitempty"`
	Brightness        int `yaml:"brightness,omitempty"`
	MetarRefreshRateS int `yaml:"metar_refresh_rate_s,omitempty"` // seconds
	LedRefreshRateMS  int `yaml:"led_refresh_rate_ms,omitempty"`  // milliseconds
}

func GetConfig() Config {

	c := Config{}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Fatalf("error: %v", err)
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
