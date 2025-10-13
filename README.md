![cover image](static/img/circles-og-image.jpeg)

A social platform for communities, creators and collaborators.

Read the manifesto: https://circles.diy/

## Production Deployment (Ubuntu VM)

### Prerequisites
- Ubuntu 24.04 LTS VM with public IP
- Docker and Docker Compose installed
- Domain pointing to your VM's IP
- Ports 80 and 443 open in firewall

### Step-by-Step Setup

1. **Install Docker (if needed):**
   ```bash
   sudo apt update
   sudo apt install docker.io docker-compose-v2 -y
   sudo usermod -aG docker $USER
   # Log out and back in
   ```

2. **Configure firewall:**
   ```bash
   sudo ufw allow 22/tcp   # SSH
   sudo ufw allow 80/tcp   # HTTP
   sudo ufw allow 443/tcp  # HTTPS
   sudo ufw enable
   ```

3. **Clone and deploy:**
   ```bash
   git clone git@github.com:circlesdiy/circle-node.git
   cd circles-node
   ./init-letsencrypt.sh yourdomain.com your@email.com
   ```

### What the Setup Does
The `init-letsencrypt.sh` script:
- ✅ Validates your domain and email
- ✅ Checks domain DNS resolution  
- ✅ Starts services with HTTP-only nginx
- ✅ Requests Let's Encrypt SSL certificate
- ✅ Switches to HTTPS configuration
- ✅ Tests the final deployment
- ✅ Sets up automatic certificate renewal with certbot

### Troubleshooting

**Certificate request fails:**
```bash
# Check domain resolution
dig yourdomain.com A

# Check HTTP accessibility
curl -I http://yourdomain.com

# View nginx logs
docker compose logs nginx
```

**Service not starting:**
```bash
# Check all container status
docker compose ps

# View app logs
docker compose logs circles-diy

# Restart services
docker compose restart
```

### Manual Operations

**Manual SSL renewal:**
```bash
docker compose --profile renewal run --rm certbot
docker compose restart nginx
```

**View logs:**
```bash
docker compose logs -f
```

**Update and redeploy:**
```bash
git pull
docker compose build --no-cache
docker compose up -d
```

### Security Features
- ✅ HTTPS with Let's Encrypt
- ✅ Rate limiting (10 req/min general, 5 req/min feedback)
- ✅ Security headers (HSTS, CSP, XSS protection)
- ✅ Input validation and sanitization
- ✅ Non-root container execution
- ✅ CSRF protection

### Monitoring
- **View logs:** `docker compose logs -f`
- **Check certificates:** `docker compose exec nginx nginx -t`

### Local Development

#### Option 1: Direct Go (No Database)
```bash
# Run with mock data (no database required)
go run main.go
# Access at http://localhost:8080
```

#### Option 2: Full Stack with Docker
```bash
# Start all services (PostgreSQL + Redis + App + Nginx)
make docker-up

# Access the app
open http://localhost              # Main app
open http://docs.localhost         # Documentation (Docmost)

# View logs
make docker-logs

# Stop services
make docker-down
```

#### Option 3: Database Only (Hybrid)
```bash
# Start just database and Redis
docker-compose up -d db redis

# Run app locally
go run main.go
# Access at http://localhost:8080
```

### Available Make Commands

```bash
make build        # Build the application binary
make run          # Build and run
make dev          # Development mode with auto-reload
make test         # Run tests
make clean        # Clean build artifacts
make docker-up    # Start all Docker services
make docker-down  # Stop all Docker services
make docker-logs  # View logs
make deps         # Install dependencies
make fmt          # Format code
```

### Configuration

The app uses YAML configuration with environment variable overrides:

```bash
# Copy example config
cp config.example.yaml config.yaml

# Or use environment variables
export DB_HOST=localhost
export DB_PORT=5432
export REDIS_HOST=localhost
export LOG_LEVEL=debug

# Run
go run main.go
```

See [config.example.yaml](config.example.yaml) for all available options.

### Nginx Configuration

Three nginx configurations are available:

- **`nginx/nginx-local.conf`** - Local development (HTTP, no SSL)
- **`nginx/nginx.conf`** - Production (HTTPS with Let's Encrypt)
- **`nginx/nginx-init.conf`** - Initial setup for SSL certificate

See [nginx/README.md](nginx/README.md) for detailed nginx documentation.