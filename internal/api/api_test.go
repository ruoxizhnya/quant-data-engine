package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestServer_HealthCheck  测试健康检查接口
func TestServer_HealthCheck(t *testing.T) {
	// 创建API服务器
	server := NewServer()

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
	// 创建API服务器
	server := NewServer()

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
	// 创建API服务器
	server := NewServer()

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
	// 创建API服务器
	server := NewServer()

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
