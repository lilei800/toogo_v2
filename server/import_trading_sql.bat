@echo off
cd /d D:\go\src\hotgo_v2\server\storage\data
echo Importing trading_system.sql...
mysql -u root -proot hotgo < trading_system.sql
echo.
echo Done!
pause

