#!/bin/sh
set -e

# Load DOMAIN from .env
if [ -f .env ]; then
  DOMAIN=$(grep -E '^DOMAIN=' .env | cut -d= -f2- | tr -d ' "'"'"'')
fi

if [ -z "$DOMAIN" ]; then
  echo "Error: DOMAIN is not set. Add DOMAIN=your-domain.com to .env"
  exit 1
fi

CERT_PATH="./certbot/conf/live/$DOMAIN"

if [ -d "$CERT_PATH" ]; then
  echo "Certificate already exists for $DOMAIN, skipping."
  echo "To force renewal: docker compose run --rm certbot renew --force-renewal"
  exit 0
fi

echo "==> Requesting certificate for $DOMAIN ..."

# Create dirs
mkdir -p certbot/conf certbot/www

# Start nginx in HTTP-only mode for ACME challenge
# Use a temporary config that only listens on port 80
docker compose -f docker-compose.local-self.yml up -d nginx

# Wait for nginx to be ready
sleep 3

# Request certificate
docker compose -f docker-compose.local-self.yml run --rm certbot certonly \
  --webroot \
  -w /var/www/certbot \
  -d "$DOMAIN" \
  --email "${CERTBOT_EMAIL:-admin@$DOMAIN}" \
  --agree-tos \
  --no-eff-email

# Reload nginx to pick up the new certificate
docker compose -f docker-compose.local-self.yml exec nginx nginx -s reload

echo "==> Done! Certificate issued for $DOMAIN"
echo "    Run: docker compose -f docker-compose.local-self.yml up -d"
