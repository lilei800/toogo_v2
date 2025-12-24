@echo off
chcp 65001 >nul
echo ================================================
echo 🔧 数据库修复脚本
echo ================================================
echo.
echo 正在修复 hg_sys_serve_log 表...
echo.

cd /d %~dp0server\storage\data

:: 提示用户输入密码
echo 请输入MySQL root密码:
mysql -u root -p hotgo < EXECUTE_FIX_NOW.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ 数据库修复成功！
    echo.
    echo 下一步：
    echo 1. 重启后端服务（Ctrl+C 停止，然后重新 go run main.go）
    echo 2. 刷新浏览器页面
    echo 3. 重新启动机器人
    echo.
) else (
    echo.
    echo ❌ 修复失败！
    echo.
    echo 请检查：
    echo 1. MySQL 服务是否运行
    echo 2. 用户名密码是否正确
    echo 3. hotgo 数据库是否存在
    echo.
)

pause

