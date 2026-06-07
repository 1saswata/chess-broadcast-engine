---

### 2. Updated `ARCHITECTURE.md`

# Technical Design Document: Chess Broadcast Engine

## 1. System Overview
The **Chess Broadcast Engine** is a unidirectional, real-time data streaming platform. The primary goal is to accept authenticated chess moves from a Grandmaster's client and route those moves to thousands of active spectators with microsecond latency. It guarantees strict chronological order, prevents data loss during hydration, enforces strict role-and-color-based access control, and permanently archives completed matches.


---

## 2. Component Architecture
The system is decoupled into isolated, independently scalable microservices.

### The Ingest Node (gRPC & HTTP API)
Written in Go, this node serves as the entry point. It receives `Move` payloads via gRPC, ensuring strict typing via Protocol Buffers. It utilizes a layered gRPC Unary Interceptor pipeline for authentication, rate limiting, and color validation. It also exposes a decoupled HTTP API (`internal/api`) for user registration, admin provisioning, and manual match archival.

### The Message Broker (RabbitMQ)
Acting as the decoupled nervous system, an AMQP **Fanout Exchange** receives validated moves from the Ingest Node. It allows the backend to fan out data to an arbitrary number of downstream Broadcaster nodes without adding latency or coupling to the ingestion layer.

### The Broadcast Node (WebSockets & UI Server)
A highly concurrent Go service utilizing the `gorilla/websocket` library. It subscribes to the AMQP queue, maintains an internal map of active WebSocket connections, and pushes JSON payloads to connected spectators. It also serves the static vanilla HTML/JS client assets.

### The Visual Client (Vanilla JS)
A lightweight frontend using `chessboard.js`. It utilizes the native `fetch` API for authentication and maintains a `Map`-based sequence buffer to idempotently render the board state, resolving out-of-order network anomalies on the client side.

### The State Store (Redis)
Serves specialized, high-performance roles:
* **Atomic Sequence Generator:** Uses `INCR` to assign an absolute chronological integer to every incoming move.
* **Event-Sourced Replay Buffer:** Uses `RPUSH` to maintain an append-only log of the match history for instant spectator hydration.
* **Distributed Rate Limiter:** Executes atomic Lua scripts to manage Token Bucket/Fixed Window quotas.
* **Spatial Authorization Matrix:** Uses Redis Hashes (`HSET`) to lock specific UUIDs to specific piece colors for ultra-fast validation.

### The Persistent Vault (PostgreSQL)
A relational database used for cold storage and source-of-truth identity management. Schema migrations are managed via `golang-migrate` and embedded directly into the Go binary using `go:embed`.

---

## 3. Security, Identity & Infrastructure Defense

The architecture employs a multi-layered security fortress evaluated prior to any business logic execution.

1. **Credential Vault:** Users register via HTTP. Passwords are cryptographically hashed using Bcrypt and stored in PostgreSQL. Upon login, a stateless JSON Web Token (JWT) is minted.
2. **Atomic Rate Limiting:** The gRPC interceptor intercepts requests and delegates a Fixed Window algorithm check to Redis via an embedded Lua script. This guarantees atomicity, preventing distributed race conditions while shielding the system from DDoS bursts.
3. **Admin Provisioning & Color-Locking:** VIP matches must be explicitly scheduled via a protected `/admin/matches` endpoint. This maps Grandmaster UUIDs to `PLAYER_WHITE` or `PLAYER_BLACK` in a Redis Hash. The gRPC interceptor dynamically extracts the piece color from the incoming protobuf payload and cross-references it against the Redis Hash, instantly rejecting spoofed moves.

---

## 4. Data Flow & State Synchronization

> **Architectural Challenge: "The Interleave" Race Condition.** > When a user connects to a live data stream, there is a microsecond gap between fetching historical data and subscribing to the live stream. If a move occurs during this gap, the state becomes fractured.

We solved this using the **Subscribe -> Fetch** pattern and a **Client-Side Sequence Buffer**.

### Flow Execution
1. **Move Ingestion:** * A move hits the Ingest Node, passes the security/rate-limit interceptors.
   * The node requests an atomic sequence number: `INCR match:{id}:sequence`.
   * The returned integer is stamped onto the Protobuf payload.
2. **Caching & Publishing:** * The node uses a Redis Pipeline to `RPUSH` the serialized protobuf into the `match:{id}:latest` list.
   * The node publishes the payload to RabbitMQ.
3. **Spectator Hydration & Client Interleave:** * When a spectator connects, the Broadcaster registers the client to the WebSocket Hub room *first*. Live messages immediately queue in the client's buffered channel.
   * *Second*, the handler fetches the full historical event log from Redis (`LRANGE`) and pushes it behind the live messages.
   * **The Client Resolution:** The JS frontend maintains an `expectedSequence` pointer and a `pendingBuffer` map. If a live move (e.g., Seq 51) arrives before the cached history (Seq 1-50), the client buffers the future move and drops duplicates, guaranteeing perfect chronological rendering.

---

## 5. ACID Transactional Archival

Because Redis is optimized for ephemeral state, completed matches must be archived to PostgreSQL. 
* To bridge the gap between binary transport and text analytics, the archival worker deserializes the Protobuf byte arrays and re-marshals them into native text to leverage PostgreSQL's `JSONB` querying capabilities.
* The archival process is wrapped in a strict SQL Transaction (`sql.Tx`). If the `matches` table row is inserted, but the loop fails while inserting the `moves` rows, the entire operation triggers a `Rollback()`. This guarantees the database is never left in a corrupted or orphaned state.

---

## 6. W3C Trace Context Propagation (OpenTelemetry)
To trace the lifecycle of a move across a distributed network, the engine employs W3C Trace Context propagation. 
* **Injection:** Before publishing to RabbitMQ, the Ingest Node extracts the `traceparent` from the active Go context and injects it into a custom carrier wrapped around the `amqp091.Table` headers.
* **Extraction:** The Broadcast Node's consumer loop extracts the headers from the incoming AMQP message, reconstructs the Go context, and continues the trace timeline. This bridges the physical network gap in the Jaeger UI waterfall.
