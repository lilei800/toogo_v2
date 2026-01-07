# HotGo 强密码生成脚本 (PowerShell版本)
# 使用方法: .\generate_passwords.ps1

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  HotGo 强密码生成工具" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# 生成随机字符函数
function Get-RandomPassword {
    param(
        [int]$Length = 16,
        [switch]$IncludeSpecialChars
    )
    
    $chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    if ($IncludeSpecialChars) {
        $chars += "!@#$%^&*"
    }
    
    $password = ""
    for ($i = 0; $i -lt $Length; $i++) {
        $password += $chars[(Get-Random -Maximum $chars.Length)]
    }
    
    # 确保包含大小写字母、数字和特殊字符
    if ($IncludeSpecialChars) {
        $password = $password -replace '.', {param($c) if ((Get-Random) % 4 -eq 0) { (Get-Random -InputObject @('A'..'Z')) } elseif ((Get-Random) % 4 -eq 1) { (Get-Random -InputObject @('a'..'z')) } elseif ((Get-Random) % 4 -eq 2) { (Get-Random -InputObject @('0'..'9')) } else { (Get-Random -InputObject @('!','@','#','$','%','^','&','*')) } }
    }
    
    return $password
}

# 生成PostgreSQL密码（16字符，包含特殊字符）
$PG_PASSWORD = Get-RandomPassword -Length 16 -IncludeSpecialChars

# 生成Redis密码（16字符，包含特殊字符）
$REDIS_PASSWORD = Get-RandomPassword -Length 16 -IncludeSpecialChars

# 生成TOKEN_SECRET_KEY（64字符十六进制）
$bytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
$TOKEN_SECRET_KEY = [System.BitConverter]::ToString($bytes).Replace("-", "").ToLower()

# 生成TCP_CRON_SECRET_KEY（32字符十六进制）
$bytes = New-Object byte[] 16
[System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
$TCP_CRON_SECRET_KEY = [System.BitConverter]::ToString($bytes).Replace("-", "").ToLower()

# 生成TCP_AUTH_SECRET_KEY（32字符十六进制）
$bytes = New-Object byte[] 16
[System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
$TCP_AUTH_SECRET_KEY = [System.BitConverter]::ToString($bytes).Replace("-", "").ToLower()

Write-Host "✅ 密码生成完成！" -ForegroundColor Green
Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  生成的密码和密钥" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. PostgreSQL 数据库密码:" -ForegroundColor Yellow
Write-Host "   $PG_PASSWORD" -ForegroundColor White
Write-Host ""
Write-Host "2. Redis 缓存密码:" -ForegroundColor Yellow
Write-Host "   $REDIS_PASSWORD" -ForegroundColor White
Write-Host ""
Write-Host "3. TOKEN_SECRET_KEY (64字符):" -ForegroundColor Yellow
Write-Host "   $TOKEN_SECRET_KEY" -ForegroundColor White
Write-Host ""
Write-Host "4. TCP_CRON_SECRET_KEY (32字符):" -ForegroundColor Yellow
Write-Host "   $TCP_CRON_SECRET_KEY" -ForegroundColor White
Write-Host ""
Write-Host "5. TCP_AUTH_SECRET_KEY (32字符):" -ForegroundColor Yellow
Write-Host "   $TCP_AUTH_SECRET_KEY" -ForegroundColor White
Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# 保存到文件
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$PASSWORD_FILE = "generated_passwords_$timestamp.txt"

$content = @"
# HotGo 密码和密钥
# 生成时间: $(Get-Date)
# ⚠️  请妥善保管此文件，不要提交到Git仓库！

==========================================
PostgreSQL 数据库密码
==========================================
$PG_PASSWORD

==========================================
Redis 缓存密码
==========================================
$REDIS_PASSWORD

==========================================
TOKEN_SECRET_KEY (登录令牌密钥)
==========================================
$TOKEN_SECRET_KEY

==========================================
TCP_CRON_SECRET_KEY (TCP定时任务密钥)
==========================================
$TCP_CRON_SECRET_KEY

==========================================
TCP_AUTH_SECRET_KEY (TCP认证密钥)
==========================================
$TCP_AUTH_SECRET_KEY

==========================================
环境变量配置 (.env格式)
==========================================
PG_DB=hotgo
PG_USER=hotgo_user
PG_PASSWORD=$PG_PASSWORD
REDIS_PASSWORD=$REDIS_PASSWORD
TOKEN_SECRET_KEY=$TOKEN_SECRET_KEY
TCP_CRON_SECRET_KEY=$TCP_CRON_SECRET_KEY
TCP_AUTH_SECRET_KEY=$TCP_AUTH_SECRET_KEY

==========================================
Systemd服务环境变量配置
==========================================
在 /etc/systemd/system/hotgo.service 的 [Service] 部分添加：

Environment="TOKEN_SECRET_KEY=$TOKEN_SECRET_KEY"
Environment="TCP_CRON_SECRET_KEY=$TCP_CRON_SECRET_KEY"
Environment="TCP_AUTH_SECRET_KEY=$TCP_AUTH_SECRET_KEY"
Environment="REDIS_PASSWORD=$REDIS_PASSWORD"
Environment="PG_PASSWORD=$PG_PASSWORD"

==========================================
配置文件更新说明 (config.yaml)
==========================================
1. database.default.pass: "$PG_PASSWORD"
   或 database.default.link: "pgsql:hotgo_user:$PG_PASSWORD@tcp(127.0.0.1:5432)/hotgo"

2. redis.default.pass: "$REDIS_PASSWORD"

3. token.secretKey: "`${TOKEN_SECRET_KEY:$TOKEN_SECRET_KEY}"

4. tcp.client.cron.secretKey: "`${TCP_CRON_SECRET_KEY:$TCP_CRON_SECRET_KEY}"

5. tcp.client.auth.secretKey: "`${TCP_AUTH_SECRET_KEY:$TCP_AUTH_SECRET_KEY}"

"@

$content | Out-File -FilePath $PASSWORD_FILE -Encoding UTF8
(Get-Item $PASSWORD_FILE).Attributes = "Hidden, Archive"

Write-Host "✅ 密码已保存到: $PASSWORD_FILE" -ForegroundColor Green
Write-Host "⚠️  文件权限已设置为隐藏和归档" -ForegroundColor Yellow
Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  下一步操作" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. 备份密码文件到安全位置" -ForegroundColor Yellow
Write-Host "2. 使用 update_passwords.sh 脚本更新配置" -ForegroundColor Yellow
Write-Host "   或手动更新以下文件：" -ForegroundColor Yellow
Write-Host "   - PostgreSQL: ALTER USER hotgo_user WITH PASSWORD '$PG_PASSWORD';" -ForegroundColor White
Write-Host "   - Redis: /etc/redis/redis.conf (requirepass $REDIS_PASSWORD)" -ForegroundColor White
Write-Host "   - config.yaml: 更新对应配置项" -ForegroundColor White
Write-Host "   - systemd服务: 添加环境变量" -ForegroundColor White
Write-Host ""
