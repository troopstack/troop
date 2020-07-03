package utils

import (
	"log"
	"os"
	"os/signal"

	taskCache "github.com/troopstack/troop/src/modules/general/cache/task"
)

func SignalListen() {
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh)

	s := <-signalCh // 阻塞直至有信号传入
	if s == os.Interrupt || s == os.Kill {
		taskCache.SaveTasksToLocal()
		log.Println("Exiting on", s)
		os.Exit(0)
	} else {
		SignalListen()
	}
}
