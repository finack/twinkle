package signals

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func CatchSignals(stopMetarClient chan bool, stopLedUpdate chan bool, stopApplication chan bool) {
	// func CatchSignals(stopMetarClient chan bool, stopLedUpdate chan bool, stopApplication chan bool, stopAutomaticDimmer chan bool) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sigs
		log.Info().Str("signal", fmt.Sprintf("%v", s)).Msg("Shutting down Twinkle!")
		// stopAutomaticDimmer <- true
		stopLedUpdate <- true
		stopMetarClient <- true
		time.Sleep(time.Millisecond * 500)
		stopApplication <- true
		return
	}()
}
