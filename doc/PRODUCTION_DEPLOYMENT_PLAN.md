# Production Deployment Plan

## Current State: ⚠️ PARTIALLY READY

The application has a solid foundation, but several **critical issues** prevent a production deployment with `docker compose up -d`.

---

## ❌ Critical Issues Blocking Production Deployment

### 1. Dockerfile Is Outdated

**Current Issues:**
- References `index.html` (line 37) which doesn't exist anymore
- Uses outdated build process
- Missing template and static file embedding
- Health check points to wrong endpoint (line 61)
- Doesn't copy necessary files for embedded assets

**Impact**: The container won't build or will build a non-functional app.

**Location**: [Dockerfile](../Dockerfile)

### 2. Missing Production Nginx Config

**Current Issue:**
[docker-compose.yml](../docker-compose.yml#L90) references `nginx-local.conf`:
```yaml
- ./nginx/nginx-local.conf:/etc/nginx/nginx.conf:ro
```

**Impact**: Nginx configuration is for local development, not production (no SSL, etc.).

**What's Needed**: Create `nginx/nginx.conf` with:
- SSL/TLS configuration
- Proxy to circles-diy:8080
- Security headers (HSTS, CSP, X-Frame-Options, etc.)
- Rate limiting
- WebSocket support for future SSE
- Gzip compression
- Static file caching

### 3. Template/Static Volume Mounts Won't Work

**Current Issue:**
[docker-compose.yml](../docker-compose.yml#L24-25) mounts templates/static:
```yaml
- ./static:/app/static:rw
- ./templates:/app/templates:ro
```

But these directories don't exist at the root:
- Templates are in `internal/templates/html/` and embedded via `embed.FS`
- Static files are at root but aren't in the built container

**Impact**: App will fail to find templates/static files when running in container.

**Solution**:
- Remove these volume mounts for production
- Update Dockerfile to properly copy and embed all assets
- Static files should be served from embedded filesystem

### 4. Database/Redis Ports Exposed in Production

**Security Risk:**
[docker-compose.yml](../docker-compose.yml#L70-71,78-79) exposes database ports publicly:
```yaml
db:
  ports:
    - "5432:5432"  # PostgreSQL - EXPOSED TO INTERNET

redis:
  ports:
    - "6379:6379"  # Redis - EXPOSED TO INTERNET
```

**Impact**: Direct database access from internet in production = **CRITICAL SECURITY RISK**

**Solution**: Remove `ports` sections in production, only expose via Docker network.

### 5. Missing Environment File

**Current Issue:**
No `.env` file exists with production secrets. Will use insecure defaults:
- `CIRCLES_DB_PASSWORD` → defaults to `circles_secure_password`
- `CSRF_SECRET` → defaults to `change-this-in-production`
- `POSTGRES_PASSWORD` → defaults to `postgres`
- No `DOMAIN` set

**Impact**:
- Insecure default passwords
- CSRF attacks possible
- WebAuthn won't work (wrong domain/origin)

### 6. No SSL/TLS Certificates

**Current Issue:**
Certbot is configured but certificates don't exist yet. Initial deployment will fail on HTTPS.

**Solution**: Bootstrap process needed:
1. Start with HTTP-only nginx (nginx-init.conf)
2. Get certificates with certbot
3. Switch to SSL nginx.conf
4. Restart nginx

---

## ✅ What's Working Well

- ✅ Health checks configured
- ✅ Proper restart policies (`unless-stopped`)
- ✅ Network isolation (app-network)
- ✅ Volume persistence (db_data, redis_data)
- ✅ Non-root user in Dockerfile (security+)
- ✅ Production override file exists (`docker-compose.prod.yml`)
- ✅ Graceful shutdown in application code
- ✅ Database migrations run automatically
- ✅ Authentication system fully implemented

---

## 🛠️ Required Fixes for Production Deployment

### Priority 1: Update Dockerfile

**Create new Dockerfile that:**

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application (templates are embedded via embed.FS)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o circles-diy ./cmd/server

# Production stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN adduser -D -s /sbin/nologin -H appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/circles-diy .

# Copy static files (will be served from filesystem)
COPY --from=builder /app/static ./static

# Create data directory
RUN mkdir -p /app/data && \
    chown appuser:appuser /app/data && \
    chmod 750 /app/data

# Set permissions
RUN chmod 755 /app/circles-diy

# Switch to non-root user
USER appuser

# Environment variables
ENV PORT=8080
ENV TZ=UTC

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./circles-diy"]
```

### Priority 2: Create Production .env File

**Create `.env` file:**

```bash
# Generate secrets with:
# openssl rand -base64 32

# PostgreSQL superuser password
POSTGRES_PASSWORD=<generate-strong-password>

# Circles.DIY database password
CIRCLES_DB_PASSWORD=<generate-strong-password>

# CSRF protection secret (32+ bytes)
CSRF_SECRET=<generate-32-byte-secret>

# Your domain
DOMAIN=circles.diy

# Optional: Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Optional: SSL email for Let's Encrypt
LETSENCRYPT_EMAIL=admin@circles.diy
```

**Create `.env.example`:**
```bash
# Copy this to .env and fill in values

POSTGRES_PASSWORD=generate_with_openssl_rand_base64_32
CIRCLES_DB_PASSWORD=generate_with_openssl_rand_base64_32
CSRF_SECRET=generate_with_openssl_rand_base64_32
DOMAIN=yourdomain.com
LETSENCRYPT_EMAIL=admin@yourdomain.com
LOG_LEVEL=info
LOG_FORMAT=json
```

### Priority 3: Fix Docker Compose Config

**Update `docker-compose.yml` for production:**

```yaml
services:
  circles-diy:
    build:
      context: .
      dockerfile: Dockerfile
    expose:
      - "8080"
    environment:
      - PORT=8080
      - ENVIRONMENT=production
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=circles
      - DB_PASSWORD=${CIRCLES_DB_PASSWORD}
      - DB_DATABASE=circles_diy
      - DB_SSL_MODE=disable
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - CSRF_SECRET=${CSRF_SECRET}
      - WEBAUTHN_RP_ID=${DOMAIN}
      - WEBAUTHN_RP_ORIGIN=https://${DOMAIN}
      - LOG_LEVEL=${LOG_LEVEL:-info}
      - LOG_FORMAT=${LOG_FORMAT:-json}
    volumes:
      # Only mount data directory in production
      - ./data:/app/data
    restart: unless-stopped
    depends_on:
      - db
      - redis
    tmpfs:
      - /tmp:nosuid,size=100m
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - app-network

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    restart: unless-stopped
    volumes:
      - db_data:/var/lib/postgresql/data
      - ./init-db.sh:/docker-entrypoint-initdb.d/init-db.sh:ro
    networks:
      - app-network
    # REMOVED: ports section - not exposed to host in production

  redis:
    image: redis:7.2-alpine
    restart: unless-stopped
    volumes:
      - redis_data:/data
    networks:
      - app-network
    # REMOVED: ports section - not exposed to host in production

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certbot/conf:/etc/letsencrypt:ro
      - ./certbot/www:/var/www/certbot:ro
    restart: unless-stopped
    depends_on:
      - circles-diy
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  certbot:
    image: certbot/certbot
    volumes:
      - ./certbot/conf:/etc/letsencrypt
      - ./certbot/www:/var/www/certbot
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew --webroot -w /var/www/certbot --quiet; sleep 12h & wait $${!}; done;'"

networks:
  app-network:
    driver: bridge

volumes:
  db_data:
    driver: local
  redis_data:
    driver: local
```

### Priority 4: Create Production Nginx Config

**Create `nginx/nginx.conf`:**

```nginx
# Production Nginx Configuration for Circles.DIY

events {
    worker_connections 1024;
}

http {
    # Basic Settings
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 20M;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;

    # Gzip Compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript
               application/json application/javascript application/xml+rss
               application/rss+xml font/truetype font/opentype
               application/vnd.ms-fontobject image/svg+xml;

    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=general:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=auth:10m rate=5r/m;

    # Upstream to Circles.DIY app
    upstream circles_diy {
        server circles-diy:8080;
    }

    # HTTP -> HTTPS Redirect
    server {
        listen 80;
        listen [::]:80;
        server_name _;

        # Allow certbot validation
        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        # Redirect everything else to HTTPS
        location / {
            return 301 https://$host$request_uri;
        }
    }

    # HTTPS Server
    server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;
        server_name _;

        # SSL Configuration
        ssl_certificate /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;

        # SSL Settings (Mozilla Intermediate)
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
        ssl_prefer_server_ciphers off;
        ssl_session_cache shared:SSL:10m;
        ssl_session_timeout 10m;

        # OCSP Stapling
        ssl_stapling on;
        ssl_stapling_verify on;
        ssl_trusted_certificate /etc/letsencrypt/live/${DOMAIN}/chain.pem;
        resolver 8.8.8.8 8.8.4.4 valid=300s;
        resolver_timeout 5s;

        # Security Headers
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header Referrer-Policy "strict-origin-when-cross-origin" always;
        add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;

        # Health Check (no rate limiting)
        location /health {
            proxy_pass http://circles_diy;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            access_log off;
        }

        # Auth endpoints (stricter rate limiting)
        location ~ ^/auth/(login|register) {
            limit_req zone=auth burst=10 nodelay;

            proxy_pass http://circles_diy;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Static files (aggressive caching)
        location /static/ {
            proxy_pass http://circles_diy;
            proxy_http_version 1.1;
            proxy_set_header Host $host;

            # Cache for 1 year
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # All other requests
        location / {
            limit_req zone=general burst=20 nodelay;

            proxy_pass http://circles_diy;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Forwarded-Host $host;

            # WebSocket support (for future SSE)
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";

            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }
    }
}
```

### Priority 5: SSL Certificate Bootstrap

**Create `scripts/ssl-setup.sh`:**

```bash
#!/bin/bash
set -e

DOMAIN="${DOMAIN:-localhost}"
EMAIL="${LETSENCRYPT_EMAIL:-admin@${DOMAIN}}"

echo "Setting up SSL for domain: $DOMAIN"

# Step 1: Start with HTTP-only nginx for certificate challenge
echo "Starting nginx with HTTP-only configuration..."
cp nginx/nginx-init.conf nginx/nginx.conf
docker compose up -d nginx

# Step 2: Get certificates
echo "Obtaining SSL certificates..."
docker compose run --rm certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    -d $DOMAIN

# Step 3: Switch to production nginx with SSL
echo "Switching to SSL configuration..."
cp nginx/nginx-prod.conf nginx/nginx.conf

# Step 4: Restart nginx with SSL
echo "Restarting nginx with SSL..."
docker compose restart nginx

echo "SSL setup complete! Your site is now available at https://$DOMAIN"
```

---

## 📋 Step-by-Step Production Deployment

### Initial Setup

1. **Clone repository to production server**
   ```bash
   git clone <repo-url>
   cd circle-node
   ```

2. **Create production environment file**
   ```bash
   cp .env.example .env
   nano .env  # Fill in all production secrets
   ```

3. **Generate secrets**
   ```bash
   # PostgreSQL password
   openssl rand -base64 32

   # Circles DB password
   openssl rand -base64 32

   # CSRF secret
   openssl rand -base64 32
   ```

4. **Set your domain**
   ```bash
   export DOMAIN=circles.diy
   export LETSENCRYPT_EMAIL=admin@circles.diy
   ```

5. **Update init-db.sh with production password**
   - Ensure `CIRCLES_DB_PASSWORD` env var matches `.env`

### First Deployment

```bash
# 1. Build and start services
docker compose up -d db redis

# 2. Wait for database to initialize
docker compose logs -f db

# 3. Bootstrap SSL certificates
./scripts/ssl-setup.sh

# 4. Start full stack
docker compose up -d

# 5. Check logs
docker compose logs -f circles-diy

# 6. Verify health
curl https://circles.diy/health
```

### Subsequent Deployments

```bash
# Pull latest code
git pull

# Rebuild and restart
docker compose build circles-diy
docker compose up -d

# Check logs
docker compose logs -f circles-diy
```

---

## 🔒 Security Checklist

Before going to production:

- [ ] Strong, unique passwords for all services (32+ characters)
- [ ] CSRF secret is 32+ bytes and random
- [ ] Database ports NOT exposed to internet
- [ ] Redis port NOT exposed to internet
- [ ] SSL/TLS certificates configured
- [ ] HSTS header enabled
- [ ] Security headers configured (CSP, X-Frame-Options, etc.)
- [ ] Rate limiting enabled
- [ ] Non-root user in container
- [ ] Firewall configured (allow only 80, 443, SSH)
- [ ] Regular backups configured for database
- [ ] Log rotation configured
- [ ] Monitoring/alerting setup
- [ ] `.env` file NOT in git (.gitignore)

---

## 📊 Monitoring & Maintenance

### Health Checks

```bash
# Check all services
docker compose ps

# Check application health
curl https://circles.diy/health

# Check logs
docker compose logs -f circles-diy
docker compose logs -f nginx
docker compose logs -f db
```

### Database Backups

```bash
# Backup database
docker compose exec db pg_dump -U circles circles_diy > backup-$(date +%Y%m%d).sql

# Restore database
docker compose exec -T db psql -U circles circles_diy < backup.sql
```

### Certificate Renewal

Certbot renews automatically, but you can manually trigger:
```bash
docker compose run --rm certbot renew
docker compose restart nginx
```

---

## 🚀 Quick Fix (Current Development Setup)

**For now, use this hybrid approach:**

```bash
# Start only database and Redis
docker compose up -d db redis

# Run application locally
export DB_PASSWORD="circles_secure_password"
go run main.go

# Or build and run binary
go build -o bin/circles-diy .
export DB_PASSWORD="circles_secure_password"
./bin/circles-diy
```

This works perfectly for development and testing!

---

## 📝 Next Steps

When ready for production:

1. ✅ Complete all Priority 1-5 fixes above
2. ✅ Test full stack locally with `docker compose up`
3. ✅ Set up production server (VPS/cloud)
4. ✅ Configure DNS for your domain
5. ✅ Deploy using deployment script
6. ✅ Run through security checklist
7. ✅ Set up monitoring and backups

---

## 🎯 Estimated Time to Production Ready

- **Dockerfile update**: 30 minutes
- **Environment setup**: 15 minutes
- **Docker Compose fixes**: 20 minutes
- **Nginx config**: 45 minutes
- **SSL setup script**: 20 minutes
- **Testing**: 1 hour
- **Documentation**: 30 minutes

**Total**: ~3-4 hours to make fully production-ready

---

## 💡 Tips

- Test the full stack locally first with Docker
- Use `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d` for production overrides
- Keep `.env` file secure and NEVER commit to git
- Use environment-specific configs (dev, staging, prod)
- Set up automated backups from day 1
- Use a reverse proxy like Cloudflare for additional DDoS protection
- Monitor logs for suspicious activity
- Keep Go and Alpine base images updated

---

**Status**: Development setup is working. Production deployment requires the 5 priority fixes above.

**Recommendation**: Continue development with hybrid approach, implement production fixes when ready to deploy.
