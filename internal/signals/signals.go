package signals

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func CatchSignals(stopMetarClient chan bool, stopLedUpdate chan bool, stopApplication chan bool) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Caught SIGINIT/SIGTERM. Gracefully exiting...")
		stopLedUpdate <- true
		stopMetarClient <- true
		time.Sleep(time.Millisecond * 500)
		stopApplication <- true
		return
	}()
}
