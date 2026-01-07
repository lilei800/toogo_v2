@echo off
chcp 65001 >nul
echo ========================================
echo 应用钱包交易明细页面优化
echo ========================================
echo.

echo 正在应用数据库索引优化...
echo.

REM 设置数据库连接信息（根据实际情况修改）
set DB_HOST=localhost
set DB_PORT=5432
set DB_NAME=hotgo
set DB_USER=postgres

echo 请输入数据库密码:
set /p DB_PASSWORD=

echo.
echo 执行索引优化SQL...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f storage\data\migrations\optimize_order_history_indexes.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo ✅ 优化成功！
    echo ========================================
    echo.
    echo 优化内容:
    echo 1. 缓存时间从10秒延长到5分钟（已修改代码）
    echo 2. 添加数据库索引优化查询性能
    echo.
    echo 预期效果:
    echo - API调用减少 90%%+
    echo - 页面响应速度提升 80%%+
    echo - 数据库查询速度提升 50-80%%
    echo.
    echo 请重启服务使代码更改生效:
    echo   go run main.go
    echo.
) else (
    echo.
    echo ========================================
    echo ❌ 优化失败，请检查错误信息
    echo ========================================
    echo.
)

pause

