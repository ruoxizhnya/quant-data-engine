// Package api 提供API服务
// @title Quant Data Engine API
// @version 1.0
// @description 量化数据引擎API文档
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com
// @host localhost:8080
// @BasePath /
// @schemes http
package api

import (
	"fmt"
	"net/http"
	"quant-data-engine/internal/datasource"
	"quant-data-engine/internal/models"
	"quant-data-engine/internal/storage"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server API服务器
type Server struct {
	router        *gin.Engine
	mutex         sync.RWMutex
	tushareClient datasource.TushareClientInterface
	storage       storage.StorageInterface
}

// NewServer 创建API服务器
func NewServer(tushareClient datasource.TushareClientInterface, storage storage.StorageInterface) *Server {
	router := gin.Default()

	// 配置CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 添加Swagger UI路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/swagger.json"),
	))

	// 直接提供swagger.json文件
	router.GET("/swagger.json", func(c *gin.Context) {
		c.File("./docs/swagger.json")
	})

	server := &Server{
		router:        router,
		tushareClient: tushareClient,
		storage:       storage,
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	// 健康检查
	s.router.GET("/health", s.healthCheck)

	// 回测数据相关
	backtest := s.router.Group("/backtest")
	{
		backtest.GET("/data", s.getBacktestData)
		backtest.GET("/parquet", s.getParquetData)
	}

	// 市场数据相关
	market := s.router.Group("/market")
	{
		market.GET("/data", s.getMarketData)
	}

	// 股票数据相关
	stock := s.router.Group("/stock")
	{
		stock.POST("/fetch-list", s.fetchStockList)
	}
}

// healthCheck 健康检查
// @Summary 健康检查
// @Description 检查量化数据引擎API是否正常运行
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Router /health [get]
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Quant Data Engine API is running",
		Data: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"status":    "healthy",
		},
	})
}

// getBacktestData 获取回测数据
// @Summary 获取回测数据
// @Description 获取指定交易对的回测数据
// @Tags 回测
// @Accept json
// @Produce json
// @Param symbol query string true "交易对符号，例如 BTCUSDT"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /backtest/data [get]
func (s *Server) getBacktestData(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Symbol is required",
		})
		return
	}

	// 模拟回测数据
	data := models.BacktestData{
		ID:        fmt.Sprintf("backtest_%s_%d", symbol, time.Now().Unix()),
		Symbol:    symbol,
		Strategy:  "MA Cross",
		StartDate: time.Now().AddDate(-1, 0, 0),
		EndDate:   time.Now(),
		Results:   `{"profit": 12.5, "drawdown": 5.2, "trades": 120}`,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Backtest data retrieved successfully",
		Data:    data,
	})
}

// getParquetData 获取Parquet格式的回测数据
// @Summary 获取Parquet格式的回测数据
// @Description 获取指定交易对和日期范围的Parquet格式回测数据
// @Tags 回测
// @Accept json
// @Produce json
// @Param symbol query string true "交易对符号，例如 BTCUSDT"
// @Param start_date query string false "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string false "结束日期，格式：YYYY-MM-DD"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /backtest/parquet [get]
func (s *Server) getParquetData(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Symbol is required",
		})
		return
	}

	// 解析日期参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid start_date format, use YYYY-MM-DD",
			})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid end_date format, use YYYY-MM-DD",
			})
			return
		}
	} else {
		endDate = time.Now()
	}

	// 验证日期范围
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "end_date must be after start_date",
		})
		return
	}

	// 验证日期范围不超过1年
	if endDate.Sub(startDate) > 366*24*time.Hour {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Date range must not exceed 1 year",
		})
		return
	}

	// 生成模拟Parquet数据
	var data []map[string]interface{}
	current := startDate
	for current.Before(endDate) {
		data = append(data, map[string]interface{}{
			"timestamp": current.Unix(),
			"symbol":    symbol,
			"price":     1000 + float64(current.Day())*10,
			"volume":    10000 + float64(current.Hour())*1000,
		})
		current = current.Add(1 * time.Hour)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Parquet data retrieved successfully",
		Data:    data,
	})
}

// getMarketData 获取市场数据
// @Summary 获取市场数据
// @Description 获取指定交易对的市场数据
// @Tags 市场
// @Accept json
// @Produce json
// @Param symbol query string true "交易对符号，例如 BTCUSDT"
// @Param limit query int false "返回数据条数，默认10"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /market/data [get]
func (s *Server) getMarketData(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Symbol is required",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// 模拟市场数据
	var data []models.MarketData
	for i := 0; i < limit; i++ {
		data = append(data, models.MarketData{
			ID:        fmt.Sprintf("market_%s_%d", symbol, time.Now().Unix()-int64(i)),
			Symbol:    symbol,
			Price:     1000 + float64(i)*10,
			Volume:    10000 + float64(i)*1000,
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			Source:    "Binance",
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Market data retrieved successfully",
		Data:    data,
	})
}

// Run 运行API服务器
func (s *Server) Run(port string) error {
	logrus.Infof("Starting API server on port %s", port)
	return s.router.Run(":" + port)
}

// fetchStockList 手动触发获取股票列表
// @Summary 手动触发获取股票列表
// @Description 手动触发从 Tushare API 获取股票列表并保存到数据库中
// @Tags 股票
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /stock/fetch-list [post]
func (s *Server) fetchStockList(c *gin.Context) {
	logrus.Info("Manually triggering stock list fetch")

	// 准备请求参数
	req := &datasource.StockBasicRequest{
		ListStatus: "L", // 只获取上市的股票
	}
	fields := []string{
		"ts_code", "symbol", "name", "area", "industry", "fullname", "enname", "cnspell",
		"market", "exchange", "curr_type", "list_status", "list_date", "delist_date", "is_hs",
		"act_name", "act_ent_type",
	}

	// 输出请求详细信息
	logrus.Debugf("Requesting stock basic info with params: %+v", req)
	logrus.Debugf("Requesting fields: %+v", fields)

	// 调用 Tushare API 获取股票基础信息
	resp, err := s.tushareClient.GetStockBasic(req, fields)

	// 输出响应详细信息
	if resp != nil {
		logrus.Debugf("Tushare API response code: %d", resp.Code)
		logrus.Debugf("Tushare API response message: %s", resp.Message)
		if resp.Data != nil {
			logrus.Debugf("Tushare API response fields: %+v", resp.Data.Fields)
			logrus.Debugf("Tushare API response items count: %d", len(resp.Data.Items))
			if len(resp.Data.Items) > 0 {
				logrus.Debugf("First item: %+v", resp.Data.Items[0])
			}
		}
	}

	if err != nil {
		logrus.Errorf("Failed to fetch stock list: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: fmt.Sprintf("Failed to fetch stock list: %v", err),
		})
		return
	}

	// 解析响应数据
	var stockList []models.StockBasic
	if resp.Data != nil && len(resp.Data.Items) > 0 {
		for _, item := range resp.Data.Items {
			stock := models.StockBasic{}

			// 解析字段
			for i, field := range resp.Data.Fields {
				if i < len(item) {
					switch field {
					case "ts_code":
						if v, ok := item[i].(string); ok {
							stock.TSCode = v
						}
					case "symbol":
						if v, ok := item[i].(string); ok {
							stock.Symbol = v
						}
					case "name":
						if v, ok := item[i].(string); ok {
							stock.Name = v
						}
					case "area":
						if v, ok := item[i].(string); ok {
							stock.Area = v
						}
					case "industry":
						if v, ok := item[i].(string); ok {
							stock.Industry = v
						}
					case "fullname":
						if v, ok := item[i].(string); ok {
							stock.Fullname = v
						}
					case "enname":
						if v, ok := item[i].(string); ok {
							stock.Enname = v
						}
					case "cnspell":
						if v, ok := item[i].(string); ok {
							stock.Cnspell = v
						}
					case "market":
						if v, ok := item[i].(string); ok {
							stock.Market = v
						}
					case "exchange":
						if v, ok := item[i].(string); ok {
							stock.Exchange = v
						}
					case "curr_type":
						if v, ok := item[i].(string); ok {
							stock.CurrType = v
						}
					case "list_status":
						if v, ok := item[i].(string); ok {
							stock.ListStatus = v
						}
					case "list_date":
						if v, ok := item[i].(string); ok {
							stock.ListDate = v
						}
					case "delist_date":
						if v, ok := item[i].(string); ok {
							stock.DelistDate = v
						}
					case "is_hs":
						if v, ok := item[i].(string); ok {
							stock.IsHS = v
						}
					case "act_name":
						if v, ok := item[i].(string); ok {
							stock.ActName = v
						}
					case "act_ent_type":
						if v, ok := item[i].(string); ok {
							stock.ActEntType = v
						}
					}
				}
			}

			// 添加到列表
			stockList = append(stockList, stock)
		}
	}

	logrus.Infof("Fetched %d stocks from Tushare API", len(stockList))
	if len(stockList) > 0 {
		logrus.Debugf("First 5 stocks: %+v", stockList[:min(5, len(stockList))])
	}

	// 保存到数据库
	if len(stockList) > 0 {
		logrus.Debugf("Saving %d stocks to database", len(stockList))
		if err := s.storage.SaveStockBasic(stockList); err != nil {
			logrus.Errorf("Failed to save stock list: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: fmt.Sprintf("Failed to save stock list: %v", err),
			})
			return
		}
		logrus.Infof("Successfully saved %d stocks to database", len(stockList))
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Stock list fetched and saved successfully",
		Data: map[string]interface{}{
			"count": len(stockList),
		},
	})
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
