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
* **Current Task:** Event Sourcing & The Replay Buffer - Transition the caching layer from a simple "latest state" Key-Value store to an append-only Event Log using Redis Lists. Update the WebSocket Hub to blast the entire historical array of moves to any newly connected client so their browser can replay the match to the current state.