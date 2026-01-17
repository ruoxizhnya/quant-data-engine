package kafka

import (
	"quant-data-engine/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestValidateMarketDataForKafka 测试Kafka市场数据验证
func TestValidateMarketDataForKafka(t *testing.T) {
	// 测试有效数据
	validData := models.MarketData{
		ID:        uuid.New().String(),
		Symbol:    "BTCUSDT",
		Price:     10000.0,
		Volume:    100.0,
		Timestamp: time.Now(),
		Source:    "binance",
	}
	err := validateMarketDataForKafka(validData)
	assert.NoError(t, err)

	// 测试无效数据
	invalidData := models.MarketData{
		ID:        uuid.New().String(),
		Symbol:    "",
		Price:     10000.0,
		Volume:    100.0,
		Timestamp: time.Time{},
		Source:    "",
	}
	err = validateMarketDataForKafka(invalidData)
	assert.Error(t, err)

	// 测试缺少Symbol
	noSymbolData := validData
	noSymbolData.Symbol = ""
	err = validateMarketDataForKafka(noSymbolData)
	assert.Error(t, err)

	// 测试缺少Timestamp
	noTimestampData := validData
	noTimestampData.Timestamp = time.Time{}
	err = validateMarketDataForKafka(noTimestampData)
	assert.Error(t, err)

	// 测试缺少Source
	noSourceData := validData
	noSourceData.Source = ""
	err = validateMarketDataForKafka(noSourceData)
	assert.Error(t, err)
}

// TestSendMarketData 测试发送市场数据
func TestSendMarketData(t *testing.T) {
	// 注意：这里是一个示例测试，实际测试需要连接到真实的Kafka集群
	// 或者使用mock库来模拟Kafka操作

	// 测试空数据
	var emptyData []models.MarketData
	err := (&KafkaProducer{}).SendMarketData(emptyData)
	assert.NoError(t, err)

	// 测试无效数据
	invalidData := []models.MarketData{
		{
			ID:        uuid.New().String(),
			Symbol:    "",
			Price:     10000.0,
			Volume:    100.0,
			Timestamp: time.Time{},
			Source:    "",
		},
	}
	err = (&KafkaProducer{}).SendMarketData(invalidData)
	assert.Error(t, err)
}
