@echo off
title GitHub项目部署脚本

echo 🚀 开始初始化GitHub仓库...

REM 检查是否已初始化git
if not exist ".git" (
    echo 🔧 初始化Git仓库...
    git init
)

REM 添加所有文件
echo 📦 添加文件到Git...
git add .

REM 创建初始提交
echo 📝 创建初始提交...
git commit -m "feat: 初始化AI Agent项目
- 实现Think-Execute自主决策循环
- 集成工具调用框架
- 实现RAG向量检索功能
- 支持多模型切换架构
- 添加SSE实时推送机制
- 配置Gin HTTP服务框架
- 完善测试用例和文档"

REM 设置远程仓库提示
echo.
echo 🔗 设置远程仓库...
echo 请将下面的URL替换为您自己的GitHub仓库地址：
echo git remote add origin https://github.com/yourusername/your-repo-name.git
echo.

REM 推送到GitHub提示
echo 📤 推送到GitHub...
echo 请依次执行以下命令：
echo git remote add origin https://github.com/yourusername/your-repo-name.git
echo git branch -M main
echo git push -u origin main
echo.

echo ✅ 完成！项目已准备好上传到GitHub
echo.
pause