package datasource

import (
	"math/rand"
	"quant-data-engine/internal/models"
	"time"

	"github.com/google/uuid"
)

// ExchangeDataSource 交易所数据源实现
type ExchangeDataSource struct {
	name      string
	apiKey    string
	apiSecret string
}

// NewExchangeDataSource 创建交易所数据源
func NewExchangeDataSource(name, apiKey, apiSecret string) *ExchangeDataSource {
	return &ExchangeDataSource{
		name:      name,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

// GetMarketData 获取市场数据
func (e *ExchangeDataSource) GetMarketData(symbol string) ([]models.MarketData, error) {
	// 模拟获取市场数据
	rand.Seed(time.Now().UnixNano())

	data := []models.MarketData{
		{
			ID:        uuid.New().String(),
			Symbol:    symbol,
			Price:     1000 + rand.Float64()*100,
			Volume:    10000 + rand.Float64()*1000,
			Timestamp: time.Now(),
			Source:    e.name,
		},
	}

	return data, nil
}

// GetHistoricalData 获取历史数据
func (e *ExchangeDataSource) GetHistoricalData(symbol string, startTime, endTime string) ([]models.MarketData, error) {
	// 模拟获取历史数据
	rand.Seed(time.Now().UnixNano())

	var data []models.MarketData

	// 解析时间
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		start = time.Now().Add(-24 * time.Hour)
	}

	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		end = time.Now()
	}

	// 生成模拟数据
	current := start
	for current.Before(end) {
		data = append(data, models.MarketData{
			ID:        uuid.New().String(),
			Symbol:    symbol,
			Price:     1000 + rand.Float64()*100,
			Volume:    10000 + rand.Float64()*1000,
			Timestamp: current,
			Source:    e.name,
		})
		current = current.Add(1 * time.Hour)
	}

	return data, nil
}

// Name 获取数据源名称
func (e *ExchangeDataSource) Name() string {
	return e.name
}
