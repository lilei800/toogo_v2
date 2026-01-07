@echo off
chcp 65001 >nul
echo ================================================
echo   HotGo V2 服务器启动脚本
echo ================================================
echo.

cd /d "%~dp0"

echo [1/3] 检查并停止旧进程...
taskkill /F /IM main.exe >nul 2>&1
if %errorlevel%==0 (
    echo √ 已停止旧进程
    timeout /t 2 /nobreak >nul
) else (
    echo - 没有发现旧进程
)
echo.

echo [2/3] 检查端口占用...
netstat -ano | findstr ":8000 " | findstr "LISTENING" >nul
if %errorlevel%==0 (
    echo ! 警告: 端口8000仍被占用
    echo   请手动检查并关闭占用端口的程序
    pause
) else (
    echo √ 端口8000空闲
)
echo.

echo [3/3] 启动服务器...
echo √ 服务器正在启动...
echo.
echo ================================================
echo   修复内容：
echo   - 机器人启动时创建运行区间记录
echo   - 同步功能正确更新运行时长
echo   - 运行时长包括运行中的时间
echo ================================================
echo.
echo 按 Ctrl+C 停止服务器
echo.

main.exe http

echo.
echo 服务器已停止
pause


