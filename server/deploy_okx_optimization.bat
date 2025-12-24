@echo off
chcp 65001 >nul
echo ========================================
echo OKX机器人下单问题诊断和优化部署脚本
echo ========================================
echo.

:: 设置数据库连接信息
set PGHOST=127.0.0.1
set PGPORT=5432
set PGDATABASE=hotgo
set PGUSER=postgres
set PGPASSWORD=postgres

echo [步骤1] 执行诊断SQL...
echo ----------------------------------------
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -f diagnose_okx_robot.sql -o diagnose_result.txt
if %ERRORLEVEL% NEQ 0 (
    echo 错误：诊断SQL执行失败！
    echo 请检查：
    echo 1. PostgreSQL是否已启动
    echo 2. 数据库连接信息是否正确
    echo 3. 数据库用户权限是否足够
    pause
    exit /b 1
)
echo 诊断完成！结果已保存到 diagnose_result.txt
echo.

echo [步骤2] 查看诊断结果...
echo ----------------------------------------
type diagnose_result.txt
echo.

echo [步骤3] 执行数据库升级（增加失败分类字段）...
echo ----------------------------------------
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -f add_failure_category_fields.sql
if %ERRORLEVEL% NEQ 0 (
    echo 错误：数据库升级失败！
    pause
    exit /b 1
)
echo 数据库升级完成！
echo.

echo [步骤4] 验证字段已添加...
echo ----------------------------------------
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "SELECT column_name, data_type, character_maximum_length FROM information_schema.columns WHERE table_name = 'hg_trading_execution_log' AND column_name IN ('failure_category', 'failure_reason') ORDER BY ordinal_position;"
echo.

echo [步骤5] 查看OKX机器人配置...
echo ----------------------------------------
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "SELECT id, robot_name, auto_trade_enabled, auto_close_enabled, dual_side_position FROM hg_trading_robot WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL;"
echo.

echo [步骤6] 查看最近失败原因统计...
echo ----------------------------------------
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -c "SELECT message AS failure_reason, COUNT(*) AS count, MAX(created_at) AS last_occurrence FROM hg_trading_execution_log WHERE robot_id IN (SELECT id FROM hg_trading_robot WHERE platform = 'okx' AND status = 2) AND status = 'failed' AND created_at > NOW() - INTERVAL '24 hours' GROUP BY message ORDER BY count DESC LIMIT 10;"
echo.

echo ========================================
echo 部署完成！
echo ========================================
echo.
echo 下一步操作：
echo 1. 查看 diagnose_result.txt 了解详细诊断结果
echo 2. 根据失败原因统计，针对性解决问题
echo 3. 重新编译和部署代码：
echo    cd D:\go\src\hotgo_v2\server
echo    go build -o hotgo.exe main.go
echo 4. 重启服务后观察执行日志
echo.
pause

