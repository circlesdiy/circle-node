# Nginx Configuration

This directory contains nginx configurations for different environments.

## Configuration Files

### `nginx-local.conf`
**Local development configuration** - HTTP only, no SSL
- Used by default in `docker-compose.yml`
- Proxies to `circles-diy:8080`
- HTTP on port 80
- No SSL/TLS
- Relaxed security headers
- WebSocket support for HTMX SSE
- Health check at `/health`

### `nginx.conf`
**Production configuration** - HTTPS with Let's Encrypt
- SSL/TLS with modern ciphers
- HTTP → HTTPS redirect
- HSTS headers
- Rate limiting
- Aggressive caching for static files
- WebSocket support
- OCSP stapling
- Gzip compression

### `nginx-init.conf`
**Initial setup configuration** - Minimal HTTP for Let's Encrypt
- Used during initial SSL certificate setup
- Supports ACME challenge path
- Routes all traffic to app until SSL is configured

## Local Development

The default `docker-compose.yml` uses `nginx-local.conf`:

```bash
# Start services (HTTP only on localhost:80)
docker-compose up -d

# Access the app
open http://localhost
```

### Testing Locally

```bash
# Test nginx config
docker-compose exec nginx nginx -t

# Reload nginx
docker-compose exec nginx nginx -s reload

# View nginx logs
docker-compose logs -f nginx
```

### Local Hosts File

For subdomain support locally, add to `/etc/hosts`:
```
127.0.0.1 localhost circles.local
127.0.0.1 docs.localhost docs.circles.local
```

## Production Deployment

### Initial Setup (HTTP Only)

1. **First deployment** - Use `nginx-init.conf` for Let's Encrypt:

```bash
# Temporarily use init config
cp nginx/nginx-init.conf nginx/nginx.conf

# Set your domain
export DOMAIN=yourdomain.com
export CIRCLES_DB_PASSWORD=$(openssl rand -base64 32)
export CSRF_SECRET=$(openssl rand -base64 32)
export POSTGRES_PASSWORD=$(openssl rand -base64 32)

# Start services
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

2. **Obtain SSL certificates**:

```bash
# Run Let's Encrypt
./init-letsencrypt.sh
```

3. **Switch to production config with SSL**:

```bash
# Copy production config (already done if you cloned correctly)
# nginx/nginx.conf is the SSL-enabled version

# Restart nginx with SSL config
docker-compose -f docker-compose.yml -f docker-compose.prod.yml restart nginx
```

### Production with SSL

Once SSL certificates are obtained:

```bash
# Set environment variables
export DOMAIN=yourdomain.com
export CIRCLES_DB_PASSWORD=your_secure_password
export CSRF_SECRET=$(openssl rand -base64 32)
export POSTGRES_PASSWORD=$(openssl rand -base64 32)

# Start with production overrides
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Certificate Renewal

Certificates auto-renew via `docker-compose`:

```bash
# Manual renewal
docker-compose run --rm certbot renew

# Reload nginx after renewal
docker-compose exec nginx nginx -s reload
```

## Configuration Details

### Rate Limiting

**Production nginx.conf includes:**
- General requests: 10 req/sec with burst of 20
- API requests: 30 req/sec with burst of 20

Adjust in `nginx.conf`:
```nginx
limit_req_zone $binary_remote_addr zone=general:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=api:10m rate=30r/s;
```

### Static File Caching

- **Local**: 1 hour cache
- **Production**: 1 year cache with immutable headers

### WebSocket Support

All configs include WebSocket support for:
- HTMX SSE (Server-Sent Events)
- Future WebSocket features

### Security Headers

Production config includes:
- `Strict-Transport-Security` (HSTS)
- `X-Frame-Options`
- `X-Content-Type-Options`
- `X-XSS-Protection`
- `Referrer-Policy`
- `Permissions-Policy`

### File Upload Limits

Default: `client_max_body_size 50M`

Increase in config if needed:
```nginx
client_max_body_size 100M;
```

## Troubleshooting

### Check Configuration

```bash
# Test config syntax
docker-compose exec nginx nginx -t

# View running config
docker-compose exec nginx cat /etc/nginx/nginx.conf
```

### View Logs

```bash
# All logs
docker-compose logs nginx

# Follow logs
docker-compose logs -f nginx

# Access logs only
docker-compose exec nginx tail -f /var/log/nginx/access.log

# Error logs only
docker-compose exec nginx tail -f /var/log/nginx/error.log
```

### Common Issues

**"502 Bad Gateway"**
- App not running: `docker-compose ps circles-diy`
- Check app logs: `docker-compose logs circles-diy`
- Check app health: `curl http://circles-diy:8080/health`

**"Connection refused"**
- Check upstream services are running
- Verify network connectivity: `docker-compose exec nginx ping circles-diy`

**"413 Request Entity Too Large"**
- Increase `client_max_body_size` in nginx config

**WebSocket connection fails**
- Ensure `proxy_http_version 1.1` is set
- Check `Upgrade` and `Connection` headers are proxied

## Switching Configurations

### Development → Production

```bash
# Use production compose file
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Production → Development

```bash
# Use default compose file
docker-compose up -d
```

## Performance Tuning

### Worker Processes

Adjust based on CPU cores:
```nginx
events {
    worker_connections 2048;  # Connections per worker
}
```

### Caching

Enable proxy caching for better performance:
```nginx
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m;
```

### Gzip Compression

Already enabled for:
- HTML, CSS, JS
- JSON, XML
- SVG images

## Environment Variables

The production `nginx.conf` uses:
- `${DOMAIN}` - Your domain name (e.g., `example.com`)

Set in `docker-compose.prod.yml` or environment.
