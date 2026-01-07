@echo off
chcp 65001 >nul
echo ==========================================
echo   HotGo_v2 项目清理工具
echo ==========================================
echo.
echo 此脚本将清理以下内容：
echo   1. web_backup_old\node_modules (备份目录的依赖)
echo   2. web\node_modules (当前web目录的依赖，可重新安装)
echo   3. server\logs\* (日志文件)
echo   4. server\temp\* (临时文件)
echo   5. server\storage\cache\* (缓存文件)
echo   6. *.exe (编译生成的可执行文件)
echo   7. *.log (日志文件)
echo.
echo ⚠️  警告：这些文件可以安全删除，但删除后需要重新安装依赖
echo.
pause

set /p confirm="确认清理? (y/n): "
if /i not "%confirm%"=="y" (
    echo 已取消
    pause
    exit /b
)

echo.
echo ==========================================
echo 开始清理...
echo ==========================================
echo.

REM 1. 清理备份目录的node_modules
if exist "web\web_backup_old\node_modules" (
    echo [1/7] 清理 web_backup_old\node_modules...
    rd /s /q "web\web_backup_old\node_modules" 2>nul
    if errorlevel 1 (
        echo    ⚠️  删除失败，可能正在使用中
    ) else (
        echo    ✅ 已删除
    )
) else (
    echo [1/7] web_backup_old\node_modules 不存在，跳过
)

REM 2. 清理当前web目录的node_modules
if exist "web\node_modules" (
    echo [2/7] 清理 web\node_modules...
    rd /s /q "web\node_modules" 2>nul
    if errorlevel 1 (
        echo    ⚠️  删除失败，可能正在使用中
    ) else (
        echo    ✅ 已删除
        echo    💡 提示：需要时运行 npm install 重新安装
    )
) else (
    echo [2/7] web\node_modules 不存在，跳过
)

REM 3. 清理日志文件
if exist "server\logs" (
    echo [3/7] 清理 server\logs\*...
    del /f /q /s "server\logs\*.*" 2>nul
    echo    ✅ 已清理日志文件
) else (
    echo [3/7] server\logs 不存在，跳过
)

REM 4. 清理临时文件
if exist "server\temp" (
    echo [4/7] 清理 server\temp\*...
    del /f /q /s "server\temp\*.*" 2>nul
    echo    ✅ 已清理临时文件
) else (
    echo [4/7] server\temp 不存在，跳过
)

REM 5. 清理缓存文件
if exist "server\storage\cache" (
    echo [5/7] 清理 server\storage\cache\*...
    del /f /q /s "server\storage\cache\*.*" 2>nul
    echo    ✅ 已清理缓存文件
) else (
    echo [5/7] server\storage\cache 不存在，跳过
)

REM 6. 清理可执行文件
echo [6/7] 清理 *.exe 文件...
for /r %%f in (*.exe) do (
    echo    删除: %%f
    del /f /q "%%f" 2>nul
)
echo    ✅ 已清理可执行文件

REM 7. 清理日志文件
echo [7/7] 清理 *.log 文件...
for /r %%f in (*.log) do (
    echo    删除: %%f
    del /f /q "%%f" 2>nul
)
echo    ✅ 已清理日志文件

echo.
echo ==========================================
echo ✅ 清理完成！
echo ==========================================
echo.
echo 📋 清理摘要：
echo   - node_modules 已删除（需要时运行 npm install）
echo   - 日志文件已清理
echo   - 临时文件已清理
echo   - 缓存文件已清理
echo   - 可执行文件已清理
echo.
echo 💡 提示：
echo   - 如果需要前端开发，请运行: cd web ^&^& npm install
echo   - 日志文件会在运行时自动重新生成
echo.
pause
