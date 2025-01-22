package repository_test

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
	"github.com/stretchr/testify/require"
)

func setupNATS(t *testing.T) (*server.Server, *nats.Conn, nats.JetStreamContext) {
	t.Helper()

	opts := &test.DefaultTestOptions
	opts.JetStream = true
	opts.Port = -1
	opts.StoreDir = t.TempDir()

	s := test.RunServer(opts)

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}

	js, err := nc.JetStream(nats.MaxWait(10 * time.Second))
	if err != nil {
		t.Fatalf("Failed to create JetStream context: %v", err)
	}

	return s, nc, js
}
func TestNATSRepository(t *testing.T) {
	s, nc, js := setupNATS(t)
	defer func() {
		nc.Close()
		s.Shutdown()
	}()

	config.NATSConn = nc
	config.JetStream = js

	t.Run("Publish and Retrieve Messages", func(t *testing.T) {
		repository.CreateStream()

		testMsg := model.Message{
			RoomID:    "test-room",
			UserID:    "user-123",
			Content:   "Test message",
			Timestamp: time.Now().UTC().Truncate(time.Millisecond), // Ensure timestamp precision matches
		}

		err := repository.PublishMessageToRoom(testMsg)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		history, err := repository.RetrieveMessageHistory("test-room")
		require.NoError(t, err)
		require.Len(t, history, 1, "Should retrieve exactly one message")

		require.Equal(t, testMsg.Content, history[0].Content, "Message content mismatch")

		expectedJSON, _ := json.Marshal(testMsg)
		actualJSON, _ := json.Marshal(history[0])
		require.JSONEq(t, string(expectedJSON), string(actualJSON), "Full message JSON mismatch")
	})

	t.Run("Empty History Retrieval", func(t *testing.T) {
		history, err := repository.RetrieveMessageHistory("non-existent-room")
		require.NoError(t, err, "Should handle missing rooms gracefully")
		require.Empty(t, history, "Should return empty slice for non-existent room")
	})
}
