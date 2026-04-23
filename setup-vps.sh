#!/bin/bash
# ============================================
# Bazzar Makuku - VPS Auto Setup Script
# Run as root on Ubuntu 22.04/24.04
# Usage: curl -sSL <url> | bash
# Or:    chmod +x setup-vps.sh && ./setup-vps.sh
# ============================================

set -e

echo "🚀 Bazzar Makuku VPS Setup Starting..."

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

# ── 2. Install Nginx if not present ──
if ! command -v nginx &> /dev/null; then
    echo "📦 Installing Nginx..."
    apt-get install -y -qq nginx
    systemctl enable nginx
    echo "✅ Nginx installed"
else
    echo "✅ Nginx already installed"
fi

# ── 3. Clone / Pull repo ──
APP_DIR="/opt/bazzar"
if [ -d "$APP_DIR" ]; then
    echo "📥 Pulling latest code..."
    cd $APP_DIR
    git pull origin main
else
    echo "📥 Cloning repository..."
    git clone https://github.com/irwanprima-art/bazzar.git $APP_DIR
    cd $APP_DIR
fi

# ── 4. Create .env file ──
if [ ! -f "$APP_DIR/.env" ]; then
    JWT_SECRET=$(openssl rand -base64 32)
    cat > $APP_DIR/.env << EOF
# Bazzar Makuku Environment
PORT=8080
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=$(openssl rand -base64 16)
DB_NAME=bazzar
JWT_SECRET=${JWT_SECRET}
EOF
    echo "✅ .env created with random secrets"
else
    echo "✅ .env already exists"
fi

# Source env for docker-compose
set -a
source $APP_DIR/.env
set +a

# ── 5. Update docker-compose with env vars ──
cat > $APP_DIR/docker-compose.prod.yml << 'COMPOSE'
version: '3.8'

services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - bazzar-net

  app:
    build: .
    ports:
      - "127.0.0.1:8080:8080"
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

# ── 6. Configure Nginx ──
# Detect server IP
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')

cat > /etc/nginx/sites-available/bazzar << NGINX
server {
    listen 80;
    server_name ${SERVER_IP} _;

    client_max_body_size 50M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
}
NGINX

# Enable site
ln -sf /etc/nginx/sites-available/bazzar /etc/nginx/sites-enabled/bazzar
rm -f /etc/nginx/sites-enabled/default

# Test and restart nginx
nginx -t && systemctl restart nginx
echo "✅ Nginx configured for $SERVER_IP"

# ── 7. Build and start ──
echo "🔨 Building Docker images..."
cd $APP_DIR
docker compose -f docker-compose.prod.yml build

echo "🚀 Starting services..."
docker compose -f docker-compose.prod.yml up -d

# Wait for app to be ready
echo "⏳ Waiting for app to start..."
sleep 10

# Health check
if curl -s http://127.0.0.1:8080/api/health | grep -q '"ok"'; then
    echo ""
    echo "============================================"
    echo "✅ BAZZAR MAKUKU IS LIVE!"
    echo "============================================"
    echo ""
    echo "🌐 Access: http://${SERVER_IP}"
    echo "🔑 Default login: admin / admin123"
    echo ""
    echo "📋 Useful commands:"
    echo "   Logs:    docker compose -f /opt/bazzar/docker-compose.prod.yml logs -f"
    echo "   Restart: docker compose -f /opt/bazzar/docker-compose.prod.yml restart"
    echo "   Update:  cd /opt/bazzar && git pull && docker compose -f docker-compose.prod.yml up -d --build"
    echo ""
    echo "🔒 To add SSL with your domain:"
    echo "   1. Point your domain DNS to ${SERVER_IP}"
    echo "   2. apt install certbot python3-certbot-nginx"
    echo "   3. certbot --nginx -d yourdomain.com"
    echo "============================================"
else
    echo "⚠️ App may still be starting. Check logs:"
    echo "   docker compose -f /opt/bazzar/docker-compose.prod.yml logs -f"
fi
