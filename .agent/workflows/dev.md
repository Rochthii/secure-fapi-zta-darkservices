# Development Workflow

1. Ensure Docker Desktop is running.
2. Verify `docker/.env` exists (copy from template if missing).
3. Start the Docker Compose infra stack (OpenZiti + Postgres + Monitoring):
   ```bash
   docker compose -f docker/docker-compose.yml up -d
   ```
4. Verify container status: `docker compose -f docker/docker-compose.yml ps`.
5. Run the Go services (`idp`, `gateway`, `client`) for local dev/testing.
