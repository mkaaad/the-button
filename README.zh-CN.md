# The Button - 实时按钮挑战游戏

> 英文版本：[README.md](README.md)

一个全栈实时按钮游戏，带有排行榜排名功能。玩家通过按下重置60秒倒计时的虚拟按钮进行竞争，目标是在排行榜上获得最低的剩余时间。

## 功能特性

- **短信认证** - 通过阿里云短信进行手机验证，支持图形验证码
- **实时WebSocket** - 实时游戏状态更新和排行榜广播
- **排行榜系统** - 基于Redis有序集合的排名，使用LT（小于）更新策略
- **玩家冷却时间** - 每个用户5秒冷却锁，防止刷屏
- **游戏时间窗口** - 可配置的开始/结束时间，支持时区设置
- **现代化UI** - 响应式React界面，使用Tailwind CSS和Radix组件
- **Docker支持** - 完整的容器化部署，支持Redis持久化

## 技术栈

**前端**
- React 19 + TypeScript
- Vite 7
- Tailwind CSS + shadcn风格UI组件（Radix UI）
- WebSocket API

**后端**
- Go 1.24.4 + Gin Web框架
- SQLite（GORM）+ Redis（有序集合）
- 阿里云短信集成
- Gorilla WebSocket

**基础设施**
- Docker Compose
- Nginx（前端服务）
- Redis（排行榜和会话存储）

## 快速开始

### 方案一：Docker Compose（推荐）
```bash
# 复制环境变量示例文件
cp be/.env.example be/.env
cp fe/.env.example fe/.env

# 编辑be/.env，填入短信服务凭证
# 按需编辑fe/.env（默认配置适用于Docker环境）

# 启动所有服务
docker-compose up
```

访问游戏：`http://localhost:5173`

### 方案二：手动安装
```bash
# 1. 启动Redis
docker run -d -p 6379:6379 redis:7-alpine

# 2. 配置后端
cd be
cp .env.example .env
# 编辑.env文件，填入短信凭证和Redis设置

# 3. 运行后端
go run cmd/main.go

# 4. 配置前端
cd ../fe
cp .env.example .env
# 设置 VITE_API_BASE_URL=http://localhost:8080
# 设置 VITE_WS_URL=ws://localhost:8080/ws

# 5. 运行前端
npm install
npm run dev
```

## 环境变量配置

### 后端 (`be/.env`)
| 变量名 | 描述 | 是否必需 |
|--------|------|----------|
| `ACCESS_KEY_ID` | 阿里云短信API密钥 | 是 |
| `ACCESS_KEY_SECRET` | 阿里云短信API密钥 | 是 |
| `REDIS_ADDR` | Redis连接地址（主机:端口） | 是 |
| `REDIS_PASSWORD` | Redis密码（如有设置） | 否 |
| `REDIS_DB` | Redis数据库索引 | 否（默认：0） |
| `GAME_START_TIME` | 游戏开始时间（亚洲/上海时区） | 是 |
| `GAME_END_TIME` | 游戏结束时间（亚洲/上海时区） | 是 |
| `GAME_TIMEZONE` | 游戏时间时区 | 否（默认：Asia/Shanghai） |

### 前端 (`fe/.env`)
| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `VITE_API_BASE_URL` | API基础URL | `/api` |
| `VITE_WS_URL` | WebSocket端点 | `/ws` |

## API接口

### HTTP API
- `GET /sms/captcha` - 获取短信验证图形验证码
- `POST /sms/code` - 发送短信验证码（需要图形验证码）
- `POST /sms/verify` - 验证短信验证码并创建会话

### WebSocket
连接地址：`ws://主机:端口/ws?session_id={会话ID}`

#### 客户端命令
- `"1"` - 请求当前时间
- `"2"` - 请求排行榜
- `"3"` - 按下按钮（需要认证会话）

#### 服务端消息
```json
{"type": "time", "data": {"time": 1676543210000}}
{"type": "leaderboard", "data": {"entries": [...]}}
{"type": "button_press", "data": {"username": "玩家1"}}
{"type": "lock"}      // 5秒冷却中
{"type": "pending"}   // 游戏尚未开始
{"type": "finished"}  // 游戏已结束
{"type": "unauthorized"} // 无效会话
```

## 游戏规则

1. **60秒倒计时** - 按下按钮将计时器重置为60秒
2. **排行榜排名** - 玩家按剩余时间排名（时间越短排名越高）
3. **5秒冷却时间** - 每个玩家按下按钮后需等待5秒才能再次按下
4. **时间窗口限制** - 游戏仅在 `GAME_START_TIME` 和 `GAME_END_TIME` 之间开放
5. **短信认证** - 需要手机验证才能参与游戏

## 开发指南

### 项目结构
```
the-button/
├── fe/                 # React前端
│   ├── src/
│   │   ├── components/ui/  # Radix基础UI组件
│   │   ├── lib/utils.ts    # 工具函数
│   │   └── App.tsx         # 主应用
│   └── package.json
├── be/                 # Go后端
│   ├── cmd/main.go          # 入口点
│   ├── router/router.go     # HTTP路由
│   ├── api/                 # HTTP和WebSocket处理器
│   ├── service/             # 业务逻辑
│   ├── dao/                 # 数据访问层（Redis、SQLite）
│   ├── model/               # 数据模型
│   ├── config/              # 配置管理
│   ├── errorx/              # 错误定义
│   ├── middleware/          # CORS中间件
│   └── connx/               # 连接池
└── docker-compose.yml
```

### 开发命令
```bash
# 前端
cd fe
npm run dev      # 开发服务器
npm run build    # 生产构建
npm run lint     # 代码检查

# 后端
cd be
go run cmd/main.go
go test ./...
go mod tidy

# 压力测试
cd be/test
go run main.go   # 需要后端正在运行
```

## 部署说明

### Docker Compose生产部署
```bash
docker-compose up -d
```

服务说明：
- **Redis**（端口6379）- 排行榜和会话存储
- **Backend**（端口8080）- Go应用，包含健康检查
- **Frontend**（端口5173）- Nginx服务React构建文件

### 健康检查
- 后端：`GET /sms/captcha`（返回200表示健康）
- Redis：docker-compose中自动健康检查

## 许可证

MIT许可证