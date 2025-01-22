package service

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/sajadblnyn/disroom/config"
	"github.com/sajadblnyn/disroom/internal/model"
	"github.com/sajadblnyn/disroom/internal/repository"
)

func RunMessagesSubscribers() {
	for i := 0; i < 5; i++ {
		_, err := config.JetStream.QueueSubscribe("room.global_messages", "message-processor", func(msg *nats.Msg) {
			var message model.Message
			if err := json.Unmarshal(msg.Data, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				return
			}
			if err := repository.PublishMessageToRoom(message); err != nil {
				log.Printf("Failed to publish message: %v", err)
			}
		})
		if err != nil {
			log.Printf("Failed to subscribe to messages: %v", err)
			continue
		}
	}
}
