package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sajadblnyn/disroom/config"
	"github.com/sajadblnyn/disroom/internal/model"
	"github.com/sajadblnyn/disroom/internal/repository"
)

var ctx = context.Background()

func StartServer() {
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
	defer cleanupConnection(ctx, userID, roomID, sub, stopChan)

	fmt.Fprintln(conn, "Welcome to the DisRoom")
	fmt.Fprintln(conn, "Available commands: join <room_id>, send <message>, users, history, exit")

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
			if userID != "" && roomID != "" {
				repository.RemoveUserFromRoom(ctx, userID, roomID)
			}
			if sub != nil {
				sub.Unsubscribe()
			}
			newRoomID, newUserID, err := handleJoin(sub, conn, reader, args)
			if err != nil {
				fmt.Fprintf(conn, "Error: %v\n", err)
				continue
			}

			userID = newUserID
			roomID = newRoomID
			go sendPresenceUpdates(roomID, stopChan)

		case "send":
			if roomID == "" || userID == "" {
				fmt.Fprintln(conn, "Join a room first")
				continue
			}
			handleSend(conn, args, userID, roomID)
		case "users":
			handleUsers(conn, roomID)
		case "history":
			handleHistory(conn, roomID)
		default:
			fmt.Fprintln(conn, "Unknown command")
		}
	}
}

func cleanupConnection(ctx context.Context, userID, roomID string, sub *nats.Subscription, stopChan chan struct{}) {
	if userID != "" && roomID != "" {
		repository.RemoveUserFromRoom(ctx, userID, roomID)
	}
	if sub != nil {
		sub.Unsubscribe()
	}
	close(stopChan)
}

func sendPresenceUpdates(roomID string, stopChan chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			if roomID != "" {
				users, _ := repository.GetActiveUsers(context.Background(), roomID)
				config.JetStream.Publish(fmt.Sprintf("room.%s_presence", roomID), []byte(strings.Join(users, ",")))
			}
		}
	}
}

func handleJoin(sub *nats.Subscription, conn net.Conn, reader *bufio.Reader, args []string) (string, string, error) {
	if len(args) != 2 {
		fmt.Fprintln(conn, "Usage: join <room_id>")
		return "", "", fmt.Errorf("invalid arguments")
	}

	newRoomID := args[1]
	fmt.Fprint(conn, "Enter your user ID: ")
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("read error: %w", err)
	}
	newUserID := strings.TrimSpace(userInput)

	if err := repository.AddUserToRoom(context.Background(), newUserID, newRoomID); err != nil {
		return "", "", fmt.Errorf("join error: %w", err)
	}

	subject := fmt.Sprintf("room.%s", newRoomID)
	sub, err = config.JetStream.Subscribe(subject, func(msg *nats.Msg) {
		var m model.Message
		if err := json.Unmarshal(msg.Data, &m); err != nil {
			return
		}
		fmt.Fprintf(conn, "[%s] %s: %s\n", m.Timestamp.Format("15:04:05"), m.UserID, m.Content)
		msg.Ack()
	}, nats.DeliverNew(), nats.AckExplicit())
	if err != nil {
		return "", "", fmt.Errorf("subscription error: %w", err)
	}

	presenceSubject := fmt.Sprintf("room.%s_presence", newRoomID)
	_, err = config.JetStream.Subscribe(presenceSubject, func(msg *nats.Msg) {
		fmt.Fprintf(conn, "PRESENCE| %s\n", string(msg.Data))
	})
	if err != nil {
		return "", "", fmt.Errorf("presence subscription error: %w", err)
	}

	fmt.Fprintf(conn, "Joined room %s as %s\n", newRoomID, newUserID)
	return newRoomID, newUserID, nil
}

func handleSend(conn net.Conn, args []string, userID, roomID string) {
	if len(args) < 2 {
		fmt.Fprintln(conn, "Usage: send <message>")
		return
	}

	msg := model.Message{
		RoomID:    roomID,
		UserID:    userID,
		Content:   strings.Join(args[1:], " "),
		Timestamp: time.Now(),
	}
	if err := repository.PublishMessage(msg); err != nil {
		fmt.Fprintf(conn, "Error sending message: %v\n", err)
	}
}

func handleUsers(conn net.Conn, roomID string) {
	if roomID == "" {
		fmt.Fprintln(conn, "Join a room first")
		return
	}

	users, err := repository.GetActiveUsers(context.Background(), roomID)
	if err != nil {
		fmt.Fprintf(conn, "Error: %v\n", err)
		return
	}
	fmt.Fprintf(conn, "Users in %s: %s\n", roomID, strings.Join(users, ", "))
}

func handleHistory(conn net.Conn, roomID string) {
	if roomID == "" {
		fmt.Fprintln(conn, "Join a room first")
		return
	}

	history, err := repository.RetrieveMessageHistory(roomID)
	if err != nil {
		fmt.Fprintf(conn, "Error: %v\n", err)
		return
	}

	for _, msg := range history {
		fmt.Fprintf(conn, "[%s] %s: %s\n",
			msg.Timestamp.Format("2006-01-02 15:04:05"),
			msg.UserID,
			msg.Content)
	}
}
