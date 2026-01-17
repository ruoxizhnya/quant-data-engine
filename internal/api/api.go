package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"quant-data-engine/internal/models"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parquet-go/parquet-go"
	"github.com/sirupsen/logrus"
)

// Server API服务器
type Server struct {
	router *gin.Engine
	mutex  sync.RWMutex
}

// NewServer 创建API服务器
func NewServer() *Server {
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

	server := &Server{
		router: router,
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	api := s.router.Group("/api")
	{
		// 健康检查
		api.GET("/health", s.healthCheck)

		// 回测数据相关
		backtest := api.Group("/backtest")
		{
			backtest.GET("/data", s.getBacktestData)
			backtest.GET("/parquet", s.getParquetData)
		}

		// 市场数据相关
		market := api.Group("/market")
		{
			market.GET("/data", s.getMarketData)
		}
	}
}

// healthCheck 健康检查
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

	// 生成Parquet文件路径
	parquetDir := "./data/parquet"
	if err := os.MkdirAll(parquetDir, 0755); err != nil {
		logrus.Errorf("Failed to create parquet directory: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create parquet directory",
		})
		return
	}

	parquetFile := filepath.Join(parquetDir, fmt.Sprintf("%s_%s_%s.parquet", symbol, startDate.Format("20060102"), endDate.Format("20060102")))

	// 检查文件是否存在，如果不存在则生成
	if _, err := os.Stat(parquetFile); os.IsNotExist(err) {
		logrus.Infof("Generating parquet file for %s from %s to %s", symbol, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err := s.generateParquetFile(parquetFile, symbol, startDate, endDate); err != nil {
			logrus.Errorf("Failed to generate parquet file: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to generate parquet file",
			})
			return
		}
	}

	// 读取Parquet文件并返回数据
	data, err := s.readParquetFile(parquetFile)
	if err != nil {
		logrus.Errorf("Failed to read parquet file: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to read parquet file",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Parquet data retrieved successfully",
		Data:    data,
	})
}

// getMarketData 获取市场数据
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

// generateParquetFile 生成Parquet文件
func (s *Server) generateParquetFile(filePath, symbol string, startDate, endDate time.Time) error {
	// 定义Parquet schema
	type ParquetData struct {
		Timestamp int64   `parquet:"name=timestamp, type=INT64"`
		Symbol    string  `parquet:"name=symbol, type=UTF8"`
		Price     float64 `parquet:"name=price, type=DOUBLE"`
		Volume    float64 `parquet:"name=volume, type=DOUBLE"`
	}

	// 创建Parquet文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}
	defer file.Close()

	// 创建Parquet writer
	writer := parquet.NewWriter(file, new(ParquetData))
	defer writer.Close()

	// 生成模拟数据
	current := startDate
	for current.Before(endDate) {
		data := ParquetData{
			Timestamp: current.Unix(),
			Symbol:    symbol,
			Price:     1000 + float64(current.Day())*10,
			Volume:    10000 + float64(current.Hour())*1000,
		}

		if err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to write to parquet file: %w", err)
		}

		current = current.Add(1 * time.Hour)
	}

	return nil
}

// readParquetFile 读取Parquet文件
func (s *Server) readParquetFile(filePath string) ([]map[string]interface{}, error) {
	// 打开Parquet文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open parquet file: %w", err)
	}
	defer file.Close()

	// 创建Parquet reader
	reader := parquet.NewReader(file)
	defer reader.Close()

	var result []map[string]interface{}

	// 读取数据
	for {
		row := make(map[string]interface{})
		if err := reader.Read(row); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to read parquet row: %w", err)
		}
		result = append(result, row)
	}

	return result, nil
}

// Run 运行API服务器
func (s *Server) Run(port string) error {
	logrus.Infof("Starting API server on port %s", port)
	return s.router.Run(":" + port)
}