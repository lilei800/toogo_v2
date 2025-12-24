@echo off
chcp 65001 >nul
echo ================================================
echo 🔧 添加 is_processed 字段到 hg_trading_signal_log 表
echo ================================================
echo.
echo 正在执行SQL迁移...
echo.

cd /d %~dp0server\storage\data

:: 提示用户输入密码
echo 请输入MySQL root密码:
mysql -u root -p hotgo < add_is_processed_to_signal_log.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ 字段添加成功！
    echo.
    echo 下一步：
    echo 1. 重启后端服务（如果正在运行）
    echo 2. 系统将自动使用 is_processed 字段防止重复下单
    echo.
) else (
    echo.
    echo ❌ 执行失败！
    echo.
    echo 请检查：
    echo 1. MySQL 服务是否运行
    echo 2. 用户名密码是否正确
    echo 3. hotgo 数据库是否存在
    echo 4. 字段是否已存在（如果已存在，可以忽略错误）
    echo.
)

pause

