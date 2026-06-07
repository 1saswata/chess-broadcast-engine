# Runbook: Simulating a VIP Match

1. **Boot the cluster:** `docker compose up --build -d` (check .env.example).
2. **Register the Admin:** Send `POST /register` with role `"admin"`.
3. **Register the Players:** Send `POST /register` for two grandmasters (e.g., Magnus and Hikaru).
4. **Provision the Match:** Using the Admin JWT, send `POST /admin/match` to create `match_id: 100`.
5. **Connect the Spectator:** Open `http://localhost:8081` in your browser.
6. **Stream the Moves:** Use your gRPC client (or Postman gRPC) to stream moves for `match_id: 100` using the grandmaster JWTs. Observe the browser UI updating in real-time.