#!/bin/sh
set -e

DOMAIN=${DOMAIN:?DOMAIN is required}
CERT_DIR="/etc/letsencrypt/live/$DOMAIN"

# If no Let's Encrypt cert yet, create a self-signed placeholder so nginx can start
if [ ! -f "$CERT_DIR/fullchain.pem" ]; then
    echo "No certificate found for $DOMAIN, generating self-signed placeholder..."
    mkdir -p "$CERT_DIR"
    openssl req -x509 -nodes -days 1 -newkey rsa:2048 \
        -keyout "$CERT_DIR/privkey.pem" \
        -out "$CERT_DIR/fullchain.pem" \
        -subj "/CN=$DOMAIN" 2>/dev/null
fi

# Generate nginx config from template
envsubst '${DOMAIN}' < /etc/nginx/nginx.conf.template > /etc/nginx/conf.d/default.conf

exec nginx -g 'daemon off;'
