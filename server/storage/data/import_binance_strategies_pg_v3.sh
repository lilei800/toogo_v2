#!/bin/bash

# ============================================================
# 导入币安交易所官方策略组到PostgreSQL数据库（V3）
# ============================================================

# 设置数据库连接参数（请根据实际情况修改）
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_NAME="hotgo"

echo "============================================================"
echo "导入币安交易所官方策略组到PostgreSQL数据库（V3）"
echo "============================================================"
echo ""
echo "数据库配置:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  用户: $DB_USER"
echo "  数据库: $DB_NAME"
echo ""

# 设置PGPASSWORD环境变量（如果需要密码）
# export PGPASSWORD="your_password_here"

echo "正在导入币安BTCUSDT策略组（V3）..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f official_binance_btcusdt_v3_pg.sql
if [ $? -ne 0 ]; then
    echo "[错误] BTCUSDT策略组(V3)导入失败！"
    exit 1
fi
echo "[成功] BTCUSDT策略组(V3)导入完成！"
echo ""

echo "============================================================"
echo "导入完成！共导入1个策略组，12套策略（V3）"
echo "============================================================"


