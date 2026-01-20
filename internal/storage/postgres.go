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

// StorageInterface 存储接口
type StorageInterface interface {
	SaveStockBasic(data []models.StockBasic) error
	GetStockBasic(limit int) ([]models.StockBasic, error)
	SaveMarketData(data []models.MarketData) error
	SaveBacktestData(data models.BacktestData) error
	GetMarketData(symbol string, limit int) ([]models.MarketData, error)
	GetHistoricalData(symbol string, startTime, endTime string) ([]models.MarketData, error)
	Close()
}

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

	// 创建股票基础信息表
	stockBasicTableSQL := `
	CREATE TABLE IF NOT EXISTS stock_basic (
		ts_code VARCHAR(20) PRIMARY KEY,
		symbol VARCHAR(20) NOT NULL,
		name VARCHAR(50) NOT NULL,
		area VARCHAR(20),
		industry VARCHAR(50),
		fullname VARCHAR(255),
		enname VARCHAR(255),
		cnspell VARCHAR(50),
		market VARCHAR(20),
		exchange VARCHAR(20),
		curr_type VARCHAR(10),
		list_status VARCHAR(10),
		list_date VARCHAR(10),
		delist_date VARCHAR(10),
		is_hs VARCHAR(10),
		act_name VARCHAR(100),
		act_ent_type VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_stock_basic_symbol ON stock_basic(symbol);
	CREATE INDEX IF NOT EXISTS idx_stock_basic_name ON stock_basic(name);
	CREATE INDEX IF NOT EXISTS idx_stock_basic_list_status ON stock_basic(list_status);
	CREATE INDEX IF NOT EXISTS idx_stock_basic_market ON stock_basic(market);
	`

	// 创建交易日历表
	tradeCalTableSQL := `
	CREATE TABLE IF NOT EXISTS trade_cal (
		exchange VARCHAR(20) NOT NULL,
		cal_date VARCHAR(10) NOT NULL,
		is_open VARCHAR(10) NOT NULL,
		pre_trade_date VARCHAR(10),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (exchange, cal_date)
	);

	CREATE INDEX IF NOT EXISTS idx_trade_cal_cal_date ON trade_cal(cal_date);
	CREATE INDEX IF NOT EXISTS idx_trade_cal_is_open ON trade_cal(is_open);
	`

	// 创建新股上市列表表
	newShareTableSQL := `
	CREATE TABLE IF NOT EXISTS new_share (
		ts_code VARCHAR(20) PRIMARY KEY,
		sub_code VARCHAR(20) NOT NULL,
		name VARCHAR(50) NOT NULL,
		ipo_date VARCHAR(10),
		issue_date VARCHAR(10),
		amount DOUBLE PRECISION,
		market_amount DOUBLE PRECISION,
		price DOUBLE PRECISION,
		pe DOUBLE PRECISION,
		limit_amount DOUBLE PRECISION,
		funds DOUBLE PRECISION,
		ballot DOUBLE PRECISION,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_new_share_ipo_date ON new_share(ipo_date);
	CREATE INDEX IF NOT EXISTS idx_new_share_issue_date ON new_share(issue_date);
	`

	// 创建上市公司基础信息表
	stockCompanyTableSQL := `
	CREATE TABLE IF NOT EXISTS stock_company (
		ts_code VARCHAR(20) PRIMARY KEY,
		com_name VARCHAR(100) NOT NULL,
		com_id VARCHAR(50),
		exchange VARCHAR(20),
		chairman VARCHAR(50),
		manager VARCHAR(50),
		secretary VARCHAR(50),
		reg_capital DOUBLE PRECISION,
		setup_date VARCHAR(10),
		province VARCHAR(20),
		city VARCHAR(20),
		introduction TEXT,
		website VARCHAR(255),
		email VARCHAR(100),
		office VARCHAR(255),
		employees INTEGER,
		main_business TEXT,
		business_scope TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_stock_company_exchange ON stock_company(exchange);
	CREATE INDEX IF NOT EXISTS idx_stock_company_province ON stock_company(province);
	`

	// 创建上市公司管理层表
	stkManagersTableSQL := `
	CREATE TABLE IF NOT EXISTS stk_managers (
		ts_code VARCHAR(20) NOT NULL,
		ann_date VARCHAR(10) NOT NULL,
		name VARCHAR(50) NOT NULL,
		gender VARCHAR(10),
		lev VARCHAR(20),
		title VARCHAR(100),
		edu VARCHAR(50),
		national VARCHAR(50),
		birthday VARCHAR(20),
		begin_date VARCHAR(10),
		end_date VARCHAR(10),
		resume TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (ts_code, ann_date, name)
	);

	CREATE INDEX IF NOT EXISTS idx_stk_managers_ts_code ON stk_managers(ts_code);
	CREATE INDEX IF NOT EXISTS idx_stk_managers_ann_date ON stk_managers(ann_date);
	`

	// 创建管理层薪酬和持股表
	stkRewardsTableSQL := `
	CREATE TABLE IF NOT EXISTS stk_rewards (
		ts_code VARCHAR(20) NOT NULL,
		ann_date VARCHAR(10) NOT NULL,
		end_date VARCHAR(10) NOT NULL,
		name VARCHAR(50) NOT NULL,
		title VARCHAR(100),
		reward DOUBLE PRECISION,
		hold_vol DOUBLE PRECISION,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (ts_code, ann_date, end_date, name)
	);

	CREATE INDEX IF NOT EXISTS idx_stk_rewards_ts_code ON stk_rewards(ts_code);
	CREATE INDEX IF NOT EXISTS idx_stk_rewards_end_date ON stk_rewards(end_date);
	`

	// 创建A股日线行情表
	dailyTableSQL := `
	CREATE TABLE IF NOT EXISTS daily (
		ts_code VARCHAR(20) NOT NULL,
		trade_date VARCHAR(10) NOT NULL,
		open DOUBLE PRECISION,
		high DOUBLE PRECISION,
		low DOUBLE PRECISION,
		close DOUBLE PRECISION,
		pre_close DOUBLE PRECISION,
		change DOUBLE PRECISION,
		pct_chg DOUBLE PRECISION,
		vol DOUBLE PRECISION,
		amount DOUBLE PRECISION,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (ts_code, trade_date)
	);

	CREATE INDEX IF NOT EXISTS idx_daily_ts_code ON daily(ts_code);
	CREATE INDEX IF NOT EXISTS idx_daily_trade_date ON daily(trade_date);
	CREATE INDEX IF NOT EXISTS idx_daily_pct_chg ON daily(pct_chg);
	`

	// 执行SQL语句
	if _, err := s.pool.Exec(context.Background(), marketDataTableSQL); err != nil {
		return fmt.Errorf("failed to create market_data table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), backtestDataTableSQL); err != nil {
		return fmt.Errorf("failed to create backtest_data table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), stockBasicTableSQL); err != nil {
		return fmt.Errorf("failed to create stock_basic table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), tradeCalTableSQL); err != nil {
		return fmt.Errorf("failed to create trade_cal table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), newShareTableSQL); err != nil {
		return fmt.Errorf("failed to create new_share table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), stockCompanyTableSQL); err != nil {
		return fmt.Errorf("failed to create stock_company table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), stkManagersTableSQL); err != nil {
		return fmt.Errorf("failed to create stk_managers table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), stkRewardsTableSQL); err != nil {
		return fmt.Errorf("failed to create stk_rewards table: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), dailyTableSQL); err != nil {
		return fmt.Errorf("failed to create daily table: %w", err)
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

// SaveStockBasic 保存股票基础信息
func (s *PostgresStorage) SaveStockBasic(data []models.StockBasic) error {
	if len(data) == 0 {
		return nil
	}

	// 使用批量插入
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// 直接执行SQL语句
	query := `
		INSERT INTO stock_basic (
			ts_code, symbol, name, area, industry, fullname, enname, cnspell, 
			market, exchange, curr_type, list_status, list_date, delist_date, is_hs, 
			act_name, act_ent_type, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, CURRENT_TIMESTAMP
		) ON CONFLICT (ts_code) DO UPDATE SET
			symbol = $2, name = $3, area = $4, industry = $5, fullname = $6, enname = $7, cnspell = $8, 
			market = $9, exchange = $10, curr_type = $11, list_status = $12, list_date = $13, delist_date = $14, is_hs = $15, 
			act_name = $16, act_ent_type = $17, updated_at = CURRENT_TIMESTAMP
	`

	for _, d := range data {
		_, err := tx.Exec(context.Background(), query,
			d.TSCode, d.Symbol, d.Name, d.Area, d.Industry, d.Fullname, d.Enname, d.Cnspell,
			d.Market, d.Exchange, d.CurrType, d.ListStatus, d.ListDate, d.DelistDate, d.IsHS,
			d.ActName, d.ActEntType,
		)
		if err != nil {
			return fmt.Errorf("failed to insert stock basic data: %w", err)
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logrus.Infof("Saved %d stock basic data records", len(data))
	return nil
}

// GetStockBasic 获取股票基础信息
func (s *PostgresStorage) GetStockBasic(limit int) ([]models.StockBasic, error) {
	rows, err := s.pool.Query(context.Background(), `
		SELECT ts_code, symbol, name, area, industry, fullname, enname, cnspell, 
			market, exchange, curr_type, list_status, list_date, delist_date, is_hs, 
			act_name, act_ent_type, created_at, updated_at
		FROM stock_basic
		ORDER BY ts_code ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock basic data: %w", err)
	}
	defer rows.Close()

	var data []models.StockBasic
	for rows.Next() {
		var d models.StockBasic
		if err := rows.Scan(
			&d.TSCode, &d.Symbol, &d.Name, &d.Area, &d.Industry, &d.Fullname, &d.Enname, &d.Cnspell,
			&d.Market, &d.Exchange, &d.CurrType, &d.ListStatus, &d.ListDate, &d.DelistDate, &d.IsHS,
			&d.ActName, &d.ActEntType, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan stock basic data: %w", err)
		}
		data = append(data, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock basic data rows: %w", err)
	}

	return data, nil
}
