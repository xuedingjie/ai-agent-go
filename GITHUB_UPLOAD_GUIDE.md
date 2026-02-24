# GitHub上传完整指南

## 📋 准备工作

### 1. 创建GitHub仓库
1. 登录到 [GitHub](https://github.com)
2. 点击右上角的 "+" 号，选择 "New repository"
3. 输入仓库名称（如：aigent）
4. 选择公开或私有
5. **不要**初始化README、.gitignore或license（我们已经有这些文件）
6. 点击 "Create repository"

### 2. 获取仓库URL
创建完成后，复制仓库的HTTPS URL，格式类似：
```
https://github.com/yourusername/your-repo-name.git
```

## 🚀 上传步骤

### 方法一：使用脚本（推荐）

#### Windows系统：
```cmd
# 运行批处理脚本
deploy.bat
```

#### Linux/Mac系统：
```bash
# 给脚本执行权限
chmod +x deploy.sh

# 运行脚本
./deploy.sh
```

### 方法二：手动上传

1. **初始化Git仓库**（如果还没有）：
```bash
git init
```

2. **添加所有文件**：
```bash
git add .
```

3. **创建初始提交**：
```bash
git commit -m "feat: 初始化AI Agent项目

- 实现Think-Execute自主决策循环
- 集成工具调用框架
- 实现RAG向量检索功能
- 支持多模型切换架构
- 添加SSE实时推送机制
- 配置Gin HTTP服务框架
- 完善测试用例和文档"
```

4. **设置远程仓库**（替换为您的实际URL）：
```bash
git remote add origin https://github.com/yourusername/your-repo-name.git
```

5. **设置主分支**：
```bash
git branch -M main
```

6. **推送到GitHub**：
```bash
git push -u origin main
```

## ✅ 验证上传

上传成功后，您应该在GitHub上看到：
- 所有源代码文件
- README.md 文档
- LICENSE 许可证
- CONTRIBUTING.md 贡献指南
- .gitignore 忽略文件配置
- Dockerfile 容器配置
- GitHub Actions CI/CD配置

## 🛠️ 后续配置

### 1. GitHub Actions设置
如果需要启用自动构建和部署：
1. 在GitHub仓库设置中配置Docker Hub密钥
2. 在仓库Settings → Webhooks中配置通知

### 2. 项目徽章
在README.md中更新徽章链接：
```markdown
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/your-repo-name)](https://goreportcard.com/report/github.com/yourusername/your-repo-name)
[![Build Status](https://github.com/yourusername/your-repo-name/workflows/Go%20CI/CD/badge.svg)](https://github.com/yourusername/your-repo-name/actions)
```

### 3. 项目描述
在GitHub仓库的About部分添加：
- 项目描述
- 网站链接（如果有）
- 相关话题标签

## 🎯 最佳实践

1. **保持提交信息规范**
2. **定期同步fork的上游仓库**
3. **使用GitHub Issues跟踪问题**
4. **编写清晰的文档**
5. **设置适当的分支保护规则**

## 🆘 常见问题

### Q: 推送被拒绝？
A: 确保远程URL正确，可能需要：
```bash
git remote set-url origin https://github.com/yourusername/your-repo-name.git
```

### Q: 文件过大无法推送？
A: 检查.gitignore是否正确配置，移除不必要的大文件

### Q: 需要更新已推送的内容？
A: 修改后重新提交：
```bash
git add .
git commit -m "fix: 修复描述"
git push
```

现在您的AI Agent项目已经准备好在GitHub上展示了！🎉