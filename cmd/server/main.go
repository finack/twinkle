package main

import (
	"log"

	"github.com/finack/twinkle/internal/config"
	"github.com/finack/twinkle/internal/display"
	"github.com/finack/twinkle/internal/metardata"
	"github.com/finack/twinkle/internal/signals"
)

func main() {

	c := config.GetConfig()

	log.Printf("Starting : leds=%v stations=%v MetarRefreshRateS=%v", c.LedCount, len(c.Stations), c.MetarRefreshRateS)

	stopApplication := make(chan bool)
	stopLedUpdate, _, ledChannel := display.UpdateRoutine(c)
	stopMetarUpdate := metardata.FetchRoutine(c, ledChannel)

	signals.CatchSignals(stopMetarUpdate, stopLedUpdate, stopApplication)

	<-stopApplication
}
