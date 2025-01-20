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
%% Enhanced System Architecture Diagram
graph TD
    subgraph Clients
        C1[Client]
        C2[Client]
        Cn[Client]
    end

    subgraph Load Balancer
        LB[HAProxy/NGINX\nLoad Balancer\nPort 8080]
    end

    subgraph Go Chat Server Cluster
        GS1[Go Server\nInstance 1]
        GS2[Go Server\nInstance 2]
        GSn[Go Server\nInstance N]
    end

    subgraph Data Layer
        R[(Redis\nSingle Primary\n+ Replicas)]
        NATS{NATS JetStream Cluster}
    end

    subgraph NATS Cluster
        N1[NATS Node 1\nnats://host1:4222]
        N2[NATS Node 2\nnats://host2:4223]
        N3[NATS Node 3\nnats://host3:4224]
        N1 <-->|Raft Consensus| N2
        N2 <-->|Raft Consensus| N3
        N3 <-->|Raft Consensus| N1
    end

    C1 -->|TCP| LB
    C2 -->|TCP| LB
    Cn -->|TCP| LB
    
    LB -->|TCP Connections| GS1
    LB -->|TCP Connections| GS2
    LB -->|TCP Connections| GSn
    
    GS1 -->|User Presence| R
    GS1 -->|Pub/Sub| NATS
    GS2 -->|User Presence| R
    GS2 -->|Pub/Sub| NATS
    GSn -->|User Presence| R
    GSn -->|Pub/Sub| NATS

    NATS -->|Stream Replication| N1
    NATS -->|Stream Replication| N2
    NATS -->|Stream Replication| N3

    classDef cluster fill:#f9f9f9,stroke:#999,stroke-width:2px;
    classDef component fill:#e6f3ff,stroke:#3399ff,stroke-width:2px;
    classDef storage fill:#ffe6e6,stroke:#ff6666,stroke-width:2px;
    classDef queue fill:#e6ffe6,stroke:#33cc33,stroke-width:2px;
    
    class Clients,Go Chat Server Cluster,NATS Cluster cluster;
    class LB,GS1,GS2,GSn component;
    class R storage;
    class NATS,N1,N2,N3 queue;
