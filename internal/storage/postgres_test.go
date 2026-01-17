package storage

import (
	"quant-data-engine/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPostgresStorage 模拟PostgresStorage
type MockPostgresStorage struct {
	mock.Mock
}

// TestValidateMarketData 测试市场数据验证
func TestValidateMarketData(t *testing.T) {
	// 测试有效数据
	validData := models.MarketData{
		ID:        uuid.New().String(),
		Symbol:    "BTCUSDT",
		Price:     10000.0,
		Volume:    100.0,
		Timestamp: time.Now(),
		Source:    "binance",
	}
	err := validateMarketData(validData)
	assert.NoError(t, err)

	// 测试无效数据
	invalidData := models.MarketData{
		ID:        "",
		Symbol:    "",
		Price:     0.0,
		Volume:    -1.0,
		Timestamp: time.Time{},
		Source:    "",
	}
	err = validateMarketData(invalidData)
	assert.Error(t, err)

	// 测试价格为负数
	negativePriceData := validData
	negativePriceData.Price = -10000.0
	err = validateMarketData(negativePriceData)
	assert.Error(t, err)

	// 测试交易量为负数
	negativeVolumeData := validData
	negativeVolumeData.Volume = -100.0
	err = validateMarketData(negativeVolumeData)
	assert.Error(t, err)

	// 测试时间戳为零值
	zeroTimestampData := validData
	zeroTimestampData.Timestamp = time.Time{}
	err = validateMarketData(zeroTimestampData)
	assert.Error(t, err)
}

// TestValidateBacktestData 测试回测数据验证
func TestValidateBacktestData(t *testing.T) {
	// 测试有效数据
	validData := models.BacktestData{
		ID:        uuid.New().String(),
		Symbol:    "BTCUSDT",
		Strategy:  "MA Cross",
		StartDate: time.Now().AddDate(0, -1, 0),
		EndDate:   time.Now(),
		Results:   `{"profit": 12.5, "drawdown": 5.2}`,
		Timestamp: time.Now(),
	}
	err := validateBacktestData(validData)
	assert.NoError(t, err)

	// 测试无效数据
	invalidData := models.BacktestData{
		ID:        "",
		Symbol:    "",
		Strategy:  "",
		StartDate: time.Time{},
		EndDate:   time.Time{},
		Results:   "",
		Timestamp: time.Time{},
	}
	err = validateBacktestData(invalidData)
	assert.Error(t, err)

	// 测试结束日期早于开始日期
	wrongDateData := validData
	wrongDateData.EndDate = validData.StartDate.AddDate(0, -1, 0)
	err = validateBacktestData(wrongDateData)
	assert.Error(t, err)
}

// TestSaveMarketData 测试保存市场数据
func TestSaveMarketData(t *testing.T) {
	// 注意：这里是一个示例测试，实际测试需要连接到真实的数据库
	// 或者使用mock库来模拟数据库操作

	// 测试空数据
	var emptyData []models.MarketData
	err := (&PostgresStorage{}).SaveMarketData(emptyData)
	assert.NoError(t, err)

	// 测试无效数据
	invalidData := []models.MarketData{
		{
			ID:        "",
			Symbol:    "",
			Price:     0.0,
			Volume:    -1.0,
			Timestamp: time.Time{},
			Source:    "",
		},
	}
	err = (&PostgresStorage{}).SaveMarketData(invalidData)
	assert.Error(t, err)
}

// TestSaveBacktestData 测试保存回测数据
func TestSaveBacktestData(t *testing.T) {
	// 注意：这里是一个示例测试，实际测试需要连接到真实的数据库
	// 或者使用mock库来模拟数据库操作

	// 测试无效数据
	invalidData := models.BacktestData{
		ID:        "",
		Symbol:    "",
		Strategy:  "",
		StartDate: time.Time{},
		EndDate:   time.Time{},
		Results:   "",
		Timestamp: time.Time{},
	}
	err := (&PostgresStorage{}).SaveBacktestData(invalidData)
	assert.Error(t, err)
}
