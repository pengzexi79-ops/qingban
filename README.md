# 青伴 QingBan

> 情感陪伴 AI 应用 - 温暖每一刻

青伴是一款情感陪伴型应用，为用户提供可自定义的 AI 伴侣，通过长期记忆与个性化互动，带来温暖的陪伴体验。

## 功能特性

- **可自定义 AI 伴侣** - 自由设定名字、性格、角色背景和开场白
- **长期记忆库** - AI 伴侣可积累与用户的互动记忆
- **主动消息** - AI 伴侣可随机主动发送消息，增强陪伴感
- **好友列表** - 仿微信风格的聊天列表界面
- **数据管理** - 支持记忆导出与数据备份

## 技术栈

- **前端框架**: Vue 3 (CDN)
- **数据存储**: localStorage
- **样式**: 原生 CSS (移动端优先)
- **架构**: 单页应用 (SPA)

## 快速开始

### 直接运行

1. 克隆仓库
   `ash
   git clone https://github.com/yourusername/qingban.git
   cd qingban
   `

2. 使用任意静态服务器运行
   `ash
   # Python
   python -m http.server 8080
   
   # Node.js (npx)
   npx serve .
   `

3. 打开浏览器访问 http://localhost:8080

### 移动端体验

建议使用浏览器开发者工具切换到移动设备模式，或直接用手机浏览器访问。

## 项目结构

`
qingban/
├── index.html          # 主页面
├── css/
│   └── style.css       # 全局样式
├── js/
│   ├── app.js          # 主应用逻辑 (Vue 3)
│   ├── store.js        # 数据存储模块
│   └── components/     # 组件目录 (预留)
├── assets/             # 静态资源
├── README.md           # 项目说明
└── .gitignore          # Git 忽略规则
`

## 路线图

- [x] 基础框架搭建
- [x] 好友列表界面
- [x] 聊天对话功能
- [x] AI 伴侣管理 (添加/编辑/删除)
- [x] 本地数据持久化
- [x] 主动消息机制
- [ ] 接入真实 AI API
- [ ] 长期记忆系统增强
- [ ] 语音交互
- [ ] 情绪识别
- [ ] 多端同步
- [ ] 深色模式

## 开发说明

当前版本为初版框架，AI 回复为模拟回复。后续将接入 OpenAI / Claude / 其他大模型 API 实现真实对话。

## 许可证

MIT License
