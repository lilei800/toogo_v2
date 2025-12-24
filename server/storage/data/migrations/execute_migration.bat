@echo off
chcp 65001 >nul
echo ====================================
echo 执行策略模板字段优化迁移
echo ====================================
echo.

mysql -h 127.0.0.1 -P 3306 -u root -proot hotgo < optimize_strategy_template_fields.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ SQL迁移执行成功！
) else (
    echo.
    echo ❌ SQL迁移执行失败，请检查错误信息
)

pause

















