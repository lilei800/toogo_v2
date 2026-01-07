# 修复配置文件YAML错误

## 🔍 问题

配置文件第123行YAML语法错误：
```
yaml: line 123: mapping values are not allowed in this context
```

## ✅ 解决方案

### 方法1：检查并修复配置文件（推荐）

在服务器上执行：

```bash
# 1. 查看第123行附近的内容
sed -n '120,130p' /opt/toogo/toogo_v2/server/manifest/config/config.yaml

# 2. 检查YAML语法
cd /opt/toogo/toogo_v2/server
python3 -c "import yaml; yaml.safe_load(open('manifest/config/config.yaml'))" 2>&1 | head -20

# 或使用yamllint（如果安装了）
yamllint manifest/config/config.yaml
```

### 方法2：重新从模板复制配置文件

```bash
# 1. 备份当前配置（如果有重要修改）
cp /opt/toogo/toogo_v2/server/manifest/config/config.yaml /opt/toogo/toogo_v2/server/manifest/config/config.yaml.backup

# 2. 从模板重新复制
cd /opt/toogo/toogo_v2/server
cp manifest/config/config.example.yaml manifest/config/config.yaml

# 3. 修改关键配置
vim manifest/config/config.yaml
```

**需要修改的关键配置：**
- 数据库密码：`Toogo2027!@#$888`
- Redis密码：`Redis2027!@#$888`
- 服务器地址：`127.0.0.1:8000`

### 方法3：使用sed修复常见问题

```bash
# 修复常见的YAML问题
cd /opt/toogo/toogo_v2/server/manifest/config

# 1. 将tab替换为空格
sed -i 's/\t/  /g' config.yaml

# 2. 确保冒号后有空格
sed -i 's/:\([^ ]\)/: \1/g' config.yaml

# 3. 检查第123行
sed -n '120,130p' config.yaml
```

---

## 🚀 快速修复步骤

### 步骤1：停止服务

```bash
sudo systemctl stop toogo
```

### 步骤2：备份并重新创建配置文件

```bash
cd /opt/toogo/toogo_v2/server

# 备份
cp manifest/config/config.yaml manifest/config/config.yaml.backup

# 从模板复制
cp manifest/config/config.example.yaml manifest/config/config.yaml
```

### 步骤3：修改配置文件

```bash
vim manifest/config/config.yaml
```

**关键修改：**

```yaml
# 数据库配置（找到database.default部分）
database:
  default:
    link: "pgsql:hotgo_user:Toogo2027!@#$888@tcp(127.0.0.1:5432)/hotgo"
    # 或使用分离格式：
    host: "127.0.0.1"
    port: "5432"
    user: "hotgo_user"
    pass: "Toogo2027!@#$888"
    name: "hotgo"
    type: "pgsql"

# Redis配置（找到redis.default部分）
redis:
  default:
    address: "127.0.0.1:6379"
    db: "2"
    pass: "Redis2027!@#$888"

# 服务器配置
server:
  address: "127.0.0.1:8000"
```

### 步骤4：验证YAML语法

```bash
# 使用Python验证（如果安装了Python3）
python3 -c "import yaml; yaml.safe_load(open('manifest/config/config.yaml'))" && echo "✅ YAML语法正确" || echo "❌ YAML语法错误"

# 或手动检查
cat manifest/config/config.yaml | grep -n ":" | head -20
```

### 步骤5：启动服务

```bash
sudo systemctl start toogo
sudo systemctl status toogo
```

---

## 🔧 常见YAML错误修复

### 错误1：缩进问题

```bash
# 将所有tab替换为2个空格
sed -i 's/\t/  /g' manifest/config/config.yaml
```

### 错误2：冒号后缺少空格

```bash
# 修复冒号后缺少空格的情况
sed -i 's/:\([^ ]\)/: \1/g' manifest/config/config.yaml
```

### 错误3：特殊字符问题

```bash
# 检查特殊字符
cat manifest/config/config.yaml | grep -n "[^[:print:]]"
```

### 错误4：引号问题

```bash
# 检查引号匹配
grep -n '"' manifest/config/config.yaml | head -20
```

---

## 📋 一键修复脚本

在服务器上创建并运行：

```bash
cat > /tmp/fix_config.sh << 'EOF'
#!/bin/bash
cd /opt/toogo/toogo_v2/server

# 停止服务
sudo systemctl stop toogo

# 备份配置
cp manifest/config/config.yaml manifest/config/config.yaml.backup.$(date +%Y%m%d_%H%M%S)

# 从模板复制
cp manifest/config/config.example.yaml manifest/config/config.yaml

# 修改数据库密码
sed -i 's|pgsql:postgres:postgres@tcp(127.0.0.1:5432)/hotgo|pgsql:hotgo_user:Toogo2027!@#$888@tcp(127.0.0.1:5432)/hotgo|g' manifest/config/config.yaml

# 修改Redis密码（如果配置中有）
sed -i 's|pass: "${REDIS_PASSWORD:}"|pass: "Redis2027!@#$888"|g' manifest/config/config.yaml

# 修改服务器地址
sed -i 's|address: ":8000"|address: "127.0.0.1:8000"|g' manifest/config/config.yaml

# 验证YAML（如果安装了Python）
if command -v python3 &> /dev/null; then
    python3 -c "import yaml; yaml.safe_load(open('manifest/config/config.yaml'))" && echo "✅ YAML语法正确" || echo "⚠️  YAML验证失败，请手动检查"
fi

echo "✅ 配置文件已修复"
echo "请手动检查并修改其他配置项"
EOF

chmod +x /tmp/fix_config.sh
bash /tmp/fix_config.sh
```

---

## 🎯 推荐操作流程

```bash
# === 1. 停止服务 ===
sudo systemctl stop toogo

# === 2. 备份并重新创建配置 ===
cd /opt/toogo/toogo_v2/server
cp manifest/config/config.yaml manifest/config/config.yaml.backup
cp manifest/config/config.example.yaml manifest/config/config.yaml

# === 3. 修改关键配置 ===
vim manifest/config/config.yaml
# 修改数据库密码、Redis密码、服务器地址

# === 4. 验证配置 ===
# 手动检查或使用Python验证

# === 5. 启动服务 ===
sudo systemctl start toogo
sudo systemctl status toogo

# === 6. 查看日志确认 ===
sudo journalctl -u toogo -f
```

---

**先执行备份和重新创建配置文件，然后修改关键配置项！** 🔧
