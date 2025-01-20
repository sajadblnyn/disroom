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
        C1[Client]
        C2[Client]
        Cn[Client]
    end

    subgraph LoadBalancer[Load Balancer]
        LB[HAProxy/NGINX]
    end

    subgraph GoServerCluster[Go Server Cluster]
        GS1[Instance 1]
        GS2[Instance 2]
        GSn[Instance N]
    end

    subgraph RedisCluster[Redis Cluster]
        R1[Primary]
        R2[Replica]
        R3[Replica]
    end

    subgraph NATSCluster[NATS Cluster]
        N1[Node 1]
        N2[Node 2]
        N3[Node 3]
    end

    C1 -->|TCP| LB
    C2 -->|TCP| LB
    Cn -->|TCP| LB
    
    LB -->|Distribute| GS1
    LB -->|Connections| GS2
    LB -->|Across Cluster| GSn
    
    GS1 -->|Presence Data| RedisCluster
    GS2 -->|Presence Data| RedisCluster
    GSn -->|Presence Data| RedisCluster
    
    GS1 -->|Pub/Sub| NATSCluster
    GS2 -->|Pub/Sub| NATSCluster
    GSn -->|Pub/Sub| NATSCluster

    N1 <-->|Raft Consensus| N2
    N2 <-->|Raft Consensus| N3
    N3 <-->|Raft Consensus| N1

    R1 <-->|Replication| R2
    R1 <-->|Replication| R3

    classDef cluster fill:#f9f9f9,stroke:#999,stroke-width:2px;
    classDef component fill:#e6f3ff,stroke:#3399ff;
    classDef storage fill:#ffe6e6,stroke:#ff6666;
    classDef queue fill:#e6ffe6,stroke:#33cc33;
    
    class Clients,GoServerCluster,NATSCluster,RedisCluster cluster;
    class LB,GS1,GS2,GSn component;
    class R1,R2,R3 storage;
    class N1,N2,N3 queue;
