# chess-broadcast-engine

A distributed backend for broadcasting live chess matches to thousands of simultaneous spectators (similar to Twitch's telemetry backend). 


**The 6-Week Blueprint:**

* **Week 1: gRPC & Protocol Buffers (The Ingest Node).** Leaving HTTP/JSON behind. Writing `.proto` files to define strict contracts. Building a gRPC server that accepts and validates chess moves.

* **Week 2: The Message Broker (RabbitMQ/Redis).** Decoupling the system. When a move is validated, the Ingest Node publishes it to an exchange/topic rather than processing it directly.

* **Week 3: The Distributed WebSocket Hub (The Broadcast Nodes).** Scaling Phase 2. Building multiple worker nodes that subscribe to the Message Broker and push real-time updates to connected WebSocket clients.

* **Week 4: Distributed State & Caching.** Storing the current state of the chessboard in Redis so newly connected spectators instantly receive the current board without querying a primary database.

* **Week 5: Observability & Tracing.** Integrating OpenTelemetry. Tracing a single move as it travels from the gRPC Ingest Node -> RabbitMQ -> Broadcast Node to detect latency bottlenecks.

* **Week 6: Infrastructure Orchestration.** Writing advanced `docker-compose.yml` (or local Kubernetes manifests) to spin up the entire distributed cluster (Ingest, Broadcast replicas, Redis, RabbitMQ) with a single command.


## Current Status

* **Current Week:** 4
* **Current Task:** The Redis Cache Layer - Spin up a local Redis instance using Docker Compose. Create a strict Cache interface in Ingest Node, and implement a Redis client that updates the latest move in memory every time a move is recorded.