# DisRoom - Distributed Chat Room System

A real-time chat system with distributed messaging capabilities using Go, Redis, and NATS JetStream.

## Features

- Real-time message broadcasting
- Multiple chat rooms support
- User presence tracking
- Message history retrieval
- Active users listing
- NATS-based message persistence
- Redis-backed user presence management

## System Architecture

```mermaid
graph TD
    Client[TCP Client] -->|Connects| GoServer[Go TCP Server]
    GoServer -->|Stores/Loads| Redis[(Redis)]
    GoServer -->|Publishes/Consumes| NATS[NATS JetStream]
    
    subgraph Data Storage
        Redis -.->|Active Users<br>room:*:users| Users[User Presence]
        NATS -.->|Persistent Messages<br>room.*| Messages[Message History]
    end

    style GoServer fill:#74b9ff,stroke:#0984e3
    style Redis fill:#ff7675,stroke:#d63031
    style NATS fill:#55efc4,stroke:#00b894
