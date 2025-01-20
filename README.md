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

## System Architecture Diagram

```mermaid
graph TD
    subgraph Clients
        C1[Client 1]
        C2[Client 2]
        Cn[Client N]
    end

    subgraph Go Chat Server
        GS[TCP Server :8080]
        GS -->|Handle Connections| HC[Connection Handler]
        HC -->|User Commands| P[Protocol Processor]
    end

    subgraph Data Layer
        GS -->|Store/Retrieve Active Users| R[(Redis)]
        GS -->|Publish/Subscribe Messages| NATS
    end

    subgraph NATS Cluster
        NATS{NATS JetStream}
        NATS -->|Stream Persistence| STORAGE[(File Storage)]
        N1[NATS Node 1]
        N2[NATS Node 2]
        N3[NATS Node 3]
    end

    C1 -->|TCP| GS
    C2 -->|TCP| GS
    Cn -->|TCP| GS
    NATS -.- N1
    NATS -.- N2
    NATS -.- N3
```

## 🧰 Components

### Core System Elements
- **Go Application Server**  
  - TCP listener (:8080)  
  - Command processor (join/send/users/history/exit)  
  - Redis client integration  
  - NATS JetStream client  
  - Connection handler

- **Redis Database**  
  - Stores active users per room using Sets  
  - Key format: `room:<room_id>:users`  
  - Handles real-time presence updates

- **NATS JetStream**  
  - Persistent message streaming  
  - Stream name: `ChatRooms`  
  - Subjects: `room.*` (wildcard per room)  
  - Message retention policy: File storage

- **TCP Client**  
  - User interface via netcat/telnet  
  - Simple text-based interaction

## ⚙️ Installation

### Prerequisites
- Go 1.19+
- Redis Server
- NATS Server v2.9+

```bash
# 1. Clone repository
git clone https://github.com/yourusername/disroom.git
cd disroom

# 2. Install dependencies
go mod tidy

# 3. Start Redis
redis-server --port 6379

# 4. Start NATS cluster (3-node example)
nats-server -p 4222 -cluster nats://127.0.0.1:6222 -routes nats://127.0.0.1:6222,nats://127.0.0.1:6223 &
nats-server -p 4223 -cluster nats://127.0.0.1:6223 -routes nats://127.0.0.1:6222,nats://127.0.0.1:6224 &
nats-server -p 4224 -cluster nats://127.0.0.1:6224 -routes nats://127.0.0.1:6222,nats://127.0.0.1:6223 &

# 5. Build and run server
go build -o disroom && ./disroom
```
## Infrastructure Characteristics

The system architecture is designed with the following key characteristics:

- **Horizontal Scalability**  
  _NATS Cluster Scaling_: The NATS cluster can elastically scale to handle increased message throughput, supporting dynamic addition/removal of nodes while maintaining consistent message delivery.

- **Fault Tolerance**  
  _Message Redundancy_: NATS clustering provides automatic message replication across nodes, ensuring continuous availability even during node failures.  
  _Automatic Failover_: Built-in Raft consensus protocol maintains cluster coordination and leadership election.

- **Persistence**  
  _Durable Message Storage_: JetStream persists messages to disk with configurable retention policies (time-based, size-based, or interest-based).  
  _Crash Recovery_: Guaranteed message durability through Write-Ahead Logging (WAL) and checksum verification.

- **Real-time Updates**  
  _Instant Presence Tracking_: Redis-backed user presence system provides sub-millisecond response times for:  
  • User join/leave operations  
  • Active user listings  
  • Presence heartbeat updates  
  _Cluster Synchronization_: Redis pub/sub channels maintain consistent presence state across server instances.

- **Lightweight Protocol**  
  _TCP Efficiency_: Binary-based plain TCP protocol minimizes overhead compared to HTTP-based alternatives.  
  _Broad Compatibility_: Simple text-based command structure supports integration with:  
  • Terminal clients  
  • GUI applications  
  • IoT devices  
  • WebSocket gateways  
  _Connection Resilience_: Built-in reconnection logic handles network interruptions gracefully.
