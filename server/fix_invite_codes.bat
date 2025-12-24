@echo off
chcp 65001 >nul
echo ========================================
echo    修正永久邀请码工具
echo ========================================
echo.
echo 此工具将修正不符合规范的永久邀请码
echo 符合规范的格式：4位大写字母 + 4位数字（不含4）
echo.
echo 按任意键开始执行...
pause >nul

echo.
echo 正在执行...
echo.

cd /d %~dp0
go run cmd/fix_invite_codes/main.go

echo.
echo ========================================
echo 执行完毕，按任意键退出...
pause >nul

