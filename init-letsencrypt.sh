#!/bin/bash

# Initial Let's Encrypt certificate setup script
# Usage: ./init-letsencrypt.sh yourdomain.com your@email.com

set -e  # Exit on any error

if [ $# -ne 2 ]; then
    echo "Usage: $0 <domain> <email>"
    echo "Example: $0 circles.yourdomain.com your@email.com"
    exit 1
fi

DOMAIN=$1
EMAIL=$2

echo "### Validating domain and setup..."

# Validate domain format
if ! [[ $DOMAIN =~ ^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$ ]]; then
    echo "Error: Invalid domain format"
    exit 1
fi

# Validate email format
if ! [[ $EMAIL =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]]; then
    echo "Error: Invalid email format"
    exit 1
fi

# Check if domain resolves to this server
SERVER_IP=$(curl -s ifconfig.me || curl -s icanhazip.com || echo "unknown")
DOMAIN_IP=$(dig +short $DOMAIN | tail -n1)

if [ "$DOMAIN_IP" != "$SERVER_IP" ] && [ "$SERVER_IP" != "unknown" ]; then
    echo "Warning: Domain $DOMAIN resolves to $DOMAIN_IP but server IP is $SERVER_IP"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo "### Setting up Let's Encrypt certificate for $DOMAIN..."

# Create directories
mkdir -p ./certbot/conf/live/$DOMAIN
mkdir -p ./certbot/www

# Create working copy of nginx config and prepare SSL version
cp ./nginx/nginx.conf ./nginx/nginx.conf.bak 2>/dev/null || true
cp ./nginx/nginx-init.conf ./nginx/nginx.conf

# Replace domain in the full SSL config (using the original backed up version)
sed "s/DOMAIN/$DOMAIN/g" ./nginx/nginx.conf.bak > ./nginx/nginx-ssl.conf.tmp

echo "### Starting services with HTTP-only configuration..."

# Start services with HTTP-only nginx
docker compose up -d

# Wait for services to be ready
echo "### Waiting for services to start..."
sleep 10

# Test that the app is accessible via HTTP
if ! curl -sf "http://$DOMAIN/" > /dev/null; then
    echo "Warning: Could not reach http://$DOMAIN/ - Let's Encrypt verification may fail"
fi

echo "### Requesting Let's Encrypt certificate for $DOMAIN and docs.$DOMAIN..."

# Request certificate for both domains
docker run --rm \
    -v "$PWD/certbot/conf:/etc/letsencrypt" \
    -v "$PWD/certbot/www:/var/www/certbot" \
    certbot/certbot \
    certonly --webroot -w /var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    --force-renewal \
    -d $DOMAIN

# Switch to SSL configuration
echo "### Switching to HTTPS configuration..."
cp ./nginx/nginx-ssl.conf.tmp ./nginx/nginx.conf
rm ./nginx/nginx-ssl.conf.tmp

# Restart nginx with SSL config
docker compose restart nginx

# Wait and test HTTPS
echo "### Testing HTTPS setup..."
sleep 5
if curl -sf "https://$DOMAIN/" > /dev/null; then
    echo "✅ HTTPS is working!"
else
    echo "⚠️  HTTPS test failed, but certificate may still be valid"
fi

echo "### Setting up certificate renewal..."
(crontab -l 2>/dev/null | grep -v "certbot renew"; echo "0 12 * * * cd $PWD && docker compose run --rm certbot renew --webroot -w /var/www/certbot && docker compose restart nginx") | crontab -

echo "### Cleanup..."
rm -f ./nginx/nginx.conf.bak

echo "### Setup complete! 🎉"
echo "Your site is available at:"
echo "  circles.diy:  https://$DOMAIN"
echo "Certificate renewal is scheduled daily at noon."