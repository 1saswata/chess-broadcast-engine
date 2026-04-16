# Chess Broadcast Engine

## Overview
The **Chess Broadcast Engine** is a high-throughput, distributed telemetry router designed to broadcast real-time chess matches to thousands of simultaneous spectators. It ingests strict protobuf telemetry, decouples processing via message brokering, synchronizes state using an event-sourced replay buffer, and streams updates to authenticated clients over highly concurrent WebSockets.

---

## Key Features
* **Zero-Trust Security:** Secured via stateless JSON Web Tokens (JWT). A gRPC Unary Interceptor protects the ingestion perimeter, while query-parameter handshakes secure the WebSocket upgrades.
* **Visual Telemetry Client:** Includes a vanilla JavaScript frontend that connects to the stream, implements a client-side sequence buffer to handle out-of-order network anomalies, and renders live state using `chessboard.js`.
* **Strict Contracts:** Ingestion via **gRPC** and **Protocol Buffers** ensures strict type safety and high-performance serialization.
* **Event-Driven Architecture:** Utilizes **RabbitMQ** (AMQP Fanout Exchanges) to completely decouple the ingestion layer from the broadcasting layer.
* **Event-Sourced Replay Buffer:** Leverages **Redis Lists** and **Atomic Counters** (`INCR`) to guarantee chronological state hydration and prevent distributed race conditions.
* **Concurrent Broadcasting:** A thread-safe **Gorilla WebSocket Hub** pushes real-time updates to connected clients using Go's lightweight goroutines and buffered channels.
* **Enterprise Observability:** Fully instrumented with **OpenTelemetry** and **Jaeger**. W3C Trace Context is manually propagated across the AMQP network boundary for distributed waterfall tracing.

---

## Architecture Deep Dive
For a comprehensive breakdown of the system design, data flow, concurrency models, and state synchronization strategies (including how the engine handles WebSocket hydration race conditions), please see the [ARCHITECTURE.md](./ARCHITECTURE.md) file.

---

## Quick Start

### Prerequisites
* Docker and Docker Compose
* Modern Web Browser

### Bootstrapping the Cluster
Run the following command to spin up the entire distributed cluster (Ingest Node, Broadcast Node, RabbitMQ, Redis, and Jaeger):
```bash
docker compose up --build -d```

### Endpoints

* **Visual Frontend UI:** http://localhost:8081

* **gRPC Ingest Node:** localhost:50051

* **HTTP Token Dispenser:** localhost:8080/login

* **Jaeger UI:** http://localhost:16686

## Project Scope & Known Limitations
**Technical Debt & Scope Acknowledgment: As a portfolio piece focusing strictly on backend distributed systems and real-time data streaming, this project explicitly bounds its scope. The following elements represent known limitations and areas for future expansion:

**No Visual Frontend Client:** Currently, the system hydrates state via raw JSON over WebSockets. A frontend UI (e.g., vanilla JS with an HTML5 canvas) is required to parse the JSON and visually render the chessboard.

**No Cold Storage / Archival:** The active match state is stored ephemerally in Redis. Once a match concludes, the event log is lost. A persistent database (like PostgreSQL) and an asynchronous dead-letter worker are needed to archive finished games.

**Resilience Engineering:** The architecture currently assumes infrastructure health. Implementing the Circuit Breaker Pattern and Exponential Backoff Retries for Redis and RabbitMQ connections would drastically improve fault tolerance during network partitions.


---