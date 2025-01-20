# DisRoom - Distributed Chat System

![Architecture Diagram](https://mermaid.ink/svg/pako:eNplkE1vgzAMhv_K5XQ7oNpJ2g7s0AOkHXqYVFWqJIIkRlCJqkri36eQj6nbwZaf53V4bYwVcGFSmYF-7Y0K0F6U1gXo1s0yY2K_3ZqF6lUe7dFZ5Bp4iRq8Q1NQZ7Dc7M2sOQ8h5H5H6jH3c0oHlB6HfUQp1O0N5FqD9lOaF1eQvQYdO5Lv7O6lQ3E5mDlFqA9qX1VqjvK_5JdJZg9VvV9p4XZc6t_1Zk5NsnW0a6dTj0Ht9O5V0VH1fJg7VlM4r6Vh8G3cHc9LhH4lK0t3cQ)

```mermaid
graph TD
    Client[Client] -->|TCP| GoServer[Go Server]
    GoServer -->|User Presence| Redis[(Redis)]
    GoServer -->|Message Streaming| NATS[NATS JetStream]
    NATS -->|Message Persistence| Storage[File Storage]
    
    subgraph Docker Network
        GoServer
        Redis
        NATS
    end
    
    subgraph Clients
        Client
        Client2[Client]
        Client3[Client]
    end