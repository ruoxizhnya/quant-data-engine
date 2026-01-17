package datasource

import (
	"testing"
)

func TestDataSourceFactory(t *testing.T) {
	// 创建数据源工厂
	factory := NewDataSourceFactory()

	// 注册数据源
	factory.Register("binance", NewExchangeDataSource("binance", "key", "secret"))
	factory.Register("okx", NewExchangeDataSource("okx", "key", "secret"))

	// 测试获取数据源
	binance := factory.GetDataSource("binance")
	if binance == nil {
		t.Error("Expected binance datasource to be registered")
	}

	okx := factory.GetDataSource("okx")
	if okx == nil {
		t.Error("Expected okx datasource to be registered")
	}

	// 测试获取不存在的数据源
	nonexistent := factory.GetDataSource("nonexistent")
	if nonexistent != nil {
		t.Error("Expected nonexistent datasource to be nil")
	}
}

func TestExchangeDataSource(t *testing.T) {
	// 创建交易所数据源
	source := NewExchangeDataSource("binance", "key", "secret")

	// 测试数据源名称
	if source.Name() != "binance" {
		t.Errorf("Expected datasource name to be 'binance', got '%s'", source.Name())
	}

	// 测试获取市场数据
	data, err := source.GetMarketData("BTCUSDT")
	if err != nil {
		t.Errorf("Expected no error when getting market data, got '%v'", err)
	}

	if len(data) == 0 {
		t.Error("Expected at least one market data record")
	}

	// 测试获取历史数据
	historicalData, err := source.GetHistoricalData("BTCUSDT", "2023-01-01T00:00:00Z", "2023-01-02T00:00:00Z")
	if err != nil {
		t.Errorf("Expected no error when getting historical data, got '%v'", err)
	}

	if len(historicalData) == 0 {
		t.Error("Expected at least one historical data record")
	}
}
