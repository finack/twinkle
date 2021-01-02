package main 

import (
	"os"
	"fmt"
	"flag"
	"sort"

	"github.com/finack/twinkle/internal/config"
	"github.com/finack/twinkle/internal/display"

	"golang.org/x/image/colornames"
  "github.com/rs/zerolog"
  "github.com/rs/zerolog/log"
)

func main() {
  configFile := flag.String("config", "config.yaml", "Path to configuration file")

  flag.Parse()

  c := config.GetConfig(configFile)
  log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	leds, err := display.New(c.Brightness, c.LedCount)
	if err != nil {
		log.Fatal().Err(err).Caller().Msg("Could not setup LEDS")
	}

  keys := make([]int, len(c.Leds))
	i := 0
	for k := range c.Leds {
			keys[i] = k
			i++
	}
	sort.Ints(keys)

	log.Info().Msg("Hit enter to continue")
	for _, k := range keys {
		num := k
		station := c.Leds[k]
		log.Info().Int("ledNum", num).Str("station", station).Msg("Displaying")
		leds.Clear()
		leds.Display(num, colornames.Crimson)
		if err := leds.Ws.Render(); err != nil {
			log.Error().Err(err).Caller().Msg("Could not render leds")
			continue
		}

		var input string
		fmt.Scanln(&input)
	}
  leds.Clear()
}

