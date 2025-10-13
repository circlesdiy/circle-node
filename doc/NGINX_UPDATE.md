# Nginx Configuration Update - Complete

## Summary

Updated nginx configuration to support both local development and production deployment with comprehensive documentation.

## What Was Added

### 1. Nginx Configuration Files

#### **`nginx/nginx-local.conf`** ✨ NEW
Local development configuration:
- HTTP only (no SSL)
- Proxies to `circles-diy:8080` and `docmost:3000`
- WebSocket support for HTMX SSE
- Relaxed security headers for development
- 1-hour cache for static files
- Health check at `/health`
- Supports `localhost` and `*.local` domains

#### **`nginx/nginx.conf`** ✨ NEW
Production configuration:
- HTTPS with Let's Encrypt SSL
- HTTP → HTTPS redirect
- Modern SSL configuration (TLSv1.2, TLSv1.3)
- Strict security headers (HSTS, CSP, etc.)
- Rate limiting (10 req/s general, 30 req/s API)
- 1-year cache for static files
- WebSocket support
- OCSP stapling
- Gzip compression
- Separate config for main app and Docmost docs

#### **`nginx/nginx-init.conf`** (existing)
Initial setup configuration for SSL certificate acquisition

### 2. Docker Configuration Updates

#### **`docker-compose.yml`** 🔧 UPDATED
- Uses `nginx-local.conf` by default
- Updated health check to use `/health` endpoint
- Comment indicates how to switch to production config

#### **`docker-compose.prod.yml`** ✨ NEW
Production overrides:
- Switches to production `nginx.conf` with SSL
- Sets ENVIRONMENT=production
- JSON logging
- Removes development volume mounts
- Hides database port

### 3. Documentation

#### **`nginx/README.md`** ✨ NEW
Comprehensive nginx documentation:
- Configuration file explanations
- Local development instructions
- Production deployment guide
- Rate limiting configuration
- WebSocket support details
- Security headers explanation
- Troubleshooting guide
- Performance tuning tips
- Certificate renewal instructions

#### **`doc/LOCAL_SETUP.md`** ✨ NEW
Complete local development guide:
- Three ways to run locally (no DB, full stack, hybrid)
- Configuration options (YAML vs env vars)
- Database setup and reset
- Testing instructions
- Development workflow
- IDE setup (VS Code, GoLand)
- Troubleshooting common issues
- Performance tips

#### **`README.md`** 🔧 UPDATED
Enhanced with:
- Three local development options
- Make command reference
- Configuration instructions
- Links to nginx and setup documentation

## File Structure

```
nginx/
├── nginx-local.conf        # Local dev (HTTP only)
├── nginx.conf              # Production (HTTPS)
├── nginx-init.conf         # Initial setup
└── README.md               # Nginx documentation

doc/
├── LOCAL_SETUP.md          # Local development guide
├── NGINX_UPDATE.md         # This file
├── PHASE1_COMPLETE.md      # Phase 1 summary
└── MIGRATION_PLAN.md       # Overall migration plan

docker-compose.yml          # Default (local dev)
docker-compose.prod.yml     # Production overrides
```

## How to Use

### Local Development (Default)

```bash
# Start with local nginx config (HTTP only)
docker-compose up -d

# Access
open http://localhost
open http://docs.localhost
```

### Production Deployment

```bash
# Set environment variables
export DOMAIN=yourdomain.com
export CIRCLES_DB_PASSWORD=$(openssl rand -base64 32)
export CSRF_SECRET=$(openssl rand -base64 32)

# Start with production overrides
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Switching Configurations

**Local → Production:**
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

**Production → Local:**
```bash
docker-compose up -d
```

## Key Features

### For Local Development
- ✅ Simple HTTP access on port 80
- ✅ No SSL certificate needed
- ✅ Works with `localhost` out of the box
- ✅ Support for `.local` domains
- ✅ WebSocket support for HTMX
- ✅ Health check endpoint
- ✅ Subdomain support (docs.localhost)

### For Production
- ✅ Automatic HTTPS with Let's Encrypt
- ✅ Strong SSL/TLS configuration
- ✅ HTTP/2 support
- ✅ Rate limiting (DDoS protection)
- ✅ Security headers (HSTS, CSP, etc.)
- ✅ Aggressive static file caching
- ✅ Gzip compression
- ✅ WebSocket support
- ✅ OCSP stapling
- ✅ Separate subdomain for docs

## Configuration Highlights

### Rate Limiting

Production nginx includes rate limiting:
- **General requests**: 10 req/sec (burst 20)
- **API requests**: 30 req/sec (burst 20)

### Caching Strategy

- **Local**: 1 hour cache for development
- **Production**: 1 year immutable cache for static assets

### Security Headers

Production includes:
- `Strict-Transport-Security` (HSTS with preload)
- `X-Frame-Options: SAMEORIGIN`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy` (restricts camera, mic, geolocation)

### WebSocket Support

All configurations include WebSocket support for:
- HTMX Server-Sent Events (SSE)
- Future real-time features
- Docmost collaboration

## Testing

### Test Local Configuration

```bash
# Start services
docker-compose up -d

# Check nginx config
docker-compose exec nginx nginx -t

# Test endpoints
curl http://localhost/health              # Should return: healthy
curl -I http://localhost/static/css/      # Check caching headers
curl http://docs.localhost                # Test Docmost subdomain
```

### Test Production Configuration

```bash
# Start with production config
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Verify SSL (after certificate obtained)
curl -I https://yourdomain.com
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com

# Check security headers
curl -I https://yourdomain.com | grep -i "strict-transport"
```

## Troubleshooting

See [nginx/README.md](../nginx/README.md) for detailed troubleshooting:
- 502 Bad Gateway
- Connection refused
- File upload limits
- WebSocket failures
- SSL certificate issues

## Next Steps

1. ✅ Local nginx configuration complete
2. ✅ Production nginx configuration complete
3. ✅ Documentation complete
4. ⬜ Test with full stack deployment
5. ⬜ Configure domain DNS for production
6. ⬜ Run SSL certificate setup

## Related Documentation

- [nginx/README.md](../nginx/README.md) - Nginx reference
- [doc/LOCAL_SETUP.md](LOCAL_SETUP.md) - Local development guide
- [doc/PHASE1_COMPLETE.md](PHASE1_COMPLETE.md) - Phase 1 summary
- [README.md](../README.md) - Main project README
