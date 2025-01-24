package service_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/sajadblnyn/disroom/config"
	"github.com/sajadblnyn/disroom/internal/model"
	"github.com/sajadblnyn/disroom/internal/repository"
	"github.com/sajadblnyn/disroom/internal/service"
	"github.com/stretchr/testify/require"
)

func startJetStreamServer() *server.Server {
	opts := test.DefaultTestOptions
	opts.JetStream = true
	opts.Port = -1
	return test.RunServer(&opts)
}

func TestMessageService(t *testing.T) {
	s := startJetStreamServer()
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	js, err := nc.JetStream(nats.MaxWait(2 * time.Second))
	require.NoError(t, err)

	config.NATSConn = nc
	config.JetStream = js
	repository.CreateStream()

	t.Run("Message Routing", func(t *testing.T) {
		service.RunMessagesSubscribers(5)

		testMsg := model.Message{
			RoomID:    "routed-room",
			UserID:    "user-789",
			Content:   "Test message",
			Timestamp: time.Now(),
		}

		err = repository.PublishMessage(testMsg)
		require.NoError(t, err)

		sub, err := js.PullSubscribe("room.routed-room", "test-sub")
		require.NoError(t, err)

		msgs, err := sub.Fetch(1, nats.MaxWait(2*time.Second))
		require.NoError(t, err)
		require.Len(t, msgs, 1)

		var received model.Message
		require.NoError(t, json.Unmarshal(msgs[0].Data, &received))
		require.Equal(t, testMsg.Content, received.Content)
	})

}
