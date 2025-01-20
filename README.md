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

%% System Architecture Diagram
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
