@echo off
chcp 65001 >nul
echo ==========================================
echo   切换到普通用户部署 - 命令参考
echo ==========================================
echo.
echo 请在服务器上执行以下命令：
echo.
echo === 步骤1: 确保项目目录权限（root用户执行）===
echo sudo chown -R toogo:toogo /opt/toogo/toogo_v2
echo.
echo === 步骤2: 复制部署脚本（root用户执行）===
echo cp /root/部署脚本-新加坡服务器.sh /opt/toogo/toogo_v2/
echo chown toogo:toogo /opt/toogo/toogo_v2/部署脚本-新加坡服务器.sh
echo chmod +x /opt/toogo/toogo_v2/部署脚本-新加坡服务器.sh
echo.
echo === 步骤3: 确保toogo用户在sudo组（root用户执行）===
echo usermod -aG sudo toogo
echo.
echo === 步骤4: 切换到toogo用户 ===
echo su - toogo
echo 密码: Singapore2027!@#
echo.
echo === 步骤5: 进入项目目录并运行脚本 ===
echo cd /opt/toogo/toogo_v2
echo bash 部署脚本-新加坡服务器.sh
echo.
echo ==========================================
pause
