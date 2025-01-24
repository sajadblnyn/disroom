package main

import (
	"log"
	"time"

	"github.com/sajadblnyn/disroom/config"
	"github.com/sajadblnyn/disroom/internal/handler"
	"github.com/sajadblnyn/disroom/internal/repository"
	"github.com/sajadblnyn/disroom/internal/service"
)

func main() {
	err := config.Initialize()
	if err != nil {
		log.Fatalf("error in initializing config  file : %s", err.Error())
	}

	config.InitRedis()
	config.InitNATS()

	time.Sleep(5 * time.Second)
	repository.CreateStream()

	workerCount := config.GetMessagesSubscribersWorkersCount()
	go service.RunMessagesSubscribers(workerCount)
	go service.RunPresenceMessagesSubscribers(workerCount)

	handler.StartServer()
}
