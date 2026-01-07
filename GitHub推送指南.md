# GitHub 推送指南

## 📋 当前状态

- ✅ Git仓库已初始化
- ✅ 远程仓库已配置：`https://github.com/lilei800/toogo_v2.git`
- ✅ 当前分支：`main`
- ✅ `.gitignore` 已更新
- ⚠️ 需要配置GitHub认证才能推送

---

## 🔐 配置GitHub认证

### 方法1：使用Personal Access Token（推荐）

#### 1. 生成Personal Access Token

1. 访问：https://github.com/settings/tokens
2. 点击 "Generate new token" → "Generate new token (classic)"
3. 设置：
   - **Note**: `toogo_v2_push`
   - **Expiration**: 选择过期时间（建议90天或更长）
   - **Scopes**: 勾选 `repo`（完整仓库权限）
4. 点击 "Generate token"
5. **重要**：复制生成的token（只显示一次！）

#### 2. 配置Git使用Token

```bash
# 方法1：使用Git Credential Manager（推荐）
git config --global credential.helper manager-core

# 方法2：在URL中嵌入token（临时）
git remote set-url origin https://你的token@github.com/lilei800/toogo_v2.git
```

**或者使用SSH方式（更安全）：**

### 方法2：使用SSH密钥（推荐，更安全）

#### 1. 生成SSH密钥

```bash
# 检查是否已有SSH密钥
ls ~/.ssh/id_rsa.pub

# 如果没有，生成新的SSH密钥
ssh-keygen -t ed25519 -C "your_email@example.com"
# 按回车使用默认路径，设置密码（可选）
```

#### 2. 添加SSH密钥到GitHub

```bash
# 复制公钥内容
cat ~/.ssh/id_ed25519.pub
# 或 Windows PowerShell
Get-Content ~/.ssh/id_ed25519.pub
```

1. 访问：https://github.com/settings/keys
2. 点击 "New SSH key"
3. **Title**: `toogo_v2_development`
4. **Key**: 粘贴刚才复制的公钥内容
5. 点击 "Add SSH key"

#### 3. 更改远程仓库URL为SSH

```bash
git remote set-url origin git@github.com:lilei800/toogo_v2.git
```

#### 4. 测试SSH连接

```bash
ssh -T git@github.com
# 应该看到: Hi lilei800! You've successfully authenticated...
```

---

## 🚀 推送代码到GitHub

### 步骤1：检查当前状态

```bash
cd D:\go\src\toogo_v2
git status
git branch
```

### 步骤2：确保在main分支

```bash
git checkout main
```

### 步骤3：添加所有更改

```bash
# 添加所有文件（.gitignore会排除不需要的文件）
git add .

# 检查将要提交的文件
git status
```

### 步骤4：提交更改

```bash
git commit -m "Initial commit: Toogo v2 trading system"
```

### 步骤5：推送到GitHub

```bash
# 首次推送
git push -u origin main

# 后续推送
git push origin main
```

---

## 🔄 如果遇到问题

### 问题1：认证失败

**错误信息：**
```
fatal: unable to access 'https://github.com/...': Failed to connect
```

**解决方法：**
1. 检查网络连接
2. 配置GitHub认证（见上方）
3. 如果使用HTTPS，确保token正确
4. 如果使用SSH，确保SSH密钥已添加到GitHub

### 问题2：分支已存在

**错误信息：**
```
fatal: a branch named 'main' already exists
```

**解决方法：**
```bash
# 拉取远程main分支
git pull origin main --allow-unrelated-histories

# 或强制推送（谨慎使用）
git push -u origin main --force
```

### 问题3：大文件推送失败

**错误信息：**
```
remote: error: File ... is 100.00 MB; this exceeds GitHub's file size limit
```

**解决方法：**
1. 使用Git LFS处理大文件
2. 或从仓库中删除大文件

```bash
# 安装Git LFS
git lfs install

# 跟踪大文件
git lfs track "*.sql"
git lfs track "*.zip"

# 重新添加和提交
git add .gitattributes
git add .
git commit -m "Add Git LFS tracking"
```

---

## 📝 推送前检查清单

- [ ] `.gitignore` 已更新，排除了敏感文件
- [ ] 没有包含密码、密钥等敏感信息
- [ ] 没有包含 `node_modules`、`logs` 等大文件
- [ ] GitHub认证已配置（Token或SSH）
- [ ] 网络连接正常
- [ ] 在正确的分支（main）

---

## 🎯 快速推送命令（已配置认证后）

```bash
cd D:\go\src\toogo_v2
git checkout main
git add .
git commit -m "Update project files"
git push origin main
```

---

## 📚 常用Git命令

```bash
# 查看状态
git status

# 查看远程仓库
git remote -v

# 查看分支
git branch -a

# 添加文件
git add .
git add 文件名

# 提交
git commit -m "提交信息"

# 推送
git push origin main

# 拉取
git pull origin main

# 查看提交历史
git log --oneline

# 撤销未提交的更改
git checkout -- 文件名
git reset HEAD 文件名
```

---

## 🔒 安全提示

1. **不要提交敏感信息**
   - 密码、密钥、API密钥
   - 配置文件中的真实密码
   - 个人隐私信息

2. **使用 `.gitignore`**
   - 确保敏感文件已被排除
   - 定期检查 `.gitignore` 配置

3. **使用环境变量**
   - 敏感配置使用环境变量
   - 提供 `.env.example` 作为模板

4. **定期更新依赖**
   - 保持依赖包最新
   - 修复安全漏洞

---

## 📞 需要帮助？

如果遇到问题：
1. 检查GitHub状态：https://www.githubstatus.com/
2. 查看GitHub文档：https://docs.github.com/
3. 检查网络连接和防火墙设置

---

**最后更新**: 2025-01-07
