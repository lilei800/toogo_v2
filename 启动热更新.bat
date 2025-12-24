@echo off
chcp 65001 >nul
echo ========================================
echo    HotGo 热更新启动脚本
echo ========================================
echo.

REM 检查是否安装了 Air
where air >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] Air 未安装！
    echo.
    echo 请先安装 Air：
    echo   go install github.com/cosmtrek/air@latest
    echo.
    echo 或者使用 GoFrame 内置热更新：
    echo   cd server
    echo   gf run main.go
    echo.
    pause
    exit /b 1
)

echo [信息] 使用 Air 启动热更新...
echo [提示] 修改代码后会自动重新编译运行
echo.

REM 启动 Air
air

pause

