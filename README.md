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
    Client[TCP Client] -->|1. Connects| GoServer[Go TCP Server]
    GoServer -->|2. Stores/Loads Active Users| Redis[(Redis Database)]
    GoServer -->|3. Publishes Messages| NATS[NATS JetStream Cluster]
    GoServer -->|4. Retrieves History| NATS
    NATS -->|5. Push Messages| GoServer
    GoServer -->|6. Sends Messages| Client

    subgraph Go_Application
        GoServer
        TCP_Listener[TCP Listener<br>:8080]
        Command_Handler[Command Handler<br>join/send/users/history]
        Presence_Manager[Presence Manager<br>(Redis Client)]
        JetStream_Manager[JetStream Manager<br>(NATS Client)]
    end

    subgraph Redis
        Redis -->|User Sets| Room1_Users[room:room1:users]
        Redis -->|User Sets| Room2_Users[room:room2:users]
    end

    subgraph NATS_JetStream
        NATS -->|Streams| Message_Stream[ChatRooms Stream<br>Subjects: room.*]
        NATS -->|Key-Value| Presence_Updates[Presence Updates<br>room.*.presence]
    end

    Client -->|7. User Commands| TCP_Listener
    TCP_Listener -->|8. Routes Requests| Command_Handler
    Command_Handler -->|9. Manages Presence| Presence_Manager
    Command_Handler -->|10. Message Operations| JetStream_Manager
    Presence_Manager -->|11. User Updates| Redis
    JetStream_Manager -->|12. Pub/Sub| NATS

    style Go_Application fill:#1e90ff,stroke:#0000ff
    style Redis fill:#ff6347,stroke:#dc143c
    style NATS_JetStream fill:#3cb371,stroke:#2e8b57
    style Client fill:#f4a460,stroke:#8b4513
    style NATS_JetStream fill:#3cb371,stroke:#2e8b57
    style Client fill:#f4a460,stroke:#8b4513
