## 一键 Docker Compose 上线（Singapore VPS / Ubuntu 24.04）

### 你需要先做的事（必须）
- **DNS 解析**：把 `www.toogo.my` 和 `toogo.my` 的 A 记录指向 `45.76.157.103`
- **开放端口**：确保 80/443 可访问（Vultr 防火墙/安全组 + 服务器 UFW）

### 在服务器上执行（复制粘贴即可）
在 `/opt/toogo/toogo_v2` 目录执行：

```bash
sudo bash -lc "cd /opt/toogo/toogo_v2 && \
  PG_PASSWORD='Toogo2027!#$888' \
  REDIS_PASSWORD='Redis2027!@#$888' \
  EMAIL='admin@toogo.my' \
  bash server/deploy/oneclick_sg.sh"
```

### 成功标志
- `docker compose ps` 显示 `postgres / redis / hotgo / nginx` 都是 `Up`
- 浏览器访问：`https://www.toogo.my`

### 数据初始化说明
`server/storage/data/initial_data.sql` 会在 **第一次创建数据库卷（pgdata 为空）** 时自动导入。
> 你如果之前是把 SQL 导入到“宿主机 PostgreSQL 14”，那和 Docker Compose 里的 `postgres` 容器是两套数据库：容器首次启动仍会在它自己的卷里导入一次（这是正常且安全的）。
如果你重跑且想重置数据库，需要先删卷数据（会清空数据库）：

```bash
cd /opt/toogo/toogo_v2/server
docker compose -f deploy/docker-compose.prod.yml down
sudo rm -rf deploy_data/pgdata
docker compose -f deploy/docker-compose.prod.yml up -d --build postgres
```

