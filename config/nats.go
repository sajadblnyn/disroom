package config

import (
	"log"
	"strings"

	"github.com/nats-io/nats.go"
)

var (
	NATSConn  *nats.Conn
	JetStream nats.JetStreamContext
)

func InitNATS() {

	var err error
	NATSConn, err = nats.Connect(strings.Join(GetNATSURLs(), ","), nats.MaxReconnects(-1))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	JetStream, err = NATSConn.JetStream()
	if err != nil {
		log.Fatalf("Failed to initialize JetStream: %v", err)
	}
}
