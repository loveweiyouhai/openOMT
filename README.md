# openOMT

一款现代化的服务器运维管理工具，支持 SFTP/FTP 文件管理和 SSH 终端，使用 Wails 构建跨平台桌面应用。

## 技术栈

| 类别 | 技术 |
|------|------|
| 后端 | Go 1.21+、Wails v2 |
| 前端 | Vue 3、Vite、Element Plus、xterm.js |
| 协议 | SFTP (golang.org/x/crypto/ssh)、FTP (github.com/jlaffaye/ftp) |
| 存储 | SQLite + AES-GCM 加密 |
| 样式 | SCSS、CSS Variables (支持主题切换) |

## 功能特性

### 连接管理
- 支持 SFTP / FTP 双协议
- 连接信息本地加密存储 (SQLite + AES)
- 支持同时连接多个服务器

### 文件管理
- 目录浏览、新建文件夹/文件
- 文件上传、下载、删除
- 右键菜单快捷操作

### SSH 终端
- 基于 xterm.js 的交互式终端
- 每个服务器支持多终端标签
- 终端会话在服务器切换时保持状态

### 界面
- 暗色 / 亮色主题切换
- 响应式布局
- 现代化 UI 设计

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+
- [Wails CLI v2](https://wails.io/)
- Windows 需安装 WebView2

### 开发模式
```bash
wails dev
```

### 构建
```bash
wails build
```

生成文件：`build/bin/openOMT.exe`

## 项目结构

```
openOMT/
├── main.go                 # 应用入口
├── app.go                  # Go 后端逻辑
├── wails.json              # Wails 配置
├── internal/
│   ├── conn/               # 连接池、SFTP/FTP 操作、Shell 会话
│   └── store/              # SQLite 存储、密码加密
├── frontend/
│   └── src/
│       ├── App.vue         # 主组件
│       ├── components/     # Vue 组件
│       └── styles/         # SCSS 样式
└── build/                  # 构建输出
```

## 许可证

MIT
