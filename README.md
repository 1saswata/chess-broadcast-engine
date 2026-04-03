# Chess Broadcast Engine

## Overview
The **Chess Broadcast Engine** is a high-throughput, distributed telemetry router designed to broadcast real-time chess matches to thousands of simultaneous spectators. It ingests strict protobuf telemetry, decouples processing via message brokering, synchronizes state using an event-sourced replay buffer, and streams updates to clients over highly concurrent WebSockets.

---

## Key Features
* **Strict Contracts:** Ingestion via **gRPC** and **Protocol Buffers** ensures strict type safety and high-performance serialization.
* **Event-Driven Architecture:** Utilizes **RabbitMQ** (AMQP Fanout Exchanges) to completely decouple the ingestion layer from the broadcasting layer.
* **Event-Sourced Replay Buffer:** Leverages **Redis Lists** and **Atomic Counters** (`INCR`) to guarantee chronological state hydration and prevent distributed race conditions.
* **Concurrent Broadcasting:** A thread-safe **Gorilla WebSocket Hub** pushes real-time updates to connected clients using Go's lightweight goroutines and buffered channels.
* **Enterprise Observability:** Fully instrumented with **OpenTelemetry** and **Jaeger**. W3C Trace Context is manually propagated across the AMQP network boundary for distributed waterfall tracing.
* **Infrastructure as Code:** Orchestrated via **Docker Compose** using optimized, multi-stage distroless/alpine Dockerfiles.

---

## Architecture Deep Dive
For a comprehensive breakdown of the system design, data flow, concurrency models, and state synchronization strategies (including how the engine handles WebSocket hydration race conditions), please see the [ARCHITECTURE.md](./ARCHITECTURE.md) file.

---

## Quick Start

### Prerequisites
* Docker and Docker Compose
* Postman (or an equivalent WebSocket/gRPC client)

### Bootstrapping the Cluster
Run the following command to spin up the entire distributed cluster (Ingest Node, Broadcast Node, RabbitMQ, Redis, and Jaeger):

```docker compose up --build -d```

### Endpoints
* **gRPC Ingest Node:** localhost:50051

* **WebSocket Broadcast Node:** ws://localhost:8081/ws?match_id={id}

* **Jaeger UI:** http://localhost:16686

## Project Scope & Known Limitations
* **Technical Debt & Scope Acknowledgment:** As a portfolio piece focusing strictly on backend distributed systems and real-time data streaming, this project explicitly bounds its scope. The following elements represent known limitations and areas for future expansion:

**No Visual Frontend Client:** Currently, the system hydrates state via raw JSON over WebSockets. A frontend UI (e.g., vanilla JS with an HTML5 canvas) is required to parse the JSON and visually render the chessboard.

**No Cold Storage / Archival:** The active match state is stored ephemerally in Redis. Once a match concludes, the event log is lost. A persistent database (like PostgreSQL) and a background worker are needed to archive finished games.

**Missing Authentication & Security:** The gRPC ingest endpoint is currently plaintext and unauthenticated. In a production environment, it requires gRPC TLS encryption and JWT/API Key authentication to prevent malicious payload injection.

**Resilience Engineering:** The architecture currently assumes infrastructure health. Implementing the Circuit Breaker Pattern and Exponential Backoff Retries for Redis and RabbitMQ connections would drastically improve fault tolerance.


---