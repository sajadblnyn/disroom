package main

import (
	"github.com/sajadblnyn/disroom/config"
	"github.com/sajadblnyn/disroom/internal/handler"
	"github.com/sajadblnyn/disroom/internal/repository"
	"github.com/sajadblnyn/disroom/internal/service"
)

func main() {
	config.InitRedis()
	config.InitNATS()
	repository.CreateStream()

	go service.RunMessagesSubscribers()
	handler.StartServer()
}
