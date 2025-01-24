package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sajadblnyn/disroom/config"
	"github.com/sajadblnyn/disroom/internal/model"
)

func PublishMessage(msg model.Message) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = config.JetStream.Publish("room.global_messages", msgJSON)
	return err
}

func PublishPresenceMessage(msg model.Message) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = config.JetStream.Publish("room.presence_messages", msgJSON)
	return err
}
func PublishMessageToRoom(msg model.Message) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = config.JetStream.Publish(fmt.Sprintf("room.%s", msg.RoomID), msgJSON)
	return err
}

func CreateStream() {
	_, err := config.JetStream.AddStream(&nats.StreamConfig{
		Name:     "ChatRooms",
		Subjects: []string{"room.*"},
		Storage:  nats.FileStorage,
	})
	if err != nil && !strings.Contains(err.Error(), "stream already exists") {
		log.Fatalf("Failed to create stream: %v", err)
	}
}

func RetrieveMessageHistory(roomID string) ([]model.Message, error) {
	subject := fmt.Sprintf("room.%s", roomID)
	sub, err := config.JetStream.PullSubscribe(subject, "history-"+roomID)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	msgs, err := sub.Fetch(100, nats.MaxWait(2*time.Second))
	if err != nil && err != nats.ErrTimeout {
		return nil, err
	}

	var history []model.Message
	for _, msg := range msgs {
		var m model.Message
		if err := json.Unmarshal(msg.Data, &m); err != nil {
			continue
		}
		history = append(history, m)
		msg.Ack()
	}
	return history, nil
}
