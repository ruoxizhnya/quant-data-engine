# Quant Data Engine

基于Go的量化数据引擎，用于从交易所或其他数据源获取数据，存储到PostgreSQL数据库，然后发送到Kafka topic，并提供REST API用于读取Parquet格式的回测数据。

## 项目结构

```
quant-data-engine/
├── cmd/
│   └── data-engine/       # 主应用程序
├── internal/
│   ├── api/               # REST API服务
│   ├── config/            # 配置管理
│   ├── datasource/        # 数据源接口和实现
│   ├── kafka/             # Kafka消息发送
│   ├── models/            # 数据模型
│   └── storage/           # 数据库存储
├── pkg/
│   └── utils/             # 工具函数
├── .env.example           # 示例配置文件
├── go.mod                 # Go模块定义
├── go.sum                 # 依赖版本锁定
└── README.md              # 项目说明
```

## 功能特性

1. **数据获取**：从交易所API获取市场数据
2. **数据存储**：将数据存储到PostgreSQL数据库
3. **消息发送**：将数据发送到Kafka topic
4. **REST API**：提供API接口用于读取回测数据
5. **Parquet支持**：支持Parquet格式的回测数据存储和读取

## 安装说明

### 前提条件

- Go 1.20+
- PostgreSQL 13+
- Kafka 2.8+

### 安装步骤

1. **克隆项目**

```bash
git clone https://github.com/yourusername/quant-data-engine.git
cd quant-data-engine
```

2. **安装依赖**

```bash
go mod tidy
```

3. **配置环境变量**

```bash
cp .env.example .env
# 编辑.env文件，填写相应的配置
```

4. **构建项目**

```bash
go build -o bin/data-engine ./cmd/data-engine
```

5. **运行项目**

```bash
./bin/data-engine
```

## 配置说明

| 配置项 | 说明 | 默认值 |
|-------|------|-------|
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户 | postgres |
| DB_PASSWORD | 数据库密码 | password |
| DB_NAME | 数据库名称 | quant_data |
| KAFKA_BROKERS | Kafka brokers | localhost:9092 |
| KAFKA_TOPIC | Kafka topic | quant_data |
| API_PORT | API服务端口 | 8080 |
| EXCHANGE_API_KEY | 交易所API密钥 | (空) |
| EXCHANGE_API_SECRET | 交易所API密钥 | (空) |
| LOG_LEVEL | 日志级别 | info |

## API接口

### 健康检查

```
GET /api/health
```

### 获取回测数据

```
GET /api/backtest/data?symbol=BTCUSDT
```

### 获取Parquet格式回测数据

```
GET /api/backtest/parquet?symbol=BTCUSDT&start_date=2023-01-01&end_date=2023-01-31
```

### 获取市场数据

```
GET /api/market/data?symbol=BTCUSDT&limit=10
```

## 使用示例

### 1. 启动数据引擎

```bash
./bin/data-engine
```

### 2. 获取市场数据

```bash
curl "http://localhost:8080/api/market/data?symbol=BTCUSDT&limit=10"
```

### 3. 获取Parquet格式回测数据

```bash
curl "http://localhost:8080/api/backtest/parquet?symbol=BTCUSDT&start_date=2023-01-01&end_date=2023-01-31"
```

## 数据流程

1. **数据获取**：从交易所API获取市场数据
2. **数据存储**：将数据存储到PostgreSQL数据库
3. **消息发送**：将数据发送到Kafka topic
4. **数据查询**：通过REST API查询回测数据和市场数据
5. **Parquet处理**：生成和读取Parquet格式的回测数据

## 扩展指南

### 添加新的数据源

1. 实现 `datasource.DataSource` 接口
2. 在 `main.go` 中注册数据源

### 添加新的API接口

1. 在 `internal/api/api.go` 中添加新的路由和处理函数

### 自定义数据模型

1. 在 `internal/models/models.go` 中定义新的数据模型
2. 在相应的存储和API模块中使用新模型

## 注意事项

1. 确保PostgreSQL和Kafka服务已启动
2. 填写正确的数据库连接信息和Kafka配置
3. 对于生产环境，建议修改默认的数据库密码和API密钥
4. 定期清理数据库中的历史数据，避免数据量过大