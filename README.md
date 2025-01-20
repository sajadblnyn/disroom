## System Architecture

```mermaid
graph TD
    %% Clients
    ClientA[Client] -->|TCP| GoServer
    ClientB[Client] -->|TCP| GoServer
    
    %% Main Components
    GoServer[Go Server<br>disroom:port] -->|"SET/GET user presence"| Redis[(Redis<br>redis:6379)]
    GoServer -->|"PUB/SUB messages"| NATS[NATS JetStream<br>nats:4222]
    NATS -->|Persist messages| Storage[(File Storage)]
    
    %% Internal Components
    subgraph Docker Network[Containerized Services]
        GoServer
        Redis
        NATS
        Storage
    end
    
    %% Data Flow
    GoServer -->|"Periodic presence updates<br>(every 30s)"| Redis
    NATS -->|"Message history<br>retrieval"| GoServer
    NATS -->|"Stream replication"| NATS_Replica[NATS Node]
    
    %% Administration
    Admin[Admin] -->|Monitoring| NATS_Monitor[NATS Monitor<br>8222]
    
    %% Styles
    classDef client fill:#e1f5fe,stroke:#039be5;
    classDef service fill:#f0f4c3,stroke:#afb42b;
    classDef storage fill:#dcedc8,stroke:#689f38;
    classDef queue fill:#ffcdd2,stroke:#e53935;
    classDef admin fill:#f3e5f5,stroke:#8e24aa;
    
    class ClientA,ClientB client;
    class GoServer,Redis,NATS service;
    class Storage storage;
    class NATS_Replica queue;
    class Admin,NATS_Monitor admin;
