@echo off
chcp 65001 >nul
echo ====================================
echo 执行订单事件表创建脚本
echo 创建 hg_trading_order_event 表
echo ====================================
echo.

REM 请根据实际情况修改数据库连接信息
REM 默认: localhost:3306, 用户: root, 密码: root, 数据库: hotgo
mysql -h 127.0.0.1 -P 3306 -u root -proot hotgo < create_order_event_table.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo SQL script executed successfully!
    echo Table created: hg_trading_order_event
    echo.
) else (
    echo.
    echo SQL script execution failed, please check error messages
    echo Please confirm:
    echo 1. MySQL service is running
    echo 2. Database connection info is correct
    echo 3. Database 'hotgo' exists
    echo 4. Table may already exist (ignore this error if exists)
)

pause

