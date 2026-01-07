@echo off
chcp 65001 >nul
echo ==========================================
echo   GitHub 推送脚本
echo ==========================================
echo.

cd /d D:\go\src\toogo_v2

echo 当前目录: %CD%
echo.

echo 检查Git状态...
git status
echo.

echo 当前分支:
git branch
echo.

set /p confirm="确认推送到GitHub? (y/n): "
if /i not "%confirm%"=="y" (
    echo 已取消
    pause
    exit /b
)

echo.
echo ==========================================
echo 开始推送...
echo ==========================================
echo.

echo [1/4] 添加所有更改...
git add .
if errorlevel 1 (
    echo    ❌ 添加失败
    pause
    exit /b
)
echo    ✅ 添加完成

echo.
echo [2/4] 检查更改...
git status --short
echo.

set /p commit_msg="请输入提交信息 (直接回车使用默认): "
if "%commit_msg%"=="" set commit_msg=Update project files

echo.
echo [3/4] 提交更改...
git commit -m "%commit_msg%"
if errorlevel 1 (
    echo    ⚠️  没有需要提交的更改或提交失败
) else (
    echo    ✅ 提交完成
)

echo.
echo [4/4] 推送到GitHub...
git push origin main
if errorlevel 1 (
    echo.
    echo ❌ 推送失败！
    echo.
    echo 可能的原因：
    echo   1. 网络连接问题
    echo   2. GitHub认证未配置
    echo   3. 分支冲突
    echo.
    echo 请查看 GitHub推送指南.md 了解详细解决方法
) else (
    echo    ✅ 推送成功！
    echo.
    echo 🎉 代码已成功推送到GitHub！
    echo 访问: https://github.com/lilei800/toogo_v2
)

echo.
pause
