# Technical Design Document: Chess Broadcast Engine

## 1. System Overview
The **Chess Broadcast Engine** is a unidirectional, real-time data streaming platform. The primary goal is to accept chess moves from a Grandmaster's client and route those moves to thousands of active spectators with microsecond latency, while guaranteeing strict chronological order and preventing data loss during new spectator connection sequences (hydration).

---

## 2. Component Architecture
The system is decoupled into isolated, independently scalable microservices.

### The Ingest Node (gRPC)
Written in Go, this node serves as the entry point. It receives `Move` payloads via gRPC, ensuring strict typing via Protocol Buffers. It is responsible for atomic ordering, caching, and publishing. It does not communicate directly with end-users.

### The Message Broker (RabbitMQ)
Acting as the decoupled nervous system, an AMQP **Fanout Exchange** receives validated moves from the Ingest Node. It allows the backend to fan out data to an arbitrary number of downstream Broadcaster nodes without adding latency or coupling to the ingestion layer.

### The Broadcast Node (WebSockets)
A highly concurrent Go service utilizing the `gorilla/websocket` library. It subscribes to the AMQP queue, maintains an internal map of active WebSocket connections (grouped by `match_id` rooms), and pushes JSON payloads to connected spectators.

### The State Store (Redis)
Serves two highly specialized roles:
* **Atomic Sequence Generator:** Uses Redis `INCR` to assign an absolute chronological integer to every incoming move, guaranteeing ordered state across horizontally scaled Ingest Nodes.
* **Event-Sourced Replay Buffer:** Uses Redis `RPUSH` to maintain an append-only log of the match history. Newly connected spectators fetch this array (`LRANGE`) to instantly construct the current state of the board.

### Observability Layer (OpenTelemetry / Jaeger)
Every component is instrumented using OpenTelemetry. Traces span from the gRPC request, through the Redis network call, across the RabbitMQ boundary, and into the WebSocket hub.

---

## 3. Data Flow & State Synchronization

> **Architectural Challenge: "The Interleave" Race Condition.** > When a user connects to a live data stream, there is a microsecond gap between fetching historical data and subscribing to the live stream. If a move occurs during this gap, the state becomes fractured.

We solved this using the **Subscribe -> Fetch** pattern and atomic counters.

### Flow Execution
1. **Move Ingestion:** * A move hits the Ingest Node.
   * The node requests an atomic sequence number: `INCR match:{id}:sequence`.
   * The returned integer is stamped onto the Protobuf payload.
2. **Caching & Publishing:** * The node uses a Redis Pipeline to `RPUSH` the serialized protobuf into the `match:{id}:latest` list and set an expiration.
   * The node publishes the payload to RabbitMQ.
3. **Spectator Hydration (The Fix):** * When a spectator connects to the Broadcast Node, the HTTP handler **first** registers the client to the WebSocket Hub room. Live messages immediately begin queuing in the client's buffered channel.
   * **Second**, the handler fetches the full historical event log from Redis (`LRANGE`).
   * The history is pushed into the buffer *behind* any live messages that may have arrived.
   * **Result:** The client receives an array of moves containing a mix of live and cached data. The frontend client utilizes the guaranteed `sequence_number` to sort and deduplicate the payload idempotently.

---

## 4. W3C Trace Context Propagation
To trace the lifecycle of a move across a distributed network, the engine employs W3C Trace Context propagation. 

* **Injection:** Before publishing to RabbitMQ, the Ingest Node extracts the `traceparent` from the active Go context and injects it into a custom carrier wrapped around the `amqp091.Table` headers.
* **Extraction:** The Broadcast Node's consumer loop extracts the headers from the incoming AMQP message, reconstructs the Go context, and continues the trace timeline. This bridges the physical network gap in the Jaeger UI waterfall.

---

## 5. Deployment & Containerization
The system is deployed via **Docker Compose**, acting as a production-grade local environment.

* **Multi-Stage Builds:** The Go binaries are compiled in a `golang:alpine` builder stage (`CGO_ENABLED=0`) and transferred to a minimal `alpine` runner stage. This reduces the image footprint from ~800MB to ~20MB and eliminates compiler attack surfaces.
* **Service Dependency:** Compose `healthcheck` directives ensure the Go nodes do not boot until RabbitMQ and Redis are fully initialized and accepting TCP connections.