#!/bin/bash
# ============================================
# Bazzar Makuku - VPS Auto Setup Script
# Domain: bazzar.souluze.com
# Run as root on Ubuntu 22.04/24.04
# Usage: chmod +x setup-vps.sh && ./setup-vps.sh
# ============================================

set -e

DOMAIN="bazzar.souluze.com"
APP_DIR="/opt/bazzar"
APP_PORT="8090"

echo "🚀 Bazzar Makuku VPS Setup Starting..."
echo "🌐 Domain: ${DOMAIN}"
echo ""

# ── 1. Install Docker if not present ──
if ! command -v docker &> /dev/null; then
    echo "📦 Installing Docker..."
    apt-get update -qq
    apt-get install -y -qq ca-certificates curl gnupg lsb-release
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    systemctl enable docker
    systemctl start docker
    echo "✅ Docker installed"
else
    echo "✅ Docker already installed"
fi

# ── 2. Install Nginx + Certbot ──
if ! command -v nginx &> /dev/null; then
    echo "📦 Installing Nginx..."
    apt-get install -y -qq nginx
    systemctl enable nginx
    echo "✅ Nginx installed"
else
    echo "✅ Nginx already installed"
fi

if ! command -v certbot &> /dev/null; then
    echo "📦 Installing Certbot for SSL..."
    apt-get install -y -qq certbot python3-certbot-nginx
    echo "✅ Certbot installed"
else
    echo "✅ Certbot already installed"
fi

# ── 3. Open firewall ports ──
if command -v ufw &> /dev/null; then
    ufw allow 80/tcp >/dev/null 2>&1 || true
    ufw allow 443/tcp >/dev/null 2>&1 || true
    echo "✅ Firewall ports 80/443 opened"
fi

# ── 4. Clone / Pull repo ──
if [ -d "$APP_DIR" ]; then
    echo "📥 Pulling latest code..."
    cd $APP_DIR
    git pull origin main
else
    echo "📥 Cloning repository..."
    git clone https://github.com/irwanprima-art/bazzar.git $APP_DIR
    cd $APP_DIR
fi

# ── 5. Create .env file ──
if [ ! -f "$APP_DIR/.env" ]; then
    DB_PASS=$(openssl rand -base64 16 | tr -d '=+/')
    JWT_SECRET=$(openssl rand -base64 32 | tr -d '=+/')
    cat > $APP_DIR/.env << EOF
PORT=8090
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=${DB_PASS}
DB_NAME=bazzar
JWT_SECRET=${JWT_SECRET}
EOF
    echo "✅ .env created with secure random secrets"
else
    echo "✅ .env already exists, keeping current values"
fi

# ── 6. Create production docker-compose ──
cat > $APP_DIR/docker-compose.prod.yml << 'COMPOSE'
version: '3.8'

services:
  db:
    image: postgres:16-alpine
    env_file: .env
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - bazzar-net

  app:
    build: .
    ports:
      - "127.0.0.1:8090:8090"
    env_file: .env
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - bazzar-net

volumes:
  pgdata:

networks:
  bazzar-net:
    driver: bridge
COMPOSE
echo "✅ docker-compose.prod.yml created"

# ── 7. Configure Nginx for domain ──
cat > /etc/nginx/sites-available/bazzar << NGINX
server {
    listen 80;
    server_name ${DOMAIN};

    client_max_body_size 50M;

    location / {
        proxy_pass http://127.0.0.1:8090;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }
}
NGINX

ln -sf /etc/nginx/sites-available/bazzar /etc/nginx/sites-enabled/bazzar
rm -f /etc/nginx/sites-enabled/default

nginx -t && systemctl restart nginx
echo "✅ Nginx configured for ${DOMAIN}"

# ── 8. Build and start ──
echo ""
echo "🔨 Building Docker images (this may take a few minutes)..."
cd $APP_DIR
docker compose -f docker-compose.prod.yml build --no-cache

echo "🚀 Starting services..."
docker compose -f docker-compose.prod.yml up -d

echo "⏳ Waiting for app to be ready..."
for i in {1..30}; do
    if curl -sf http://127.0.0.1:8090/api/health > /dev/null 2>&1; then
        echo "✅ App is running!"
        break
    fi
    sleep 2
done

# ── 9. Setup SSL with Let's Encrypt ──
echo ""
echo "🔒 Setting up SSL certificate for ${DOMAIN}..."
echo "   (Make sure DNS A record for ${DOMAIN} points to this server)"
echo ""

certbot --nginx -d ${DOMAIN} --non-interactive --agree-tos --email admin@souluze.com --redirect || {
    echo ""
    echo "⚠️  SSL setup failed. This usually means:"
    echo "    1. DNS for ${DOMAIN} is not pointing to this server yet"
    echo "    2. Port 80 is blocked by firewall"
    echo ""
    echo "    Fix DNS first, then run manually:"
    echo "    certbot --nginx -d ${DOMAIN}"
    echo ""
}

# ── 10. Setup auto-renewal cron ──
if ! crontab -l 2>/dev/null | grep -q certbot; then
    (crontab -l 2>/dev/null; echo "0 3 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'") | crontab -
    echo "✅ SSL auto-renewal cron configured"
fi

# ── Done ──
echo ""
echo "============================================"
echo "✅ BAZZAR MAKUKU DEPLOYMENT COMPLETE!"
echo "============================================"
echo ""
echo "🌐 URL:   https://${DOMAIN}"
echo "🔑 Login: admin / admin123"
echo ""
echo "📋 Useful commands:"
echo "   Logs:     cd /opt/bazzar && docker compose -f docker-compose.prod.yml logs -f"
echo "   Restart:  cd /opt/bazzar && docker compose -f docker-compose.prod.yml restart"
echo "   Update:   cd /opt/bazzar && git pull && docker compose -f docker-compose.prod.yml up -d --build"
echo "   SSL fix:  certbot --nginx -d ${DOMAIN}"
echo "   DB shell: docker compose -f /opt/bazzar/docker-compose.prod.yml exec db psql -U postgres bazzar"
echo "============================================"
