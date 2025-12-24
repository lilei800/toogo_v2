## HotGo_v2 停机迁移：MySQL → PostgreSQL（保留用户/菜单/权限/业务数据）

这份指南按“停机切库”设计：停服务 → 全量迁移 → 校验 → 切配置启动。

### 0) 前提与说明

- **PostgreSQL 没有通用默认密码**：你说的“默认用户名密码都一样”以你当前环境为准（你能在 Navicat 连上即可）。
- 本项目后端使用 GoFrame，PG 驱动已引入；但历史 SQL/脚本存在 MySQL 方言。
- 我们已在仓库补了一个 PG 兼容函数脚本：`server/storage/data/pgsql_compat_hotgo.sql`，用于兼容少量 `IF/IFNULL/YEAR` 表达式。

### 1) 迁移前准备（建议）

- **冻结写入**：停掉 `hotgo_v2/server` 服务（以及任何会写数据库的定时任务/后台脚本）。
- **确认 PG 可连接**（任选一种）：
  - **Navicat**：已连通即可
  - **命令行**（Windows PowerShell）：

```bash
Test-NetConnection 127.0.0.1 -Port 5432
```

### 2) 在 PostgreSQL 创建目标数据库（如果还没有）

用 Navicat 或 psql 创建一个空库，例如：`hotgo`（UTF-8，schema 使用默认 `public`）。

### 3) 全量迁移数据（Navicat 推荐做法）

在 Navicat 的“数据传输/数据同步”里：

- **源**：MySQL（现网库）
- **目标**：PostgreSQL（新库）
- **对象**：至少包含所有 `hg_` 前缀表（菜单/权限/用户/业务表都在里面）
- **选项建议**：
  - 勾选“**创建表**（目标不存在时）”
  - 勾选“**传输数据**”
  - 自增主键：让 Navicat 映射为 PG 的 `serial/bigserial`（或 `identity`）
  - 迁移完成后，抽样对比关键表行数

> 如果你们是线上大库/大表，建议先迁一份测试库做演练，确认字段类型/索引/唯一键都正确。

### 4) 迁移后：在 PG 执行兼容脚本（必须）

在 PostgreSQL 的 `hotgo` 库执行：

- `server/storage/data/pgsql_compat_hotgo.sql`

它会创建 3 个兼容函数：`IF`、`IFNULL`、`YEAR`，用来兜底项目里少量历史表达式，确保切库后不报错。

### 5) 最小校验（强烈建议）

在 MySQL 和 PostgreSQL 分别执行这些 SQL，比对结果一致：

- 菜单（是否迁移完整）：

```sql
SELECT COUNT(*) FROM hg_admin_menu;
```

- 角色/权限（Casbin）：

```sql
SELECT COUNT(*) FROM hg_admin_role_casbin;
```

- 用户（按你们实际表名抽查）：

```sql
SELECT COUNT(*) FROM hg_admin_member;
```

> 如果你不确定用户表名，我可以再帮你从 `server/internal/dao` 把核心表名清单整理出来。

### 6) 切换服务端配置到 PostgreSQL

编辑 `server/manifest/config/config.yaml`：

- `database.default.link` 改为：
  - `pgsql:用户名:密码@tcp(127.0.0.1:5432)/hotgo`

仓库里我也把 `database.default.link` 改成了使用环境变量的形式，默认走 PG：

- `link: "${DATABASE_URL:pgsql:postgres:postgres@tcp(127.0.0.1:5432)/hotgo}"`

你可以直接设置环境变量 `DATABASE_URL` 来切换不同环境，而不用改 YAML。

### 7) 启动服务验证

- 启动后端，确认：
  - 能正常登录后台
  - 菜单树、角色权限、用户列表可见
  - Toogo/Trading 页面能正常加载列表

如果启动时报 SQL 错误，把报错贴我，我会按“缺表/缺索引/函数不兼容/字段类型映射”快速定位。


