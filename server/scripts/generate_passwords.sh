#!/bin/bash
# 生成HotGo强密码脚本
# 使用方法: bash generate_passwords.sh

echo "=========================================="
echo "  HotGo 强密码生成工具"
echo "=========================================="
echo ""

# 生成PostgreSQL密码（16字符，包含大小写字母、数字、特殊字符）
PG_PASSWORD=$(openssl rand -base64 12 | tr -d "=+/" | cut -c1-16)
PG_PASSWORD="${PG_PASSWORD}$(openssl rand -base64 1 | tr -d "=+/" | cut -c1-1 | tr '[:lower:]' '[:upper:]')"
PG_PASSWORD="${PG_PASSWORD}$(echo '!@#$%^&*' | fold -w1 | shuf | head -1)"

# 生成Redis密码（16字符）
REDIS_PASSWORD=$(openssl rand -base64 12 | tr -d "=+/" | cut -c1-16)
REDIS_PASSWORD="${REDIS_PASSWORD}$(openssl rand -base64 1 | tr -d "=+/" | cut -c1-1 | tr '[:lower:]' '[:upper:]')"
REDIS_PASSWORD="${REDIS_PASSWORD}$(echo '!@#$%^&*' | fold -w1 | shuf | head -1)"

# 生成TOKEN_SECRET_KEY（64字符十六进制）
TOKEN_SECRET_KEY=$(openssl rand -hex 32)

# 生成TCP_CRON_SECRET_KEY（32字符十六进制）
TCP_CRON_SECRET_KEY=$(openssl rand -hex 16)

# 生成TCP_AUTH_SECRET_KEY（32字符十六进制）
TCP_AUTH_SECRET_KEY=$(openssl rand -hex 16)

echo "✅ 密码生成完成！"
echo ""
echo "=========================================="
echo "  生成的密码和密钥"
echo "=========================================="
echo ""
echo "1. PostgreSQL 数据库密码:"
echo "   $PG_PASSWORD"
echo ""
echo "2. Redis 缓存密码:"
echo "   $REDIS_PASSWORD"
echo ""
echo "3. TOKEN_SECRET_KEY (64字符):"
echo "   $TOKEN_SECRET_KEY"
echo ""
echo "4. TCP_CRON_SECRET_KEY (32字符):"
echo "   $TCP_CRON_SECRET_KEY"
echo ""
echo "5. TCP_AUTH_SECRET_KEY (32字符):"
echo "   $TCP_AUTH_SECRET_KEY"
echo ""
echo "=========================================="
echo ""

# 保存到文件
PASSWORD_FILE="generated_passwords_$(date +%Y%m%d_%H%M%S).txt"
cat > "$PASSWORD_FILE" <<EOF
# HotGo 密码和密钥
# 生成时间: $(date)
# ⚠️  请妥善保管此文件，不要提交到Git仓库！

==========================================
PostgreSQL 数据库密码
==========================================
$PG_PASSWORD

==========================================
Redis 缓存密码
==========================================
$REDIS_PASSWORD

==========================================
TOKEN_SECRET_KEY (登录令牌密钥)
==========================================
$TOKEN_SECRET_KEY

==========================================
TCP_CRON_SECRET_KEY (TCP定时任务密钥)
==========================================
$TCP_CRON_SECRET_KEY

==========================================
TCP_AUTH_SECRET_KEY (TCP认证密钥)
==========================================
$TCP_AUTH_SECRET_KEY

==========================================
环境变量配置 (.env格式)
==========================================
PG_DB=hotgo
PG_USER=hotgo_user
PG_PASSWORD=$PG_PASSWORD
REDIS_PASSWORD=$REDIS_PASSWORD
TOKEN_SECRET_KEY=$TOKEN_SECRET_KEY
TCP_CRON_SECRET_KEY=$TCP_CRON_SECRET_KEY
TCP_AUTH_SECRET_KEY=$TCP_AUTH_SECRET_KEY

==========================================
Systemd服务环境变量配置
==========================================
在 /etc/systemd/system/hotgo.service 的 [Service] 部分添加：

Environment="TOKEN_SECRET_KEY=$TOKEN_SECRET_KEY"
Environment="TCP_CRON_SECRET_KEY=$TCP_CRON_SECRET_KEY"
Environment="TCP_AUTH_SECRET_KEY=$TCP_AUTH_SECRET_KEY"
Environment="REDIS_PASSWORD=$REDIS_PASSWORD"
Environment="PG_PASSWORD=$PG_PASSWORD"

EOF

chmod 600 "$PASSWORD_FILE"
echo "✅ 密码已保存到: $PASSWORD_FILE"
echo "⚠️  文件权限已设置为600（仅所有者可读写）"
echo ""
echo "=========================================="
echo "  下一步操作"
echo "=========================================="
echo ""
echo "1. 备份密码文件到安全位置"
echo "2. 使用 update_passwords.sh 脚本更新配置"
echo "   或手动更新以下文件："
echo "   - PostgreSQL: ALTER USER hotgo_user WITH PASSWORD '$PG_PASSWORD';"
echo "   - Redis: /etc/redis/redis.conf (requirepass $REDIS_PASSWORD)"
echo "   - config.yaml: 更新对应配置项"
echo "   - systemd服务: 添加环境变量"
echo ""
