#!/bin/bash
# HotGo 密码修改脚本
# 使用方法: bash change_passwords.sh

set -e

echo "=========================================="
echo "  HotGo 密码修改工具"
echo "=========================================="
echo ""
echo "⚠️  警告: 此脚本将修改数据库和Redis密码"
echo "请确保已备份重要数据！"
echo ""
read -p "确认继续? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

CONFIG_FILE="/opt/hotgo/hotgo_v2/server/manifest/config/config.yaml"
BACKUP_FILE="${CONFIG_FILE}.backup.$(date +%Y%m%d_%H%M%S)"

# 备份配置文件
if [ -f "$CONFIG_FILE" ]; then
    echo "备份配置文件..."
    cp "$CONFIG_FILE" "$BACKUP_FILE"
    echo "✅ 配置文件已备份到: $BACKUP_FILE"
fi

# 1. PostgreSQL密码
echo ""
echo "=========================================="
echo "1. 修改PostgreSQL密码"
echo "=========================================="
read -sp "请输入新的PostgreSQL密码（至少12字符）: " PG_PASSWORD
echo ""
if [ ${#PG_PASSWORD} -lt 12 ]; then
    echo "❌ 错误: 密码长度至少12个字符"
    exit 1
fi

echo "正在修改PostgreSQL密码..."
sudo -u postgres psql -c "ALTER USER hotgo_user WITH PASSWORD '$PG_PASSWORD';" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ PostgreSQL密码修改成功"
else
    echo "❌ PostgreSQL密码修改失败"
    exit 1
fi

# 2. Redis密码
echo ""
echo "=========================================="
echo "2. 修改Redis密码"
echo "=========================================="
read -sp "请输入新的Redis密码（至少12字符）: " REDIS_PASSWORD
echo ""
if [ ${#REDIS_PASSWORD} -lt 12 ]; then
    echo "❌ 错误: 密码长度至少12个字符"
    exit 1
fi

echo "正在修改Redis密码..."
sudo sed -i "s/^requirepass.*/requirepass $REDIS_PASSWORD/" /etc/redis/redis.conf
sudo systemctl restart redis-server > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Redis密码修改成功"
else
    echo "❌ Redis密码修改失败"
    exit 1
fi

# 3. 生成TOKEN密钥
echo ""
echo "=========================================="
echo "3. 生成TOKEN_SECRET_KEY"
echo "=========================================="
TOKEN_KEY=$(openssl rand -hex 32)
echo "✅ 已生成64字符TOKEN密钥"
echo "密钥: $TOKEN_KEY"
echo ""

# 4. 生成TCP密钥（可选）
echo "=========================================="
echo "4. 生成TCP服务密钥（可选）"
echo "=========================================="
read -p "是否生成TCP服务密钥? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    TCP_CRON_KEY=$(openssl rand -hex 16)
    TCP_AUTH_KEY=$(openssl rand -hex 16)
    echo "✅ TCP_CRON_SECRET_KEY: $TCP_CRON_KEY"
    echo "✅ TCP_AUTH_SECRET_KEY: $TCP_AUTH_KEY"
else
    TCP_CRON_KEY=""
    TCP_AUTH_KEY=""
fi

# 5. 更新配置文件
echo ""
echo "=========================================="
echo "5. 更新配置文件"
echo "=========================================="
if [ -f "$CONFIG_FILE" ]; then
    read -p "是否自动更新配置文件? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # 更新Redis密码
        if grep -q "pass:" "$CONFIG_FILE"; then
            # 使用sed更新Redis密码（需要根据实际格式调整）
            sed -i "s/pass:.*/pass: \"$REDIS_PASSWORD\"/" "$CONFIG_FILE"
        fi
        
        # 更新数据库密码（如果使用link格式）
        if grep -q "link:" "$CONFIG_FILE"; then
            # 提取现有配置
            OLD_LINK=$(grep "link:" "$CONFIG_FILE" | head -1 | sed 's/.*link: *"//' | sed 's/".*//')
            if [[ $OLD_LINK == pgsql:* ]]; then
                NEW_LINK=$(echo "$OLD_LINK" | sed "s/:\([^:]*\):\([^@]*\)@/:hotgo_user:$PG_PASSWORD@/")
                sed -i "s|link:.*|link: \"$NEW_LINK\"|" "$CONFIG_FILE"
            fi
        fi
        
        # 更新TOKEN密钥
        if grep -q "secretKey:" "$CONFIG_FILE"; then
            sed -i "s|secretKey:.*TOKEN_SECRET_KEY.*|secretKey: \"\${TOKEN_SECRET_KEY:$TOKEN_KEY}\"|" "$CONFIG_FILE"
        fi
        
        echo "✅ 配置文件已更新"
        echo "⚠️  请手动检查并确认以下配置："
        echo "   - Redis密码"
        echo "   - 数据库密码"
        echo "   - TOKEN_SECRET_KEY"
    else
        echo "⚠️  请手动更新配置文件: $CONFIG_FILE"
    fi
else
    echo "⚠️  配置文件不存在: $CONFIG_FILE"
    echo "请手动创建配置文件"
fi

# 6. 生成环境变量文件
echo ""
echo "=========================================="
echo "6. 生成环境变量配置"
echo "=========================================="
ENV_FILE="/opt/hotgo/hotgo_v2/server/.env"
read -p "是否生成.env文件? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    cat > "$ENV_FILE" <<EOF
# HotGo 环境变量配置
# 生成时间: $(date)

# PostgreSQL
PG_DB=hotgo
PG_USER=hotgo_user
PG_PASSWORD=$PG_PASSWORD

# Redis
REDIS_PASSWORD=$REDIS_PASSWORD

# App Secrets
TOKEN_SECRET_KEY=$TOKEN_KEY
EOF
    
    if [ -n "$TCP_CRON_KEY" ]; then
        echo "TCP_CRON_SECRET_KEY=$TCP_CRON_KEY" >> "$ENV_FILE"
        echo "TCP_AUTH_SECRET_KEY=$TCP_AUTH_KEY" >> "$ENV_FILE"
    fi
    
    chmod 600 "$ENV_FILE"
    echo "✅ .env文件已生成: $ENV_FILE"
    echo "⚠️  文件权限已设置为600（仅所有者可读写）"
fi

# 7. 验证连接
echo ""
echo "=========================================="
echo "7. 验证连接"
echo "=========================================="

echo -n "PostgreSQL连接测试: "
PGPASSWORD="$PG_PASSWORD" psql -h 127.0.0.1 -U hotgo_user -d hotgo -c "SELECT 1;" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ 成功"
else
    echo "❌ 失败"
fi

echo -n "Redis连接测试: "
redis-cli -a "$REDIS_PASSWORD" PING > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ 成功"
else
    echo "❌ 失败"
fi

# 8. 总结
echo ""
echo "=========================================="
echo "✅ 密码修改完成！"
echo "=========================================="
echo ""
echo "📋 修改摘要:"
echo "   PostgreSQL密码: ✅ 已修改"
echo "   Redis密码: ✅ 已修改"
echo "   TOKEN_SECRET_KEY: ✅ 已生成"
if [ -n "$TCP_CRON_KEY" ]; then
    echo "   TCP服务密钥: ✅ 已生成"
fi
echo ""
echo "⚠️  重要提示:"
echo "   1. 请妥善保管所有密码和密钥"
echo "   2. 建议使用密码管理器存储"
echo "   3. 更新systemd服务文件中的环境变量（如使用）"
echo "   4. 重启服务: sudo systemctl restart hotgo"
echo ""
echo "📝 密码已保存到:"
if [ -f "$ENV_FILE" ]; then
    echo "   - $ENV_FILE"
fi
echo "   - 配置文件备份: $BACKUP_FILE"
echo ""
