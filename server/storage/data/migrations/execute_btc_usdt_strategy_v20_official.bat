@echo off
setlocal enabledelayedexpansion

REM ============================================================
REM 执行 BTC-USDT 官方策略组合 V20（稳健回撤版）迁移脚本
REM 依赖：
REM   1) mysql.exe 在 PATH 中
REM   2) 你能连到数据库（账号/密码/库名）
REM 使用示例：
REM   execute_btc_usdt_strategy_v20_official.bat 127.0.0.1 3306 root your_password hotgo
REM ============================================================

if "%~1"=="" (
  echo 用法: %~nx0 ^<host^> ^<port^> ^<user^> ^<password^> ^<database^>
  echo 示例: %~nx0 127.0.0.1 3306 root 123456 hotgo
  exit /b 1
)

where mysql >nul 2>nul
if errorlevel 1 (
  echo 错误: 找不到 mysql.exe，请把 MySQL 客户端加入 PATH
  exit /b 1
)

set "DB_HOST=%~1"
set "DB_PORT=%~2"
set "DB_USER=%~3"
set "DB_PASS=%~4"
set "DB_NAME=%~5"

set "SQL_FILE=%~dp0btc_usdt_strategy_v20_official.sql"

if not exist "%SQL_FILE%" (
  echo 错误: 找不到 SQL 文件: "%SQL_FILE%"
  exit /b 1
)

echo 正在执行: %SQL_FILE%
mysql -h "%DB_HOST%" -P "%DB_PORT%" -u "%DB_USER%" -p"%DB_PASS%" "%DB_NAME%" < "%SQL_FILE%"
if errorlevel 1 (
  echo 执行失败，请检查连接参数/权限/SQL错误
  exit /b 1
)

echo 执行成功：已导入 BTC-USDT 官方策略组合 V20
exit /b 0











