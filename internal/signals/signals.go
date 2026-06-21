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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Info().Str("signal", fmt.Sprintf("%v", sig)).Msg("Shutting down Twinkle!")
		// stopAutomaticDimmer <- true
		signal.Stop(sigs)
		stopLedUpdate <- true
		stopMetarClient <- true
		time.Sleep(time.Millisecond * 500)
		stopApplication <- true
		os.Exit(0)
	}()
}
