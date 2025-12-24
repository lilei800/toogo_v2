@echo off
chcp 65001 >nul
echo ====================================
echo 执行订单表字段添加迁移
echo 添加 market_state 和 risk_level 字段
echo ====================================
echo.

REM 请根据实际情况修改数据库连接信息
REM 默认: localhost:3306, 用户: root, 密码: root, 数据库: hotgo
mysql -h 127.0.0.1 -P 3306 -u root -proot hotgo < ..\add_order_market_state_fields.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ SQL迁移执行成功！
    echo 已添加字段: market_state, risk_level
) else (
    echo.
    echo ❌ SQL迁移执行失败，请检查错误信息
    echo 请确认：
    echo 1. MySQL服务是否运行
    echo 2. 数据库连接信息是否正确
    echo 3. 数据库 hotgo 是否存在
)

pause

