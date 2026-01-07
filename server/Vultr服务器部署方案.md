# Vultr 服务器部署方案

## 推荐配置（生产环境）⭐

### 服务器规格
```yaml
供应商: Vultr.com
套餐类型: High Frequency Compute
位置: Tokyo, Japan (nrt)

硬件配置:
  CPU: 4 vCPU @ 3.0+ GHz (Intel High Frequency)
  RAM: 8 GB DDR4
  Storage: 128 GB NVMe SSD
  Bandwidth: 4 TB/月 (1 Gbps 上行)
  IPv4: 1个独立IP

价格:
  基础费用: $48/月 (~340元人民币)
  自动备份: $4.80/月
  总计: $52.80/月 (~375元人民币)
```

### 性能指标
- **支持机器人数**: 50-100个
- **并发用户**: 10-30人
- **API请求**: 100+ req/s
- **到交易所延迟**: 20-60ms
- **数据库容量**: 500GB+ 历史数据

### 备用方案（预算有限）
```yaml
位置: Seoul, South Korea (icn)
配置: 2 vCPU / 4 GB / 80 GB NVMe SSD
价格: $24/月 (~170元人民币)
适用: 初期测试、个人使用、10个机器人以内
```

---

## 部署架构

### All-in-One 单服务器架构
```
┌─────────────────────────────────────┐
│  Vultr Tokyo - 4核8G128G NVMe       │
├─────────────────────────────────────┤
│  Nginx (80/443)                     │
│    ├─ 前端静态资源                   │
│    └─ 反向代理 → Go 后端 (8000)     │
├─────────────────────────────────────┤
│  Go 应用服务 (4个实例)               │
│    ├─ HTTP API 服务                 │
│    ├─ WebSocket 服务                │
│    ├─ 策略引擎                       │
│    └─ 定时任务                       │
├─────────────────────────────────────┤
│  Redis 6.x (内存缓存)               │
│    ├─ 持仓缓存                       │
│    ├─ K线数据缓存                    │
│    └─ Session管理                   │
├─────────────────────────────────────┤
│  PostgreSQL 14 (数据库)             │
│    ├─ 用户数据                       │
│    ├─ 交易记录                       │
│    ├─ 策略配置                       │
│    └─ 历史K线                        │
└─────────────────────────────────────┘
         │
         ├───► Binance (20-40ms)
         ├───► Bitget (30-50ms)
         ├───► OKX (30-50ms)
         └───► Gate.io (30-50ms)
```

---

## 系统优化配置

### 1. PostgreSQL 优化
```ini
# /etc/postgresql/14/main/postgresql.conf

# 内存配置 (8GB服务器)
shared_buffers = 2GB
effective_cache_size = 6GB
maintenance_work_mem = 512MB
work_mem = 32MB

# 连接配置
max_connections = 100
shared_preload_libraries = 'pg_stat_statements'

# 性能优化
random_page_cost = 1.1  # NVMe SSD
effective_io_concurrency = 200
checkpoint_completion_target = 0.9
wal_buffers = 16MB

# 日志配置
log_min_duration_statement = 1000  # 记录慢查询 (>1s)
```

### 2. Redis 配置
```ini
# /etc/redis/redis.conf

# 内存配置
maxmemory 2gb
maxmemory-policy allkeys-lru

# 持久化
save 900 1
save 300 10
save 60 10000

# 网络优化
tcp-keepalive 60
timeout 300
```

### 3. Go 应用配置
```yaml
# config.yaml

server:
  address: "0.0.0.0:8000"
  workers: 4  # 对应4核CPU
  
database:
  maxOpenConns: 50
  maxIdleConns: 10
  connMaxLifetime: "1h"
  
cache:
  redis:
    addr: "127.0.0.1:6379"
    poolSize: 20
```

---

## 安全配置

### 1. 防火墙规则
```bash
# UFW 配置
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw allow 8000/tcp  # Go API (可选：仅内网)
ufw deny from any to any port 5432  # 禁止外部访问 PostgreSQL
ufw deny from any to any port 6379  # 禁止外部访问 Redis
ufw enable
```

### 2. SSH 安全
```bash
# /etc/ssh/sshd_config
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
Port 2222  # 修改默认端口
```

### 3. 自动更新
```bash
# 安装 unattended-upgrades
apt install -y unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades
```

---

## 监控与运维

### 1. 基础监控
```bash
# 安装监控工具
apt install -y htop iotop nethogs

# 进程监控
htop

# 磁盘IO监控
iotop

# 网络监控
nethogs
```

### 2. 日志管理
```bash
# 日志轮转配置
# /etc/logrotate.d/hotgo

/var/log/hotgo/*.log {
    daily
    rotate 7
    compress
    delaycompress
    notifempty
    create 0644 hotgo hotgo
    sharedscripts
    postrotate
        systemctl reload hotgo
    endscript
}
```

### 3. 备份策略
```bash
# PostgreSQL 自动备份脚本
#!/bin/bash
# /root/backup_postgres.sh

BACKUP_DIR="/backup/postgres"
DATE=$(date +%Y%m%d_%H%M%S)

# 备份数据库
pg_dump -U postgres hotgo | gzip > "$BACKUP_DIR/hotgo_$DATE.sql.gz"

# 保留最近7天的备份
find $BACKUP_DIR -name "hotgo_*.sql.gz" -mtime +7 -delete

# 上传到对象存储（可选）
# s3cmd put $BACKUP_DIR/hotgo_$DATE.sql.gz s3://my-backups/
```

---

## 性能测试基准

### 预期性能指标（4核8G配置）
```
并发处理能力:
  - HTTP API: 500-1000 req/s
  - WebSocket: 1000+ 连接
  - 策略计算: 50-100 机器人同时运行

响应时间:
  - API平均响应: <50ms
  - 数据库查询: <10ms
  - 到交易所延迟: 20-60ms

资源使用:
  - CPU平均: 30-50%
  - 内存使用: 60-70% (5-6GB)
  - 磁盘IO: <100 MB/s
  - 网络带宽: <100 Mbps
```

---

## 成本预算

### 月度成本明细（单服务器方案）
```
服务器租用: $48.00/月
自动备份: $4.80/月
域名: $1.00/月 (可选)
SSL证书: $0 (Let's Encrypt免费)
对象存储: $2.00/月 (备份用)
────────────────────
总计: ~$56/月 (约400元人民币)
```

### 年度成本
```
基础设施: $672/年
预留升级: $200/年
意外支出: $100/年
────────────────────
总计: ~$972/年 (约7000元人民币)
```

---

## 扩展路径

### 当需要扩容时：

**阶段1: 垂直扩展**
- 升级到 6核16G: $96/月
- 升级到 8核32G: $192/月

**阶段2: 水平扩展**
- 添加第二台应用服务器: +$48/月
- 独立数据库服务器: +$96/月
- 负载均衡: +$10/月

**阶段3: 多地域部署**
- Tokyo + Seoul 双活: +$96/月
- 全球CDN: +$20/月

---

## 购买链接

Vultr 官网: https://www.vultr.com/
推荐使用优惠码: VULTR100 (新用户送$100体验金)

选择步骤:
1. 选择 "Cloud Compute"
2. 选择 "High Frequency"
3. 选择 "Tokyo, Japan"
4. 选择 "4 vCPU / 8 GB / 128 GB NVMe SSD"
5. 选择 "Ubuntu 22.04 LTS"
6. 启用自动备份

---

## 联系支持

如遇问题:
- Vultr 客服: https://my.vultr.com/support/
- 响应时间: 通常 1-4 小时
- 支持方式: Ticket 系统

---

**最后更新**: 2025-12-24
**文档版本**: v1.0




