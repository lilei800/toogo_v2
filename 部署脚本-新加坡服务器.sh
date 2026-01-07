#!/bin/bash
# Toogo_v2 部署脚本 - 专为新加坡Vultr服务器定制
# 服务器配置: 2 vCPUs, 4GB RAM, 128GB NVMe, Ubuntu 24.04 LTS
# 使用方法: bash 部署脚本-新加坡服务器.sh

set -e  # 遇到错误立即退出

echo "=========================================="
echo "  Toogo_v2 部署到新加坡Vultr服务器"
echo "  Ubuntu 24.04 LTS"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量（新加坡服务器优化配置）
DB_PASSWORD="Toogo2027!@#\$888"
REDIS_PASSWORD="Redis2027!@#\$888"
APP_USER=$(whoami)
APP_DIR="/opt/toogo"
PROJECT_DIR="$APP_DIR/toogo_v2"

echo -e "${YELLOW}新加坡服务器配置信息:${NC}"
echo "  服务器规格: 2 vCPUs, 4GB RAM, 128GB NVMe"
echo "  域名: www.toogo.my"
echo "  数据库密码: $DB_PASSWORD"
echo "  Redis密码: $REDIS_PASSWORD"
echo "  应用用户: $APP_USER"
echo "  项目目录: $PROJECT_DIR"
echo ""
read -p "确认开始部署? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

# 检查是否为root用户
if [ "$EUID" -eq 0 ]; then
   echo -e "${RED}错误: 请不要使用root用户运行此脚本${NC}"
   echo "请使用普通用户运行，脚本会在需要时自动使用sudo"
   exit 1
fi

# 第一步：系统初始化
echo -e "${GREEN}[1/10] 更新系统软件包...${NC}"
sudo apt update
sudo apt upgrade -y
sudo apt install -y curl wget git vim htop net-tools unzip build-essential software-properties-common

# 设置时区为新加坡时间
echo -e "${GREEN}[2/10] 设置时区为新加坡时间...${NC}"
sudo timedatectl set-timezone Asia/Singapore

# 配置防火墙
echo -e "${GREEN}[3/10] 配置防火墙...${NC}"
sudo apt install -y ufw
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable

# 第二步：安装Go 1.24.4
echo -e "${GREEN}[4/10] 安装Go语言环境...${NC}"
cd /tmp
if [ ! -f "go1.24.4.linux-amd64.tar.gz" ]; then
    wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz
fi
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz

# 配置Go环境变量
if ! grep -q "/usr/local/go/bin" /etc/profile; then
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
    echo 'export GOPATH=$HOME/go' | sudo tee -a /etc/profile
    echo 'export PATH=$PATH:$GOPATH/bin' | sudo tee -a /etc/profile
fi
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# 配置Go代理（新加坡服务器使用阿里云代理）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=on

echo "Go版本: $(go version)"

# 第三步：安装PostgreSQL 14
echo -e "${GREEN}[5/10] 安装PostgreSQL数据库...${NC}"
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
sudo apt update
sudo apt install -y postgresql-14 postgresql-contrib-14
sudo systemctl start postgresql
sudo systemctl enable postgresql

# 配置PostgreSQL数据库
echo -e "${GREEN}[6/10] 配置PostgreSQL数据库...${NC}"
sudo -u postgres psql <<EOF
CREATE DATABASE hotgo;
CREATE USER hotgo_user WITH PASSWORD '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE hotgo TO hotgo_user;
\q
EOF

# 优化PostgreSQL配置（针对4GB内存优化）
sudo tee -a /etc/postgresql/14/main/postgresql.conf > /dev/null <<EOF

# 新加坡Vultr服务器优化配置
shared_buffers = 1GB
effective_cache_size = 3GB
maintenance_work_mem = 256MB
work_mem = 16MB
max_connections = 50
random_page_cost = 1.1
effective_io_concurrency = 200

# 网络优化
listen_addresses = '127.0.0.1'
port = 5432
EOF

sudo systemctl restart postgresql

# 第四步：安装Redis
echo -e "${GREEN}[7/10] 安装Redis缓存...${NC}"
sudo apt install -y redis-server

# 配置Redis（针对4GB内存优化）
sudo sed -i "s/# requirepass.*/requirepass $REDIS_PASSWORD/" /etc/redis/redis.conf
sudo sed -i "s/^bind.*/bind 127.0.0.1/" /etc/redis/redis.conf
sudo sed -i "s/^# maxmemory.*/maxmemory 1gb/" /etc/redis/redis.conf
sudo sed -i "s/^# maxmemory-policy.*/maxmemory-policy allkeys-lru/" /etc/redis/redis.conf

sudo systemctl restart redis-server
sudo systemctl enable redis-server

# 第五步：安装Nginx
echo -e "${GREEN}[8/10] 安装Nginx网页服务器...${NC}"
sudo apt install -y nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# 第六步：部署应用代码
echo -e "${GREEN}[9/10] 部署应用代码...${NC}"
echo -e "${BLUE}请将本地代码上传到服务器:$NC"
echo "1. 在你的本地电脑上打开PowerShell或终端"
echo "2. 运行以下命令（替换为你的服务器IP）:"
echo "   scp -r D:\\go\\src\\toogo_v2 $USER@你的服务器IP:$APP_DIR/"
echo ""
echo -e "${YELLOW}等待你上传代码完成后，按回车继续...${NC}"
read

# 检查代码是否上传成功
if [ ! -d "$PROJECT_DIR" ]; then
    echo -e "${RED}错误: 代码目录不存在，请确认已正确上传${NC}"
    exit 1
fi

# 安装Go依赖
if [ -d "$PROJECT_DIR/server" ]; then
    cd $PROJECT_DIR/server
    echo "下载Go依赖..."
    go mod download
    go mod verify

    # 配置应用
    if [ ! -f "manifest/config/config.yaml" ]; then
        cp manifest/config/config.example.yaml manifest/config/config.yaml

        # 修改数据库配置
        sed -i "s/pass:.*/pass: \"$DB_PASSWORD\"/" manifest/config/config.yaml
        sed -i "s/user:.*/user: \"hotgo_user\"/" manifest/config/config.yaml

        # 修改Redis配置
        sed -i "s/pass:.*/pass: \"$REDIS_PASSWORD\"/" manifest/config/config.yaml

        # 修改服务器地址（只监听本地）
        sed -i 's/address:.*/address: "127.0.0.1:8000"/' manifest/config/config.yaml
    fi

    # 初始化数据库
    echo "初始化数据库..."
    if [ -f "storage/data/hotgo-pg.sql" ]; then
        PGPASSWORD="$DB_PASSWORD" psql -h 127.0.0.1 -U hotgo_user -d hotgo -f storage/data/hotgo-pg.sql || echo "数据库初始化完成"
    fi

    # 编译应用
    echo "编译Go应用程序..."
    go build -o main main.go
    chmod +x main
fi

# 第七步：配置Systemd服务
echo -e "${GREEN}[10/10] 配置Systemd服务...${NC}"
sudo tee /etc/systemd/system/toogo.service > /dev/null <<EOF
[Unit]
Description=Toogo Trading System
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=$APP_USER
Group=$APP_USER
WorkingDirectory=$PROJECT_DIR/server

ExecStart=$PROJECT_DIR/server/main http
ExecReload=/bin/kill -HUP \$MAINPID

KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
Restart=on-failure
RestartSec=5s

Environment="GF_GCFG_PATH=$PROJECT_DIR/server/manifest/config"
Environment="GF_GLOG_PATH=$PROJECT_DIR/server/logs"

LimitNOFILE=65536
LimitNPROC=4096

StandardOutput=journal
StandardError=journal
SyslogIdentifier=toogo

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable toogo
sudo systemctl start toogo

# 第八步：配置Nginx
echo -e "${GREEN}配置Nginx反向代理...${NC}"
sudo tee /etc/nginx/sites-available/toogo > /dev/null <<EOF
server {
    listen 80;
    server_name www.toogo.my toogo.my;

    access_log /var/log/nginx/toogo_access.log;
    error_log /var/log/nginx/toogo_error.log;

    client_max_body_size 50M;

    # 前端静态文件
    location / {
        root $PROJECT_DIR/web/dist;
        try_files \$uri \$uri/ /index.html;
        index index.html;
    }

    # API代理
    location /admin/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # WebSocket代理
    location /socket {
        proxy_pass http://127.0.0.1:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;

        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # 静态资源缓存
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2|ttf|eot)\$ {
        root $PROJECT_DIR/web/dist;
        expires 7d;
        add_header Cache-Control "public, immutable";
    }
}
EOF

sudo ln -sf /etc/nginx/sites-available/toogo /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl restart nginx

# 完成
echo ""
echo -e "${GREEN}=========================================="
echo "  🎉 部署完成！"
echo "==========================================${NC}"
echo ""
echo "📊 服务状态:"
sudo systemctl status toogo --no-pager -l
echo ""
echo "🔐 重要信息:"
echo "  域名: www.toogo.my"
echo "  数据库密码: $DB_PASSWORD"
echo "  Redis密码: $REDIS_PASSWORD"
echo "  项目目录: $PROJECT_DIR"
echo ""
echo "🚀 常用命令:"
echo "  查看服务状态: sudo systemctl status toogo"
echo "  查看日志: sudo journalctl -u toogo -f"
echo "  重启服务: sudo systemctl restart toogo"
echo ""
SERVER_IP="45.76.157.103"
echo "🌐 访问地址: http://$SERVER_IP"
echo ""
echo "📝 下一步:"
echo "1. 在浏览器中访问上面的地址"
echo "2. 进行系统初始化配置"
echo "3. 配置交易所API"
echo "4. 创建交易机器人"
echo ""
echo "💡 提示: 部署过程中如有问题，请查看日志或寻求帮助"