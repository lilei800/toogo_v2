@echo off
chcp 65001 >nul
echo ============================================================
echo 执行波动率配置表迁移：适配新算法
echo ============================================================
echo.

REM 设置数据库连接信息（请根据实际情况修改）
set DB_HOST=127.0.0.1
set DB_PORT=3306
set DB_USER=root
set DB_PASS=your_password
set DB_NAME=hotgo

echo 数据库连接信息:
echo   主机: %DB_HOST%
echo   端口: %DB_PORT%
echo   用户: %DB_USER%
echo   数据库: %DB_NAME%
echo.

REM 检查MySQL命令是否存在
where mysql >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 未找到mysql命令，请确保MySQL已安装并添加到PATH环境变量
    pause
    exit /b 1
)

echo 正在执行迁移脚本...
echo.

REM 执行SQL脚本
mysql -h%DB_HOST% -P%DB_PORT% -u%DB_USER% -p%DB_PASS% %DB_NAME% < volatility_config_new_algorithm_fixed.sql

if %errorlevel% equ 0 (
    echo.
    echo [成功] 迁移脚本执行完成！
    echo.
    echo 请检查数据库表结构，确认以下字段已添加：
    echo   - delta_1m, delta_5m, delta_15m, delta_30m, delta_1h
    echo   - d_threshold
    echo.
) else (
    echo.
    echo [错误] 迁移脚本执行失败，请检查：
    echo   1. 数据库连接信息是否正确
    echo   2. 数据库用户是否有足够权限
    echo   3. 字段是否已经存在（如果已存在，可以忽略此错误）
    echo.
)

pause

