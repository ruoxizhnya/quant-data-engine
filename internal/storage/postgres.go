package storage

import (
	"context"
	"fmt"
	"quant-data-engine/internal/config"
	"quant-data-engine/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// PostgresStorage PostgreSQL存储实现
type PostgresStorage struct {
	pool *pgxpool.Pool
}

// NewPostgresStorage 创建PostgreSQL存储
func NewPostgresStorage() (*PostgresStorage, error) {
	cfg := config.AppConfig

	// 构建数据库连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	// 创建连接池配置
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		logrus.Errorf("Failed to parse database config: %v", err)
		return nil, err
	}

	// 设置连接池参数
	poolConfig.MaxConns = int32(cfg.DBMaxConns)
	poolConfig.MinConns = int32(cfg.DBMaxConns / 2)
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// 创建连接池
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logrus.Errorf("Failed to connect to database: %v", err)
		return nil, err
	}

	// 测试连接
	if err := pool.Ping(context.Background()); err != nil {
		logrus.Errorf("Failed to ping database: %v", err)
		return nil, err
	}

	storage := &PostgresStorage{
		pool: pool,
	}

	// 初始化表结构
	if err := storage.initTables(); err != nil {
		logrus.Errorf("Failed to initialize tables: %v", err)
		return nil, err
	}

	logrus.Info("Connected to PostgreSQL database successfully")
	return storage, nil
}

// initTables 初始化表结构
func (s *PostgresStorage) initTables() error {
	// 创建市场数据表
	marketDataTableSQL := `
	CREATE TABLE IF NOT EXISTS market_data (
		id VARCHAR(36) PRIMARY KEY,
		symbol VARCHAR(20) NOT NULL,
		price DOUBLE PRECISION NOT NULL,
		volume DOUBLE PRECISION NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		source VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_market_data_symbol ON market_data(symbol);
	CREATE INDEX IF NOT EXISTS idx_market_data_timestamp ON market_data(timestamp);
	CREATE INDEX IF NOT EXISTS idx_market_data_source ON market_data(source);
	`

	// 创建回测数据表
	backtestDataTableSQL := `
	CREATE TABLE IF NOT EXISTS backtest_data (
		id VARCHAR(36) PRIMARY KEY,
		symbol VARCHAR(20) NOT NULL,
		strategy VARCHAR(50) NOT NULL,
		start_date TIMESTAMP NOT NULL,
		end_date TIMESTAMP NOT NULL,
		results JSONB NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_backtest_data_symbol ON backtest_data(symbol);
	CREATE INDEX IF NOT EXISTS idx_backtest_data_strategy ON backtest_data(strategy);
	CREATE INDEX IF NOT EXISTS idx_backtest_data_start_date ON backtest_data(start_date);
	CREATE INDEX IF NOT EXISTS idx_backtest_data_end_date ON backtest_data(end_date);
	`

	// 执行SQL语句
	if _, err := s.pool.Exec(context.Background(), marketDataTableSQL); err != nil {
		return fmt.Errorf("failed to create market_data table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), backtestDataTableSQL); err != nil {
		return fmt.Errorf("failed to create backtest_data table: %w", err)
	}

	return nil
}

// SaveMarketData 保存市场数据
func (s *PostgresStorage) SaveMarketData(data []models.MarketData) error {
	if len(data) == 0 {
		return nil
	}

	// 验证数据
	for i, d := range data {
		if err := validateMarketData(d); err != nil {
			return fmt.Errorf("invalid market data at index %d: %w", i, err)
		}
	}

	// 使用批量插入
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// 直接执行SQL语句，不使用预处理语句
	query := `
		INSERT INTO market_data (id, symbol, price, volume, timestamp, source)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`

	for _, d := range data {
		_, err := tx.Exec(context.Background(), query, d.ID, d.Symbol, d.Price, d.Volume, d.Timestamp, d.Source)
		if err != nil {
			return fmt.Errorf("failed to insert market data: %w", err)
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logrus.Infof("Saved %d market data records", len(data))
	return nil
}

// validateMarketData 验证市场数据
func validateMarketData(data models.MarketData) error {
	if data.ID == "" {
		return fmt.Errorf("id is required")
	}
	if data.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if data.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if data.Volume < 0 {
		return fmt.Errorf("volume cannot be negative")
	}
	if data.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	if data.Source == "" {
		return fmt.Errorf("source is required")
	}
	return nil
}

// SaveBacktestData 保存回测数据
func (s *PostgresStorage) SaveBacktestData(data models.BacktestData) error {
	// 验证数据
	if err := validateBacktestData(data); err != nil {
		return fmt.Errorf("invalid backtest data: %w", err)
	}

	// 使用事务
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		INSERT INTO backtest_data (id, symbol, strategy, start_date, end_date, results, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			symbol = $2,
			strategy = $3,
			start_date = $4,
			end_date = $5,
			results = $6,
			timestamp = $7
	`, data.ID, data.Symbol, data.Strategy, data.StartDate, data.EndDate, data.Results, data.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to save backtest data: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logrus.Infof("Saved backtest data for symbol %s", data.Symbol)
	return nil
}

// validateBacktestData 验证回测数据
func validateBacktestData(data models.BacktestData) error {
	if data.ID == "" {
		return fmt.Errorf("id is required")
	}
	if data.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if data.Strategy == "" {
		return fmt.Errorf("strategy is required")
	}
	if data.StartDate.IsZero() {
		return fmt.Errorf("start_date is required")
	}
	if data.EndDate.IsZero() {
		return fmt.Errorf("end_date is required")
	}
	if data.EndDate.Before(data.StartDate) {
		return fmt.Errorf("end_date must be after start_date")
	}
	if data.Results == "" {
		return fmt.Errorf("results is required")
	}
	if data.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}

// GetMarketData 获取市场数据
func (s *PostgresStorage) GetMarketData(symbol string, limit int) ([]models.MarketData, error) {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, symbol, price, volume, timestamp, source
		FROM market_data
		WHERE symbol = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query market data: %w", err)
	}
	defer rows.Close()

	var data []models.MarketData
	for rows.Next() {
		var d models.MarketData
		if err := rows.Scan(&d.ID, &d.Symbol, &d.Price, &d.Volume, &d.Timestamp, &d.Source); err != nil {
			return nil, fmt.Errorf("failed to scan market data: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating market data rows: %w", err)
	}

	return data, nil
}

// GetHistoricalData 获取历史数据
func (s *PostgresStorage) GetHistoricalData(symbol string, startTime, endTime string) ([]models.MarketData, error) {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, symbol, price, volume, timestamp, source
		FROM market_data
		WHERE symbol = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
	`, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query historical data: %w", err)
	}
	defer rows.Close()

	var data []models.MarketData
	for rows.Next() {
		var d models.MarketData
		if err := rows.Scan(&d.ID, &d.Symbol, &d.Price, &d.Volume, &d.Timestamp, &d.Source); err != nil {
			return nil, fmt.Errorf("failed to scan historical data: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating historical data rows: %w", err)
	}

	return data, nil
}

// Close 关闭存储
func (s *PostgresStorage) Close() {
	if s.pool != nil {
		s.pool.Close()
		logrus.Info("PostgreSQL connection pool closed")
	}
}
