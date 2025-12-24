# 集成测试（integration）运行说明

本项目默认 `go test ./...` **只保证单元/纯逻辑测试通过**，不会强依赖本机 DB/Redis/外部文件。
凡是需要外部依赖（数据库、Redis、本地文件、第三方服务）的测试，都使用 `integration` build tag 隔离。

---

## 1. 默认（推荐）快速健康检查

在 `hotgo_v2/server` 目录执行：

```bash
go test ./...
```

该命令应当可以在“无 DB/Redis/本地附件文件”的环境下直接跑通，用于 CI/日常架构健康度体检。

---

## 2. 运行集成测试（需要 DB/Redis 等依赖）

在 `hotgo_v2/server` 目录执行：

```bash
go test -tags=integration ./...
```

---

## 3. 集成测试依赖清单（最小集）

### 3.1 配置文件

GoFrame 默认会从 `manifest/config` 下加载配置（例如 `manifest/config/config.yaml`）。

- **MySQL**：需要在配置中提供可用的数据库连接（host/port/user/pass/db）。
- **Redis**：需要在配置中提供可用的 redis 配置（host/port/password/db）。

如果你不想改本地配置文件，建议复制 `manifest/config/config.example.yaml` 到 `manifest/config/config.yaml` 后再按需调整（注意：仓库里可能已存在 `config.yaml`，以实际为准）。

### 3.2 数据库

集成测试涉及到 DAO/权限/Hook 等能力时，通常需要如下（至少）表结构存在：

- `hg_sys_addons_install`（插件安装记录）
- `hg_trading_*`（交易机器人/订单等相关表）
- 以及各模块自身依赖的表

> 建议在独立的测试库运行，避免污染生产数据。

### 3.3 Redis

涉及分布式锁等组件的测试需要 Redis 可用。

---

## 4. 常用排错

- **报错：configuration not found**  
  说明测试运行时没找到配置文件。确认当前工作目录为 `hotgo_v2/server`，并检查 `manifest/config/config.yaml` 是否存在。

- **报错：no configuration found for creating redis client**  
  说明 Redis 未配置或不可连接。补齐 `manifest/config/config.yaml` 的 redis 配置，或确保 Redis 服务已启动。


