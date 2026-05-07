# 亲子积分任务系统

一个家庭积分管理系统，家长发布任务、孩子完成任务赚积分、兑换奖品。无需登录，双端独立页面，开箱即用。

## 功能概览

### 家长端
- 任务管理：添加/删除一次性任务和周期性任务（每日刷新）
- 积分调整：手动增减积分，记录变更原因
- 积分历史：查看所有积分变动记录
- 奖品管理：添加/删除兑换奖品，设置所需积分

### 孩子端
- 任务展示：两列卡片布局，区分一次性与每日任务
- 完成任务：二次确认后获得积分，完成后卡片置灰保留
- 积分显示：顶部大号积分卡片，实时更新
- 奖品商店：积分足够即可兑换，自动扣减并记录

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go + Gin |
| 数据库 | SQLite（`modernc.org/sqlite`，纯 Go 无需 CGO） |
| 前端 | HTML + CSS + JavaScript（零依赖） |

## 快速开始

```bash
# 克隆项目
git clone https://github.com/yourname/credits.git
cd credits

# 安装依赖
go mod tidy

# 启动（默认 8080 端口）
go run main.go

# 或指定端口
go run main.go -port 9090
# 或通过环境变量
PORT=9090 go run main.go
```

启动后访问：
- **孩子端**: http://localhost:8080/child.html
- **家长端**: http://localhost:8080/parent.html

## 项目结构

```
credits/
├── main.go                 # 入口：路由 + 静态文件服务
├── go.mod / go.sum         # Go 模块依赖
├── database/
│   └── db.go               # SQLite 初始化 + 6 张表自动建表
├── models/
│   └── models.go           # 数据模型定义
├── handlers/
│   ├── tasks.go            # 任务 CRUD API
│   ├── child.go            # 孩子端 API（可用任务 + 完成任务）
│   ├── points.go           # 积分 API（查询/调整/历史）
│   └── redeem.go           # 奖品 API（CRUD + 兑换）
├── static/
│   ├── parent.html         # 家长端页面
│   └── child.html          # 孩子端页面
├── deploy.sh               # Linux 部署脚本
└── README.md
```

## API 接口

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/tasks` | 获取所有任务 |
| POST | `/api/tasks` | 创建任务 |
| DELETE | `/api/tasks/:id` | 删除任务（软删除） |
| GET | `/api/child/tasks` | 孩子端任务（含完成状态） |
| POST | `/api/child/complete-task` | 完成任务 |
| GET | `/api/points` | 获取当前积分 |
| POST | `/api/points/adjust` | 手动调整积分 |
| GET | `/api/points/history` | 积分历史记录 |
| GET | `/api/redeem-items` | 获取奖品列表 |
| POST | `/api/redeem-items` | 添加奖品 |
| DELETE | `/api/redeem-items/:id` | 删除奖品 |
| POST | `/api/redeem` | 兑换奖品 |

## 数据模型

| 表名 | 说明 |
|---|---|
| `tasks` | 任务（名称/积分/类型/软删除） |
| `child_progress` | 一次性任务完成记录 |
| `daily_task_status` | 周期性任务每日状态 |
| `point_history` | 积分变动历史 |
| `redeem_items` | 兑换奖品 |
| `current_points` | 当前积分快照 |

## 部署到 Linux

```bash
# 使用部署脚本
chmod +x deploy.sh

# 交叉编译 + 打包
./deploy.sh

# 上传到服务器
scp credits-linux-amd64.tar.gz user@server:~/

# 服务器上解压运行
ssh user@server
sudo mkdir -p /opt/credits
sudo tar -xzf credits-linux-amd64.tar.gz -C /opt/credits
cd /opt/credits && nohup ./credits -port 8080 > app.log 2>&1 &

# 或注册 systemd 服务实现开机自启
./deploy.sh service | sudo tee /etc/systemd/system/credits.service
sudo systemctl daemon-reload && sudo systemctl enable --now credits
```

## 核心设计

- **周期性任务每日刷新**：后端查询时根据当前日期 + `daily_task_status` 表判断今日完成状态，无需定时器
- **纯 Go SQLite**：使用 `modernc.org/sqlite`，交叉编译无需 CGO，Linux 上零依赖运行
- **事务保护**：所有积分操作均在事务中完成，防止数据不一致
- **软删除**：任务和奖品删除仅标记状态，不影响历史记录

## License

MIT
