#!/bin/bash
# 修复配置文件YAML错误脚本

set -e

CONFIG_FILE="/opt/toogo/toogo_v2/server/manifest/config/config.yaml"
CONFIG_EXAMPLE="/opt/toogo/toogo_v2/server/manifest/config/config.example.yaml"
BACKUP_FILE="${CONFIG_FILE}.backup.$(date +%Y%m%d_%H%M%S)"

echo "=========================================="
echo "  修复配置文件YAML错误"
echo "=========================================="
echo ""

# 检查文件是否存在
if [ ! -f "$CONFIG_EXAMPLE" ]; then
    echo "❌ 错误: 配置文件模板不存在: $CONFIG_EXAMPLE"
    exit 1
fi

# 停止服务
echo "[1/6] 停止服务..."
sudo systemctl stop toogo 2>/dev/null || true
echo "✅ 服务已停止"

# 备份当前配置
if [ -f "$CONFIG_FILE" ]; then
    echo "[2/6] 备份当前配置..."
    cp "$CONFIG_FILE" "$BACKUP_FILE"
    echo "✅ 配置已备份到: $BACKUP_FILE"
else
    echo "[2/6] 配置文件不存在，跳过备份"
fi

# 从模板复制
echo "[3/6] 从模板创建新配置文件..."
cp "$CONFIG_EXAMPLE" "$CONFIG_FILE"
echo "✅ 配置文件已创建"

# 修改关键配置
echo "[4/6] 修改关键配置..."

# 修改数据库配置
if grep -q "pgsql:postgres:postgres@tcp" "$CONFIG_FILE"; then
    sed -i 's|pgsql:postgres:postgres@tcp(127.0.0.1:5432)/hotgo|pgsql:hotgo_user:Toogo2027!@#$888@tcp(127.0.0.1:5432)/hotgo|g' "$CONFIG_FILE"
    echo "  ✅ 数据库配置已更新"
fi

# 修改Redis密码
sed -i 's|pass: "${REDIS_PASSWORD:}"|pass: "Redis2027!@#$888"|g' "$CONFIG_FILE"
sed -i 's|pass: ""|pass: "Redis2027!@#$888"|g' "$CONFIG_FILE"
echo "  ✅ Redis配置已更新"

# 修改服务器地址
sed -i 's|address: ":8000"|address: "127.0.0.1:8000"|g' "$CONFIG_FILE"
echo "  ✅ 服务器地址已更新"

# 修复YAML常见问题
echo "[5/6] 修复YAML格式问题..."

# 将tab替换为空格
sed -i 's/\t/  /g' "$CONFIG_FILE"

# 确保冒号后有空格（但避免破坏URL）
sed -i 's|:\([^ /"]\)|: \1|g' "$CONFIG_FILE"

# 修复可能的引号问题
sed -i "s|'|\"|g" "$CONFIG_FILE" 2>/dev/null || true

echo "  ✅ YAML格式已修复"

# 验证YAML语法
echo "[6/6] 验证YAML语法..."
if command -v python3 &> /dev/null; then
    if python3 -c "import yaml; yaml.safe_load(open('$CONFIG_FILE'))" 2>/dev/null; then
        echo "  ✅ YAML语法验证通过"
    else
        echo "  ⚠️  YAML语法验证失败，请手动检查"
        echo "  错误信息："
        python3 -c "import yaml; yaml.safe_load(open('$CONFIG_FILE'))" 2>&1 | head -5
    fi
else
    echo "  ⚠️  Python3未安装，跳过YAML验证"
fi

echo ""
echo "=========================================="
echo "✅ 配置文件修复完成！"
echo "=========================================="
echo ""
echo "📋 下一步："
echo "  1. 检查配置文件: vim $CONFIG_FILE"
echo "  2. 确认数据库和Redis密码正确"
echo "  3. 启动服务: sudo systemctl start toogo"
echo "  4. 查看日志: sudo journalctl -u toogo -f"
echo ""
echo "📝 备份文件: $BACKUP_FILE"
echo ""
