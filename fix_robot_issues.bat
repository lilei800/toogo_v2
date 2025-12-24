@echo off
chcp 65001 >nul
echo ================================================
echo 🔧 修复机器人连接问题
echo ================================================
echo.
echo 正在执行以下修复：
echo 1. 修复 hg_sys_serve_log 表的 line 字段
echo 2. 修复机器人 Symbol 字段
echo.

cd /d %~dp0server\storage\data

echo ▶ 步骤1：修复数据库表结构...
mysql -u root -p hotgo < EXECUTE_FIX_NOW.sql
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 步骤1失败，请检查MySQL连接
    goto :error
)
echo ✅ 步骤1完成

echo.
echo ▶ 步骤2：修复机器人Symbol字段...
mysql -u root -p hotgo < fix_robot_symbol.sql  
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 步骤2失败
    goto :error
)
echo ✅ 步骤2完成

echo.
echo ================================================
echo ✅ 所有修复完成！
echo ================================================
echo.
echo 下一步操作：
echo 1. 重启后端服务（Ctrl+C 停止，然后 go run main.go）
echo 2. 刷新浏览器页面
echo 3. 重新启动机器人
echo.
echo 注意事项：
echo - 如果还有404错误，说明需要补全后端API接口
echo - 详细说明请查看: API_FIX_SUMMARY.md
echo.
pause
exit /b 0

:error
echo.
echo ================================================
echo ❌ 修复失败！
echo ================================================
echo.
echo 请检查：
echo 1. MySQL 服务是否运行
echo 2. 用户名密码是否正确  
echo 3. hotgo 数据库是否存在
echo.
echo 或者手动执行SQL：
echo - server/storage/data/EXECUTE_FIX_NOW.sql
echo - server/storage/data/fix_robot_symbol.sql
echo.
pause
exit /b 1

