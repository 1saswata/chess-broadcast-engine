# Chess Broadcast Engine

## Overview
The **Chess Broadcast Engine** is a high-throughput, distributed telemetry router designed to broadcast real-time chess matches to thousands of simultaneous spectators. It ingests strict protobuf telemetry, decouples processing via message brokering, synchronizes state using an event-sourced replay buffer, archives historical data into relational storage, and streams updates to authenticated clients over highly concurrent WebSockets.

---

## Key Features
* **Zero-Trust Security & Identity:** Backed by a PostgreSQL credential vault using Bcrypt password hashing. Stateless JSON Web Tokens (JWT) secure both the gRPC ingestion perimeter and WebSocket upgrades.
* **Strict Color-Locked Authorization:** High-performance Redis Hashes (`HSET`/`HGET`) dynamically lock Grandmaster identities to specific board orientations (White/Black), preventing piece-spoofing at the gRPC interceptor level.
* **Infrastructure Defense (Rate Limiting):** A Distributed Rate Limiter protects the ingest nodes using atomic Redis Lua scripts (Fixed Window algorithm), preventing noisy-neighbor DDoS attacks and client loop bugs.
* **ACID Transactional Archival:** Completed matches are drained from the ephemeral Redis cache and permanently archived into PostgreSQL using strict multi-table SQL Transactions (`BeginTx`, `Commit`, `Rollback`) to guarantee data integrity.
* **Event-Sourced Replay Buffer:** Leverages **Redis Lists** and **Atomic Counters** (`INCR`) to guarantee chronological state hydration and prevent distributed race conditions.
* **Event-Driven Architecture:** Utilizes **RabbitMQ** (AMQP Fanout Exchanges) to completely decouple the ingestion layer from the highly concurrent Go/Gorilla WebSocket broadcasting layer.
* **Clean Architecture:** Built using idiomatic Go standards, featuring strict struct-based dependency injection, separated HTTP routing (`internal/api`), and `go:embed` database migrations for single-binary deployments.
* **Enterprise Observability:** Fully instrumented with **OpenTelemetry** and **Jaeger**. W3C Trace Context is manually propagated across the AMQP network boundary for distributed waterfall tracing.

---

## Architecture Deep Dive
For a comprehensive breakdown of the system design, data flow, concurrency models, infrastructure defense, and state synchronization strategies (including how the engine handles WebSocket hydration race conditions), please see the [ARCHITECTURE.md](./ARCHITECTURE.md) file.

---

## Quick Start

### Prerequisites
* Docker and Docker Compose
* Modern Web Browser

### Bootstrapping the Cluster
Run the following command to spin up the entire distributed cluster (Ingest Node, Broadcast Node, RabbitMQ, Redis, PostgreSQL, and Jaeger):
```bash
docker compose up --build -d
```

### Endpoints

* **Visual Frontend UI:** http://localhost:8081

* **gRPC Ingest Node:** localhost:50051

* **HTTP Token Dispenser:** localhost:8080/login

* **Jaeger UI:** http://localhost:16686

## Project Scope, Future Expansion & Roadmap
As a portfolio piece focusing strictly on backend distributed systems, the current architecture provides a robust, production-ready foundation. If development resumes, the following roadmap outlines the next logical evolutions of the product:

**The Live Telemetry Dashboard (Concurrency Focus):** Building a highly concurrent Metrics Engine using sync.RWMutex. A background Go Worker would wake up on a time.Ticker, aggregate active spectator counts and latest game states from Redis, and broadcast a lightweight "Dashboard" JSON payload to a global /ws/dashboard endpoint for homepage viewers.

**Automated TTL-Based Archival (Sweeper Pattern):** Currently, the archival process (migrating data from Redis to Postgres) requires a manual HTTP trigger. A future iteration will introduce a background Go Cron job that scans Redis for inactive matches (e.g., no moves in 2 hours), assumes the game was abandoned/completed, archives the moves via the Postgres transaction pipeline, and automatically clears the cache.

**Cloud-Native Orchestration (Kubernetes):** Migrating the deployment strategy from docker-compose to strict Kubernetes manifests (Deployments, Services, ConfigMaps, Secrets) to demonstrate true cloud-native horizontal scalability and self-healing.

**Resilience Engineering:** Implementing the Circuit Breaker Pattern and Exponential Backoff Retries for Redis and RabbitMQ connections to drastically improve fault tolerance during network partitions.

---
