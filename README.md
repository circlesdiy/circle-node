![cover image](static/img/circles-og-image.jpeg)

# Circles.DIY

A self-hosted social platform built around your circles, not your profile.

**Read the manifesto:** https://circles.diy/

## Deploy an Instance

### Requirements

- Ubuntu 24.04 (or similar Linux with Docker)
- Domain with DNS pointing to your server
- Ports 80 and 443 open

### Quick Start

```bash
# Install Docker
sudo apt update && sudo apt install docker.io docker-compose-v2 -y
sudo usermod -aG docker $USER && newgrp docker

# Clone and deploy
git clone https://github.com/circlesdiy/circle-node.git
cd circle-node
./init-letsencrypt.sh yourdomain.com you@email.com
```

Done. Your circles.diy instance is live at `https://yourdomain.com`

## Maintenance

**View logs:**
```bash
docker compose logs -f
```

**Restart services:**
```bash
docker compose restart
```

**Update:**
```bash
git pull && docker compose build --no-cache && docker compose up -d
```

**Manual SSL renewal** (auto-renews daily):
```bash
docker compose run --rm certbot renew && docker compose restart nginx
```

## Local Development

```bash
# Full stack (recommended)
make docker-up
open http://localhost
```

See [config.example.yaml](config.example.yaml) for configuration options.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Certificate fails | Check DNS: `dig yourdomain.com` should show your server IP |
| 502 Bad Gateway | App not ready: `docker compose logs circles-diy` |
| Can't reach site | Firewall: `sudo ufw allow 80/tcp && sudo ufw allow 443/tcp` |

## Architecture

- **Go 1.24** backend with server-side rendering
- **PostgreSQL 16** database
- **Redis 7.2** for sessions
- **Nginx** reverse proxy with Let's Encrypt SSL
- **WebAuthn** passwordless authentication

## License

[License details]
