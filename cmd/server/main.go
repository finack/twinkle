package main

import (
  "fmt"
  "flag"
  "os"

	"github.com/finack/twinkle/internal/config"
	"github.com/finack/twinkle/internal/display"
	"github.com/finack/twinkle/internal/metardata"
	"github.com/finack/twinkle/internal/signals"

  "github.com/rs/zerolog"
  "github.com/rs/zerolog/log"
)

func main() {

  debug := flag.Bool("debug", false, "Sets log level to debug")
  configFile := flag.String("config", "config.yaml", "Path to configuration file")

  flag.Parse()

  c := config.GetConfig(configFile)
  log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

  if *debug {
    zerolog.SetGlobalLevel(zerolog.DebugLevel)
  } else {
    zerolog.SetGlobalLevel(zerolog.InfoLevel)
  }

  log.
    Info().
    Int("ledCount", c.LedCount).
    Int("stationCount", len(c.Stations)).
    Int("metarRefreshRateS", c.MetarRefreshRateS).
    Str("logLevel", fmt.Sprintf("%v", zerolog.GlobalLevel())).
    Msg("Starting Twinkle!")

	stopApplication := make(chan bool)
	stopLedUpdate, _, ledChannel := display.UpdateRoutine(c)
	stopMetarUpdate := metardata.FetchRoutine(c, ledChannel)

	signals.CatchSignals(stopMetarUpdate, stopLedUpdate, stopApplication)

	<-stopApplication
}
