package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
)

var (
	rdb *redis.Client
	nc  *nats.Conn
	js  nats.JetStreamContext
	ctx = context.Background()
)

type Message struct {
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func initNATS() {
	var err error
	serverURLs := []string{
		"nats://127.0.0.1:4222",
		"nats://127.0.0.1:4223",
		"nats://127.0.0.1:4224",
	}

	nc, err = nats.Connect(strings.Join(serverURLs, ","), nats.MaxReconnects(-1))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	js, err = nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to initialize JetStream: %v", err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "ChatRooms",
		Subjects: []string{"room.*"},
		Storage:  nats.FileStorage,
	})
	if err != nil && !strings.Contains(err.Error(), "stream already exists") {
		log.Fatalf("Failed to create stream: %v", err)
	}
}

func addUserToRoom(ctx context.Context, userID, roomID string) error {
	return rdb.SAdd(ctx, fmt.Sprintf("room:%s:users", roomID), userID).Err()
}

func removeUserFromRoom(ctx context.Context, userID, roomID string) error {
	return rdb.SRem(ctx, fmt.Sprintf("room:%s:users", roomID), userID).Err()
}

func getActiveUsers(ctx context.Context, roomID string) ([]string, error) {
	return rdb.SMembers(ctx, fmt.Sprintf("room:%s:users", roomID)).Result()
}

func publishMessageToRoom(msg Message) {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	if _, err = js.Publish(fmt.Sprintf("room.%s", msg.RoomID), msgJSON); err != nil {
		log.Printf("Failed to publish message: %v", err)
	}
}
func publishMessage(msg Message) {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	if _, err = js.Publish("room.global_messages", msgJSON); err != nil {
		log.Printf("Failed to publish message: %v", err)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	var (
		userID   string
		roomID   string
		sub      *nats.Subscription
		stopChan = make(chan struct{})
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer func() {
		if userID != "" && roomID != "" {
			removeUserFromRoom(ctx, userID, roomID)
		}
		if sub != nil {
			sub.Unsubscribe()
		}
		close(stopChan)
	}()

	fmt.Fprintln(conn, "Welcome to the DisRoom")
	fmt.Fprintln(conn, "Available commands: join <room_id>, send <message>, users, history, exit")

	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				if roomID != "" {
					users, _ := getActiveUsers(ctx, roomID)
					js.Publish(fmt.Sprintf("room.%s.presence", roomID), []byte(strings.Join(users, ",")))
				}
				time.Sleep(30 * time.Second)
			}
		}
	}()

	for {
		fmt.Fprint(conn, "> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Connection error: %v", err)
			return
		}

		command = strings.TrimSpace(command)
		if command == "exit" {
			fmt.Fprintln(conn, "Goodbye!")
			return
		}

		args := strings.Fields(command)
		if len(args) < 1 {
			fmt.Fprintln(conn, "Invalid command")
			continue
		}

		switch args[0] {
		case "join":
			if len(args) != 2 {
				fmt.Fprintln(conn, "Usage: join <room_id>")
				continue
			}

			newRoomID := args[1]
			if newRoomID == roomID {
				continue
			}

			fmt.Fprint(conn, "Enter your user ID: ")
			userInput, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}
			newUserID := strings.TrimSpace(userInput)

			if sub != nil {
				sub.Unsubscribe()
				removeUserFromRoom(ctx, userID, roomID)
			}

			if err := addUserToRoom(ctx, newUserID, newRoomID); err != nil {
				fmt.Fprintf(conn, "Join error: %v\n", err)
				continue
			}

			subject := fmt.Sprintf("room.%s", newRoomID)
			sub, err = js.Subscribe(subject, func(msg *nats.Msg) {
				var m Message
				if err := json.Unmarshal(msg.Data, &m); err != nil {
					return
				}
				fmt.Fprintf(conn, "[%s] %s: %s\n",
					m.Timestamp.Format("15:04:05"), m.UserID, m.Content)
				msg.Ack()
			}, nats.DeliverNew(), nats.AckExplicit())

			if err != nil {
				fmt.Fprintf(conn, "Subscription error: %v\n", err)
				removeUserFromRoom(ctx, newUserID, newRoomID)
				continue
			}

			presenceSubject := fmt.Sprintf("room.%s.presence", newRoomID)
			sub, err = nc.Subscribe(presenceSubject, func(msg *nats.Msg) {
				fmt.Fprintf(conn, "PRESENCE| %s\n", string(msg.Data))
			})
			if err != nil {
				fmt.Fprintf(conn, "Presence error: %v\n", err)
				continue
			}

			userID = newUserID
			roomID = newRoomID
			fmt.Fprintf(conn, "Joined room %s as %s\n", roomID, userID)

		case "send":
			if roomID == "" || userID == "" {
				fmt.Fprintln(conn, "Join a room first")
				continue
			}
			if len(args) < 2 {
				fmt.Fprintln(conn, "Usage: send <message>")
				continue
			}

			msg := Message{
				RoomID:    roomID,
				UserID:    userID,
				Content:   strings.Join(args[1:], " "),
				Timestamp: time.Now(),
			}
			publishMessage(msg)

		case "users":
			if roomID == "" {
				fmt.Fprintln(conn, "Join a room first")
				continue
			}
			users, err := getActiveUsers(ctx, roomID)
			if err != nil {
				fmt.Fprintf(conn, "Error: %v\n", err)
			} else {
				fmt.Fprintf(conn, "Users in %s: %s\n", roomID, strings.Join(users, ", "))
			}

		case "history":
			if roomID == "" {
				fmt.Fprintln(conn, "Join a room first")
				continue
			}
			msgs, err := retrieveMessageHistory(roomID)
			if err != nil {
				fmt.Fprintf(conn, "Error: %v\n", err)
			} else {
				for _, m := range msgs {
					fmt.Fprintf(conn, "[%s] %s: %s\n",
						m.Timestamp.Format("15:04:05"), m.UserID, m.Content)
				}
			}

		default:
			fmt.Fprintln(conn, "Unknown command")
		}
	}
}

func retrieveMessageHistory(roomID string) ([]Message, error) {
	subject := fmt.Sprintf("room.%s", roomID)
	sub, err := js.PullSubscribe(subject, "history-"+roomID)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	msgs, err := sub.Fetch(100, nats.MaxWait(2*time.Second))
	if err != nil && err != nats.ErrTimeout {
		return nil, err
	}

	var history []Message
	for _, msg := range msgs {
		var m Message
		if err := json.Unmarshal(msg.Data, &m); err != nil {
			continue
		}
		history = append(history, m)
		msg.Ack()
	}
	return history, nil
}

func startServer() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Server listening on :8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConnection(conn)
	}

}

func runMessagesSubscribers() {

	for i := 0; i < 5; i++ {
		_, err := js.QueueSubscribe("room.global_messages", "message-processor", func(msg *nats.Msg) {
			var message Message
			if err := json.Unmarshal(msg.Data, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				return
			}
			publishMessageToRoom(message)
		})
		if err != nil {
			log.Printf("Failed to subscribe to messages: %v", err)
			continue
		}

	}

}

func main() {
	initRedis()
	initNATS()

	go runMessagesSubscribers()
	startServer()

}
