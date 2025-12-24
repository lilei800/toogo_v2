@echo off
chcp 65001 >nul
echo ========================================
echo    HotGo 热更新启动脚本 (GoFrame)
echo ========================================
echo.

REM 检查是否安装了 GoFrame CLI
where gf >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] GoFrame CLI 未安装！
    echo.
    echo 请先安装 GoFrame CLI：
    echo   go install github.com/gogf/gf/cmd/gf/v2@latest
    echo.
    pause
    exit /b 1
)

echo [信息] 使用 GoFrame 启动热更新...
echo [提示] 修改代码后会自动重新编译运行
echo.

REM 进入 server 目录并启动
cd server
gf run main.go

pause

