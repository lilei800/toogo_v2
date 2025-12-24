@echo off
cd /d D:\go\src\hotgo_v2\server\storage\data
echo ========================================
echo 正在导入Trading和Payment菜单配置...
echo (HotGo v2.0 版本)
echo ========================================
echo.
mysql -u root -proot hotgo < trading_payment_menu_v2.sql
echo.
echo ========================================
echo 菜单导入完成！
echo ========================================
echo.
echo 接下来请：
echo 1. 刷新浏览器 (Ctrl+F5)
echo 2. 重新登录
echo 3. 查看左侧菜单
echo.
pause

