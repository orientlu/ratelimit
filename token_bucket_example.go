package main

import (
	"github.com/orientlu/ratelimit/tokenbucket"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	log.Info("new tokenbucket, rate 9, cap 100")
	tb, err := tokenbucket.New(100, 200, time.Duration(10)*time.Millisecond)
	if err != nil {
		log.Infoln(err)
		return
	}
	go tb.Update()

	log.Info("1 second..")
	time.Sleep(time.Duration(1000) * time.Millisecond)
	log.Infoln(tb.GetInfo())
	log.Info("4 second..")
	time.Sleep(time.Duration(3000) * time.Millisecond)
	log.Infoln(tb.GetInfo())

	var count = 0
	for i := 0; i < 1000; i++ {
		if tb.TryAcquire() {
			count++
		}
	}
	log.Info("require token:", count)
	log.Infoln(tb.GetInfo())

	log.Info("10 second run")
	count = 0
	for i := 0; i < 100; i++ {
		if tb.TryAcquire() {
			count++
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	log.Info("require token:", count)
	log.Infoln(tb.GetInfo())

	tb.StopUpdate()
}
