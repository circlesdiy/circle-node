# Local Development Setup Guide

This guide covers different ways to run Circles.DIY locally.

## Quick Start (No Database)

The simplest way to get started - uses mock data:

```bash
# Just run it
go run main.go

# Or use make
make dev

# Access at http://localhost:8080
```

This mode:
- ✅ No database required
- ✅ No Docker required
- ✅ Uses mock data for UI testing
- ✅ Hot-reload CSS
- ✅ Hot-reload templates (if configured)
- ❌ No data persistence
- ❌ No authentication

## Full Local Stack (Recommended)

Run everything with Docker Compose:

```bash
# Start all services
make docker-up

# Or manually
docker-compose up -d

# Access the application
open http://localhost              # Main app (via nginx)
open http://docs.localhost         # Documentation

# View logs
make docker-logs

# Stop everything
make docker-down
```

This gives you:
- ✅ PostgreSQL database
- ✅ Redis cache
- ✅ Nginx reverse proxy
- ✅ Docmost documentation
- ✅ Auto-migrations on startup
- ✅ Full production-like environment
- ✅ Health checks

### Services & Ports

| Service | Port | Access |
|---------|------|--------|
| Nginx | 80 | http://localhost |
| App (direct) | 8080 | http://localhost:8080 |
| PostgreSQL | 5432 | localhost:5432 |
| Redis | 6379 | localhost:6379 |
| Docmost | 3000 | http://docs.localhost |

## Hybrid Mode (Database + Local App)

Run database in Docker, app locally for faster iteration:

```bash
# 1. Start just database and Redis
docker-compose up -d db redis

# 2. Run app locally (connects to Docker database)
go run main.go

# Access at http://localhost:8080 (direct, no nginx)
```

This mode:
- ✅ Database persistence
- ✅ Fast rebuild (no Docker)
- ✅ Easy debugging
- ✅ Live reload with `air` or similar
- ❌ No nginx (different from production)

### Using with Air (Auto-reload)

Install [air](https://github.com/cosmtrek/air):

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with auto-reload
air
```

## Configuration

### Using YAML Config

```bash
# Copy example
cp config.example.yaml config.yaml

# Edit as needed
vim config.yaml

# Run
go run main.go
```

### Using Environment Variables

```bash
# Set variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=circles
export DB_PASSWORD=circles_dev_password
export REDIS_HOST=localhost
export LOG_LEVEL=debug

# Run
go run main.go
```

### Docker Compose Environment

For Docker Compose, create `.env`:

```bash
# Copy example
cp .env.circles.example .env.circles

# Edit
vim .env.circles

# Start with env file
docker-compose --env-file .env.circles up
```

## Database Setup

### First Run (Automatic)

Migrations run automatically on startup:

```bash
go run main.go
# Logs will show:
# "running database migrations..."
# "successfully applied migration: 001_initial_schema.sql"
# ...
```

### Manual Database Access

```bash
# Connect to database
docker-compose exec db psql -U circles -d circles_diy

# Or from host (if psql installed)
psql -h localhost -p 5432 -U circles -d circles_diy

# Password: circles_dev_password (default)
```

### Resetting Database

```bash
# Stop services
docker-compose down

# Remove database volume
docker volume rm circles-diy_db_data

# Start fresh
docker-compose up -d

# Migrations will run again automatically
```

## Testing

### Run Tests

```bash
# All tests
make test

# Or directly
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/storage/...
```

### Manual Testing

```bash
# Health check
curl http://localhost:8080/health

# Should return: "healthy"
```

## Development Workflow

### Typical Flow

1. **Start services**:
   ```bash
   docker-compose up -d db redis
   ```

2. **Run app locally**:
   ```bash
   go run main.go
   ```

3. **Make changes** to code

4. **Restart** (Ctrl+C and run again, or use `air`)

5. **Test** at http://localhost:8080

6. **Check logs** in terminal

### CSS Development

CSS is built automatically on startup:

```bash
# Development mode (watches for changes)
go run main.go
# CSS watcher runs in background

# Manual rebuild
make clean
make build
```

### Template Development

Templates are loaded on startup. To reload:

```bash
# Restart the app
# OR
# Enable hot-reload in config.yaml:
templates:
  hot_reload: true
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using port 8080
lsof -i :8080

# Kill it
kill -9 <PID>

# Or use different port
export PORT=8081
go run main.go
```

### Database Connection Failed

```bash
# Check if database is running
docker-compose ps db

# Check logs
docker-compose logs db

# Verify credentials
docker-compose exec db psql -U circles -d circles_diy

# Reset database
docker-compose restart db
```

### Cannot Connect to Redis

```bash
# Check Redis is running
docker-compose ps redis

# Test connection
docker-compose exec redis redis-cli ping
# Should return: PONG

# Restart Redis
docker-compose restart redis
```

### Migrations Failed

```bash
# Check migration files exist
ls internal/storage/migrations/

# Check database logs
docker-compose logs db

# Try manual migration
docker-compose exec db psql -U circles -d circles_diy -f /path/to/migration.sql
```

### Build Errors

```bash
# Update dependencies
go mod tidy
go mod download

# Clean and rebuild
make clean
make build

# Check Go version (requires 1.24)
go version
```

## IDE Setup

### VS Code

Install extensions:
- Go (official)
- Docker
- YAML

Add to `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch App",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "env": {
        "DB_HOST": "localhost",
        "REDIS_HOST": "localhost",
        "LOG_LEVEL": "debug"
      }
    }
  ]
}
```

### GoLand

1. Run Configuration → Go Build
2. Set environment variables
3. Set working directory
4. Enable "Run with debugger"

## Performance Tips

### Faster Builds

```bash
# Use build cache
go build -o bin/circles-diy .

# Parallel compilation
go build -p 8 -o bin/circles-diy .
```

### Faster Docker Builds

```bash
# Use BuildKit
export DOCKER_BUILDKIT=1
docker-compose build

# Cache dependencies
# Already done in Dockerfile with multi-stage build
```

## Next Steps

Once you have the local environment running:

1. Check out [doc/PHASE1_COMPLETE.md](PHASE1_COMPLETE.md) for what's been built
2. Review [doc/MIGRATION_PLAN.md](MIGRATION_PLAN.md) for what's next
3. See [doc/ERD.md](ERD.md) for database schema
4. Read [nginx/README.md](../nginx/README.md) for nginx setup

Happy coding! 🚀
