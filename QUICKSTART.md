# Circles.DIY - Quick Start Guide

## 🚀 Get Started in 30 Seconds

### Option 1: Quick Demo (No Setup)
```bash
go run main.go
# Open http://localhost:8080
```
Uses mock data, no database needed.

### Option 2: Full Stack
```bash
make docker-up
# Open http://localhost
```
Everything included: database, cache, nginx.

## 📋 Prerequisites

- **Go 1.24+** ([install](https://go.dev/dl/))
- **Docker & Docker Compose** ([install](https://docs.docker.com/get-docker/))
- **Make** (usually pre-installed on Linux/Mac)

## 🎯 Common Commands

| Command | What it does |
|---------|--------------|
| `make dev` | Run in development mode |
| `make build` | Build binary |
| `make docker-up` | Start all services |
| `make docker-down` | Stop all services |
| `make docker-logs` | View logs |
| `make test` | Run tests |
| `make clean` | Clean build artifacts |

## 🗄️ Database Setup

### Automatic (Recommended)
Migrations run automatically on startup:
```bash
make docker-up
# Migrations apply automatically
```

### Manual Access
```bash
# Connect to database
docker-compose exec db psql -U circles -d circles_diy

# Or from host
psql -h localhost -p 5432 -U circles -d circles_diy
# Password: circles_dev_password
```

### Reset Database
```bash
docker-compose down
docker volume rm circles-diy_db_data
docker-compose up -d
```

## 🔧 Configuration

### Quick Config (Environment Variables)
```bash
export DB_HOST=localhost
export DB_PORT=5432
export REDIS_HOST=localhost
export LOG_LEVEL=debug

go run main.go
```

### Full Config (YAML)
```bash
cp config.example.yaml config.yaml
vim config.yaml
go run main.go
```

## 🌐 Endpoints

| URL | Description |
|-----|-------------|
| http://localhost:8080 | Main app (direct) |
| http://localhost | Main app (via nginx) |
| http://localhost:8080/health | Health check |
| http://docs.localhost | Documentation |

## 🐳 Docker Services

| Service | Port | Purpose |
|---------|------|---------|
| circles-diy | 8080 | Main application |
| nginx | 80, 443 | Reverse proxy |
| db (postgres) | 5432 | Database |
| redis | 6379 | Cache |
| docmost | 3000 | Documentation |

## 🛠️ Development Modes

### 1. No Database (Quickest)
```bash
go run main.go
```
- ✅ Fastest startup
- ✅ No dependencies
- ❌ Mock data only

### 2. Hybrid (Best for Development)
```bash
docker-compose up -d db redis
go run main.go
```
- ✅ Real database
- ✅ Fast iteration
- ✅ Easy debugging

### 3. Full Stack (Production-like)
```bash
make docker-up
```
- ✅ Complete environment
- ✅ Nginx included
- ✅ All services

## 🔍 Troubleshooting

### Port 8080 Already in Use
```bash
lsof -i :8080
kill -9 <PID>
```

### Can't Connect to Database
```bash
docker-compose ps db
docker-compose logs db
docker-compose restart db
```

### Migrations Failed
```bash
docker-compose logs circles-diy
# Check logs for migration errors
```

### Build Errors
```bash
go mod tidy
go mod download
make clean
make build
```

## 📚 Documentation

- **[README.md](README.md)** - Main documentation
- **[doc/LOCAL_SETUP.md](doc/LOCAL_SETUP.md)** - Detailed local setup
- **[doc/PHASE1_COMPLETE.md](doc/PHASE1_COMPLETE.md)** - What's been built
- **[doc/MIGRATION_PLAN.md](doc/MIGRATION_PLAN.md)** - Roadmap
- **[nginx/README.md](nginx/README.md)** - Nginx configuration
- **[config.example.yaml](config.example.yaml)** - Configuration reference

## 🚢 Production Deployment

### Initial Setup
```bash
export DOMAIN=yourdomain.com
export CIRCLES_DB_PASSWORD=$(openssl rand -base64 32)
export CSRF_SECRET=$(openssl rand -base64 32)
export POSTGRES_PASSWORD=$(openssl rand -base64 32)

# Run SSL setup script
./init-letsencrypt.sh
```

### With Existing SSL
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## 🧪 Testing

```bash
# Run all tests
make test

# Health check
curl http://localhost:8080/health

# Check database connection
docker-compose exec db psql -U circles -d circles_diy -c "SELECT 1;"
```

## ✅ Verify Installation

```bash
# 1. Check Go version
go version
# Should be 1.24 or higher

# 2. Build the app
make build
# Should create bin/circles-diy

# 3. Run health check
./bin/circles-diy &
sleep 2
curl http://localhost:8080/health
# Should return: healthy

# 4. Stop
killall circles-diy
```

## 🆘 Need Help?

1. Check [doc/LOCAL_SETUP.md](doc/LOCAL_SETUP.md) for detailed setup
2. Check [nginx/README.md](nginx/README.md) for nginx issues
3. View logs: `make docker-logs`
4. Check GitHub issues

## 🎓 Next Steps

1. ✅ Get it running locally
2. Read [doc/PHASE1_COMPLETE.md](doc/PHASE1_COMPLETE.md)
3. Review [doc/ERD.md](doc/ERD.md) for database schema
4. Start building features!

---

**Ready to code?** → `make dev`

**Need full stack?** → `make docker-up`

**Going to production?** → See [README.md](README.md)
