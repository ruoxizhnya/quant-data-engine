package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"quant-data-engine/internal/datasource"
	"quant-data-engine/internal/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockTushareClient 模拟 Tushare 客户端
type MockTushareClient struct {
	GetStockBasicFunc func(req *datasource.StockBasicRequest, fields []string) (*datasource.TushareResponse, error)
}

// GetStockBasic 模拟获取股票基础信息
func (m *MockTushareClient) GetStockBasic(req *datasource.StockBasicRequest, fields []string) (*datasource.TushareResponse, error) {
	if m.GetStockBasicFunc != nil {
		return m.GetStockBasicFunc(req, fields)
	}
	return nil, nil
}

// GetTradeCal 模拟获取交易日历
func (m *MockTushareClient) GetTradeCal(req *datasource.TradeCalRequest, fields []string) (*datasource.TushareResponse, error) {
	return nil, nil
}

// GetNewShare 模拟获取新股上市列表
func (m *MockTushareClient) GetNewShare(req *datasource.NewShareRequest, fields []string) (*datasource.TushareResponse, error) {
	return nil, nil
}

// GetStockCompany 模拟获取上市公司基础信息
func (m *MockTushareClient) GetStockCompany(req *datasource.StockCompanyRequest, fields []string) (*datasource.TushareResponse, error) {
	return nil, nil
}

// GetStkManagers 模拟获取上市公司管理层
func (m *MockTushareClient) GetStkManagers(req *datasource.StkManagersRequest, fields []string) (*datasource.TushareResponse, error) {
	return nil, nil
}

// GetStkRewards 模拟获取管理层薪酬和持股
func (m *MockTushareClient) GetStkRewards(req *datasource.StkRewardsRequest, fields []string) (*datasource.TushareResponse, error) {
	return nil, nil
}

// GetDaily 模拟获取A股日线行情
func (m *MockTushareClient) GetDaily(req *datasource.DailyRequest, fields []string) (*datasource.TushareResponse, error) {
	return nil, nil
}

// MockStorage 模拟存储实例
type MockStorage struct {
	SaveStockBasicFunc func(data []models.StockBasic) error
	GetStockBasicFunc  func(limit int) ([]models.StockBasic, error)
}

// SaveStockBasic 模拟保存股票基础信息
func (m *MockStorage) SaveStockBasic(data []models.StockBasic) error {
	if m.SaveStockBasicFunc != nil {
		return m.SaveStockBasicFunc(data)
	}
	return nil
}

// GetStockBasic 模拟获取股票基础信息
func (m *MockStorage) GetStockBasic(limit int) ([]models.StockBasic, error) {
	if m.GetStockBasicFunc != nil {
		return m.GetStockBasicFunc(limit)
	}
	return nil, nil
}

// SaveMarketData 模拟保存市场数据
func (m *MockStorage) SaveMarketData(data []models.MarketData) error {
	return nil
}

// SaveBacktestData 模拟保存回测数据
func (m *MockStorage) SaveBacktestData(data models.BacktestData) error {
	return nil
}

// GetMarketData 模拟获取市场数据
func (m *MockStorage) GetMarketData(symbol string, limit int) ([]models.MarketData, error) {
	return nil, nil
}

// GetHistoricalData 模拟获取历史数据
func (m *MockStorage) GetHistoricalData(symbol string, startTime, endTime string) ([]models.MarketData, error) {
	return nil, nil
}

// Close 模拟关闭存储
func (m *MockStorage) Close() {
}

// TestServer_HealthCheck  测试健康检查接口
func TestServer_HealthCheck(t *testing.T) {
	// 创建模拟的 Tushare 客户端和存储实例
	mockTushareClient := &MockTushareClient{}
	mockStorage := &MockStorage{}

	// 创建API服务器
	server := NewServer(mockTushareClient, mockStorage)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/health", nil)

	// 调用健康检查方法
	server.healthCheck(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Quant Data Engine API is running")
}

// TestServer_GetBacktestData 测试获取回测数据接口
func TestServer_GetBacktestData(t *testing.T) {
	// 创建模拟的 Tushare 客户端和存储实例
	mockTushareClient := &MockTushareClient{}
	mockStorage := &MockStorage{}

	// 创建API服务器
	server := NewServer(mockTushareClient, mockStorage)

	// 测试缺少symbol参数
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/backtest/data", nil)
	server.getBacktestData(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Symbol is required")

	// 测试有symbol参数
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/backtest/data?symbol=BTCUSDT", nil)
	server.getBacktestData(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Backtest data retrieved successfully")
}

// TestServer_GetMarketData 测试获取市场数据接口
func TestServer_GetMarketData(t *testing.T) {
	// 创建模拟的 Tushare 客户端和存储实例
	mockTushareClient := &MockTushareClient{}
	mockStorage := &MockStorage{}

	// 创建API服务器
	server := NewServer(mockTushareClient, mockStorage)

	// 测试缺少symbol参数
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/market/data", nil)
	server.getMarketData(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Symbol is required")

	// 测试有symbol参数
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/market/data?symbol=BTCUSDT", nil)
	server.getMarketData(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Market data retrieved successfully")

	// 测试有symbol和limit参数
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/market/data?symbol=BTCUSDT&limit=5", nil)
	server.getMarketData(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Market data retrieved successfully")
}

// TestServer_GetParquetData 测试获取Parquet数据接口
func TestServer_GetParquetData(t *testing.T) {
	// 创建模拟的 Tushare 客户端和存储实例
	mockTushareClient := &MockTushareClient{}
	mockStorage := &MockStorage{}

	// 创建API服务器
	server := NewServer(mockTushareClient, mockStorage)

	// 测试缺少symbol参数
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/backtest/parquet", nil)
	server.getParquetData(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Symbol is required")

	// 测试日期范围无效
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/backtest/parquet?symbol=BTCUSDT&start_date=2023-01-01&end_date=2022-01-01", nil)
	server.getParquetData(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "end_date must be after start_date")
}

// TestServer_FetchStockList 测试获取股票列表接口
func TestServer_FetchStockList(t *testing.T) {
	// 创建模拟的 Tushare 客户端和存储实例
	mockTushareClient := &MockTushareClient{
		GetStockBasicFunc: func(req *datasource.StockBasicRequest, fields []string) (*datasource.TushareResponse, error) {
			// 返回模拟的响应数据
			return &datasource.TushareResponse{
				Code:    0,
				Message: "success",
				Data: &datasource.DataResult{
					Fields: []string{"ts_code", "symbol", "name", "area", "industry", "fullname", "enname", "cnspell", "market", "exchange", "curr_type", "list_status", "list_date", "delist_date", "is_hs", "act_name", "act_ent_type"},
					Items: [][]interface{}{
						{"600000.SH", "600000", "浦发银行", "上海", "银行", "上海浦东发展银行股份有限公司", "Shanghai Pudong Development Bank Co., Ltd.", "pfyh", "SSE", "SSE", "CNY", "L", "19990114", "", "H", "上海市国有资产监督管理委员会", "地方国有企业"},
						{"600001.SH", "600001", "邯郸钢铁", "河北", "钢铁", "河北钢铁股份有限公司", "Hebei Iron & Steel Co., Ltd.", "hdgt", "SSE", "SSE", "CNY", "L", "19961119", "", "", "河北省国有资产监督管理委员会", "地方国有企业"},
					},
				},
			}, nil
		},
	}

	mockStorage := &MockStorage{
		SaveStockBasicFunc: func(data []models.StockBasic) error {
			// 验证数据是否正确
			assert.Len(t, data, 2)
			assert.Equal(t, "600000.SH", data[0].TSCode)
			assert.Equal(t, "浦发银行", data[0].Name)
			assert.Equal(t, "600001.SH", data[1].TSCode)
			assert.Equal(t, "邯郸钢铁", data[1].Name)
			return nil
		},
	}

	// 创建API服务器
	server := NewServer(mockTushareClient, mockStorage)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/stock/fetch-list", nil)

	// 调用获取股票列表方法
	server.fetchStockList(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Stock list fetched and saved successfully")
	assert.Contains(t, w.Body.String(), "count")
}

// TestServer_FetchStockList_Error 测试获取股票列表接口失败情况
func TestServer_FetchStockList_Error(t *testing.T) {
	// 创建模拟的 Tushare 客户端和存储实例
	mockTushareClient := &MockTushareClient{
		GetStockBasicFunc: func(req *datasource.StockBasicRequest, fields []string) (*datasource.TushareResponse, error) {
			// 返回错误
			return nil, fmt.Errorf("API error")
		},
	}

	mockStorage := &MockStorage{}

	// 创建API服务器
	server := NewServer(mockTushareClient, mockStorage)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/stock/fetch-list", nil)

	// 调用获取股票列表方法
	server.fetchStockList(c)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "API error")
}

// TestServer_FetchStockList_StorageError 测试获取股票列表接口存储失败情况
func TestServer_FetchStockList_StorageError(t *testing.T) {
	// 创建模拟的 Tushare 客户端和存储实例
	mockTushareClient := &MockTushareClient{
		GetStockBasicFunc: func(req *datasource.StockBasicRequest, fields []string) (*datasource.TushareResponse, error) {
			// 返回模拟的响应数据
			return &datasource.TushareResponse{
				Code:    0,
				Message: "success",
				Data: &datasource.DataResult{
					Fields: []string{"ts_code", "symbol", "name"},
					Items: [][]interface{}{
						{"600000.SH", "600000", "浦发银行"},
					},
				},
			}, nil
		},
	}

	mockStorage := &MockStorage{
		SaveStockBasicFunc: func(data []models.StockBasic) error {
			// 返回错误
			return fmt.Errorf("Storage error")
		},
	}

	// 创建API服务器
	server := NewServer(mockTushareClient, mockStorage)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/stock/fetch-list", nil)

	// 调用获取股票列表方法
	server.fetchStockList(c)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Storage error")
}
