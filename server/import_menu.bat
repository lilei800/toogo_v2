@echo off
cd /d D:\go\src\hotgo_v2\server\storage\data
echo ========================================
echo 正在导入Trading和Payment菜单配置...
echo ========================================
echo.
mysql -u root -proot hotgo < trading_payment_menu.sql
echo.
echo ========================================
echo 菜单导入完成！
echo ========================================
echo.
echo 请按以下步骤操作：
echo 1. 刷新浏览器（Ctrl+F5）
echo 2. 重新登录系统
echo 3. 查看左侧菜单栏
echo.
echo 应该能看到：
echo   💰 量化交易
echo   💵 USDT管理
echo.
pause

