# SSH配置步骤 - 快速指南

## ✅ 已生成SSH密钥

SSH密钥已生成，请按照以下步骤完成配置：

## 📋 步骤1：复制SSH公钥

公钥内容已显示在上方，请复制完整内容（从 `ssh-ed25519` 开始到邮箱结束）

## 📋 步骤2：添加到GitHub

1. 访问：https://github.com/settings/keys
2. 点击右上角 "New SSH key" 按钮
3. 填写信息：
   - **Title**: `toogo_v2_windows` （或任意名称）
   - **Key**: 粘贴刚才复制的公钥内容
   - **Key type**: 选择 `Authentication Key`
4. 点击 "Add SSH key"
5. 可能需要输入GitHub密码确认

## 📋 步骤3：更改Git远程URL为SSH

在PowerShell中执行：

```powershell
cd D:\go\src\toogo_v2
git remote set-url origin git@github.com:lilei800/toogo_v2.git
```

## 📋 步骤4：测试SSH连接

```powershell
ssh -T git@github.com
```

**应该看到：**
```
Hi lilei800! You've successfully authenticated, but GitHub does not provide shell access.
```

如果看到这个，说明SSH配置成功！

## 📋 步骤5：推送代码

```powershell
git push origin main
```

这次应该可以成功推送了！

---

## 🔍 如果SSH测试失败

### 问题1：提示 "Permission denied"

**解决方法：**
1. 检查公钥是否正确添加到GitHub
2. 确保复制的是公钥（.pub文件），不是私钥
3. 重新添加公钥到GitHub

### 问题2：提示 "Host key verification failed"

**解决方法：**
```powershell
# 清除已知主机
ssh-keygen -R github.com

# 重新测试
ssh -T git@github.com
# 输入 yes 接受GitHub的host key
```

### 问题3：仍然连接失败

**解决方法：**
1. 检查防火墙设置
2. 检查网络连接
3. 尝试使用VPN或代理

---

## 🎯 完成后的验证

```powershell
# 1. 检查远程URL（应该是SSH格式）
git remote -v
# 应该显示: git@github.com:lilei800/toogo_v2.git

# 2. 测试SSH连接
ssh -T git@github.com

# 3. 推送代码
git push origin main
```

---

**完成这些步骤后，你的代码就可以成功推送到GitHub了！** 🎉
