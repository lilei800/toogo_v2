@echo off
chcp 65001 >nul
setlocal

REM 用法：
REM   execute_power_consume_uk.bat
REM   execute_power_consume_uk.bat 127.0.0.1 3306 root root hotgo

REM 脚本所在目录（保证从任意工作目录执行都能找到 SQL 文件）
set SCRIPT_DIR=%~dp0
set SQL_FILE=%SCRIPT_DIR%20251214_add_uk_user_order_toogo_power_consume.sql

set HOST=%1
if "%HOST%"=="" set HOST=127.0.0.1

set PORT=%2
if "%PORT%"=="" set PORT=3306

set USER=%3
if "%USER%"=="" set USER=root

set PASS=%4
if "%PASS%"=="" set PASS=root

set DB=%5
if "%DB%"=="" set DB=hotgo

echo ====================================
echo Toogo 算力扣除幂等唯一索引迁移
echo ====================================
echo HOST=%HOST% PORT=%PORT% DB=%DB% USER=%USER%
echo SQL=%SQL_FILE%
echo.

REM 1) 检查 SQL 文件是否存在
if not exist "%SQL_FILE%" (
    echo ❌ 找不到SQL文件: %SQL_FILE%
    echo    请确认脚本与SQL文件在同一目录，或从仓库重新同步该文件。
    exit /b 1
)

REM 2) 检查 mysql 是否在 PATH 中
where mysql >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 未找到mysql命令（mysql.exe不在PATH中）
    echo    解决方法：
    echo    - 把MySQL的bin目录加入系统PATH，例如：C:\Program Files\MySQL\MySQL Server 8.0\bin
    echo    - 或者把本脚本里的 mysql 改为 mysql.exe 的绝对路径
    exit /b 1
)

REM 3) 执行迁移
mysql -h %HOST% -P %PORT% -u %USER% -p%PASS% %DB% < "%SQL_FILE%"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ 迁移执行成功！
) else (
    echo.
    echo ❌ 迁移执行失败，请检查错误信息（建议先确认账号权限与表是否存在）
)

pause
endlocal


