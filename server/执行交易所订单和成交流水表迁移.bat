@echo off
chcp 65001 >nul
echo ================================================
echo   执行交易所订单和成交流水表迁移（PostgreSQL）
echo ================================================
echo.

cd /d "%~dp0"

echo [1/2] 创建交易所订单事实表 (hg_trading_exchange_order)...
echo.

type storage\data\migrations\create_trading_exchange_order_pg.sql

echo.
echo 正在执行...
psql -h localhost -p 5432 -U postgres -d hotgo -f storage\data\migrations\create_trading_exchange_order_pg.sql

if %errorlevel% neq 0 (
    echo.
    echo ❌ 创建 hg_trading_exchange_order 表失败！
    echo.
    pause
    exit /b 1
)

echo.
echo ✅ hg_trading_exchange_order 表创建成功！
echo.
echo ================================================
echo.

echo [2/2] 创建成交流水表 (hg_trading_trade_fill)...
echo.

type storage\data\migrations\create_trading_trade_fill_pg.sql

echo.
echo 正在执行...
psql -h localhost -p 5432 -U postgres -d hotgo -f storage\data\migrations\create_trading_trade_fill_pg.sql

if %errorlevel% neq 0 (
    echo.
    echo ❌ 创建 hg_trading_trade_fill 表失败！
    echo.
    pause
    exit /b 1
)

echo.
echo ✅ hg_trading_trade_fill 表创建成功！
echo.
echo ================================================
echo.
echo 🎉 所有迁移执行成功！
echo.
echo 创建的表：
echo   1. hg_trading_exchange_order - 交易所订单事实表
echo   2. hg_trading_trade_fill - 成交流水表
echo.
echo 用途：
echo   - hg_trading_exchange_order: WebSocket实时推送订单，供前端挂单列表展示
echo   - hg_trading_trade_fill: 存储交易所成交记录，精确盈亏和手续费数据
echo.
echo ================================================

pause

