package datasource

import (
	"quant-data-engine/internal/models"
	"sync"
)

// DataSource 数据源接口
type DataSource interface {
	// GetMarketData 获取市场数据
	GetMarketData(symbol string) ([]models.MarketData, error)

	// GetHistoricalData 获取历史数据
	GetHistoricalData(symbol string, startTime, endTime string) ([]models.MarketData, error)

	// Name 获取数据源名称
	Name() string
}

// DataSourceFactory 数据源工厂
type DataSourceFactory struct {
	sources map[string]DataSource
	mutex   sync.RWMutex
}

// NewDataSourceFactory 创建数据源工厂
func NewDataSourceFactory() *DataSourceFactory {
	return &DataSourceFactory{
		sources: make(map[string]DataSource),
	}
}

// Register 注册数据源
func (f *DataSourceFactory) Register(name string, source DataSource) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.sources[name] = source
}

// GetDataSource 获取数据源
func (f *DataSourceFactory) GetDataSource(name string) DataSource {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.sources[name]
}

// ListDataSources 列出所有数据源
func (f *DataSourceFactory) ListDataSources() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	var names []string
	for name := range f.sources {
		names = append(names, name)
	}
	return names
}