#!/bin/bash

# GitHub项目初始化和上传脚本

echo "🚀 开始初始化GitHub仓库..."

# 检查是否已初始化git
if [ ! -d ".git" ]; then
    echo "🔧 初始化Git仓库..."
    git init
fi

# 添加所有文件
echo "📦 添加文件到Git..."
git add .

# 创建初始提交
echo "📝 创建初始提交..."
git commit -m "feat: 初始化AI Agent项目

- 实现Think-Execute自主决策循环
- 集成工具调用框架
- 实现RAG向量检索功能
- 支持多模型切换架构
- 添加SSE实时推送机制
- 配置Gin HTTP服务框架
- 完善测试用例和文档"

# 设置远程仓库（请替换为您的GitHub仓库URL）
echo "🔗 设置远程仓库..."
echo "请替换下面的URL为您自己的GitHub仓库地址："
echo "git remote add origin https://github.com/xuedingjie/ai-agent-go.git"

# 推送到GitHub
echo "📤 推送到GitHub..."
echo "请执行以下命令："
echo "git remote add origin https://github.com/xuedingjie/ai-agent-go.git"
echo "git branch -M main"
echo "git push -u origin main"

echo "✅ 完成！项目已准备好上传到GitHub"