@echo off
chcp 65001 >nul
echo ============================================================
echo 导入币安交易所官方策略组到PostgreSQL数据库
echo ============================================================
echo.

REM 设置数据库连接参数（请根据实际情况修改）
set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=postgres
set DB_NAME=hotgo

echo 数据库配置:
echo   主机: %DB_HOST%
echo   端口: %DB_PORT%
echo   用户: %DB_USER%
echo   数据库: %DB_NAME%
echo.

REM 设置PGPASSWORD环境变量（如果需要密码）
REM set PGPASSWORD=your_password_here

echo 正在导入币安BTCUSDT策略组...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f official_binance_btcusdt_v1_pg.sql
if %errorlevel% neq 0 (
    echo [错误] BTCUSDT策略组导入失败！
    pause
    exit /b 1
)
echo [成功] BTCUSDT策略组导入完成！
echo.

echo 正在导入币安ETHUSDT策略组...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f official_binance_ethusdt_v1_pg.sql
if %errorlevel% neq 0 (
    echo [错误] ETHUSDT策略组导入失败！
    pause
    exit /b 1
)
echo [成功] ETHUSDT策略组导入完成！
echo.

echo ============================================================
echo 导入完成！共导入2个策略组，24套策略
echo ============================================================
pause

