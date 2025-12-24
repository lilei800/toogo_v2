@echo off
echo ========================================
echo 更新用户dong的上级关系
echo ========================================
echo.

cd /d D:\go\src\hotgo_v2\server
go run cmd/update_inviter/main.go

echo.
pause

