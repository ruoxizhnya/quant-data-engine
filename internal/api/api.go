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

	// 同步相关
	sync := s.router.Group("/sync")
	{
		sync.POST("/ohlcv/full", s.syncOHLCVFull)
		sync.POST("/trade-calendar", s.syncTradeCalendar)
		sync.GET("/ohlcv/status", s.getOHLCVStatus)
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

// SyncOHLCVFullRequest 全量OHLCV同步请求
type SyncOHLCVFullRequest struct {
	Symbols   []string `json:"symbols"`   // 指定股票列表，为空则同步所有
	StartYear int      `json:"start_year"` // 起始年份，默认2000
	EndYear   int      `json:"end_year"`   // 结束年份，默认当前年份
}

// syncOHLCVFull 全量OHLCV前复权数据同步
// @Summary 全量OHLCV前复权数据同步
// @Description 按年分段同步A股日线前复权(OHLCV QFQ)数据，支持指定股票列表或全量同步
// @Tags 同步
// @Accept json
// @Produce json
// @Param request body SyncOHLCVFullRequest true "同步参数"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /sync/ohlcv/full [post]
func (s *Server) syncOHLCVFull(c *gin.Context) {
	var req SyncOHLCVFullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request: " + err.Error()})
		return
	}

	// 默认参数
	if req.StartYear == 0 {
		req.StartYear = 2000
	}
	if req.EndYear == 0 {
		req.EndYear = time.Now().Year()
	}

	// 获取股票列表
	var tsCodes []string
	if len(req.Symbols) > 0 {
		tsCodes = req.Symbols
	} else {
		codes, err := s.storage.GetAllStockCodes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get stock codes: " + err.Error()})
			return
		}
		tsCodes = codes
	}

	logrus.Infof("Starting OHLCV QFQ sync for %d stocks, years %d-%d", len(tsCodes), req.StartYear, req.EndYear)

	// 同步字段
	fields := []string{"ts_code", "trade_date", "open", "high", "low", "close", "vol", "amount"}

	totalSynced := 0
	totalSkipped := 0
	totalErrors := 0

	// 遍历每只股票
	for _, tsCode := range tsCodes {
		logrus.Infof("Syncing %s", tsCode)

		// 检查已存在的日期范围
		existingMin, existingMax, _ := s.storage.GetExistingDateRangeForSymbol(tsCode)
		if existingMin != "" && existingMax != "" {
			logrus.Debugf("%s already has data from %s to %s, will fill gaps", tsCode, existingMin, existingMax)
		}

		// 按年分段同步
		for year := req.StartYear; year <= req.EndYear; year++ {
			startDate := fmt.Sprintf("%d0101", year)
			endDate := fmt.Sprintf("%d1231", year)

			// 获取日线数据（未复权）
			dailyResp, err := s.tushareClient.GetDaily(&datasource.DailyRequest{
				TSCode:    tsCode,
				StartDate: startDate,
				EndDate:   endDate,
			}, fields)
			if err != nil {
				logrus.Errorf("Failed to fetch daily for %s year %d: %v", tsCode, year, err)
				totalErrors++
				continue
			}
			if dailyResp == nil || dailyResp.Data == nil || len(dailyResp.Data.Items) == 0 {
				logrus.Debugf("No daily data for %s year %d", tsCode, year)
				totalSkipped++
				time.Sleep(250 * time.Millisecond)
				continue
			}

			// 获取复权因子
			adjResp, err := s.tushareClient.GetAdjFactor(&datasource.AdjFactorRequest{
				TSCode:    tsCode,
				StartDate: startDate,
				EndDate:   endDate,
			}, []string{"ts_code", "trade_date", "adj_factor"})
			if err != nil {
				logrus.Errorf("Failed to fetch adj_factor for %s year %d: %v", tsCode, year, err)
				totalErrors++
				time.Sleep(250 * time.Millisecond)
				continue
			}

			// 构建复权因子映射
			adjMap := make(map[string]float64)
			var lastAdjFactor float64 = 1.0
			if adjResp != nil && adjResp.Data != nil {
				for _, item := range adjResp.Data.Items {
					tradeDate := ""
					var adjFactor float64 = 1.0
					for i, field := range adjResp.Data.Fields {
						if i < len(item) {
							switch field {
							case "trade_date":
								if v, ok := item[i].(string); ok {
									tradeDate = v
								}
							case "adj_factor":
								if v, ok := item[i].(float64); ok {
									adjFactor = v
								} else if v, ok := item[i].(int); ok {
									adjFactor = float64(v)
								}
							}
						}
					}
					if tradeDate != "" {
						adjMap[tradeDate] = adjFactor
						lastAdjFactor = adjFactor
					}
				}
			}

			// 解析并应用复权
			var ohlcvList []models.OHLCVDailyQFQ
			for _, item := range dailyResp.Data.Items {
				ohlcv := models.OHLCVDailyQFQ{}
				var tradeDate string
				var unadjClose float64

				for i, field := range dailyResp.Data.Fields {
					if i < len(item) {
						switch field {
						case "ts_code":
							if v, ok := item[i].(string); ok {
								ohlcv.Symbol = v
							}
						case "trade_date":
							if v, ok := item[i].(string); ok {
								tradeDate = v
								ohlcv.TradeDate = v
							}
						case "open":
							if v, ok := item[i].(float64); ok {
								ohlcv.Open = v
							} else if v, ok := item[i].(int); ok {
								ohlcv.Open = float64(v)
							}
						case "high":
							if v, ok := item[i].(float64); ok {
								ohlcv.High = v
							} else if v, ok := item[i].(int); ok {
								ohlcv.High = float64(v)
							}
						case "low":
							if v, ok := item[i].(float64); ok {
								ohlcv.Low = v
							} else if v, ok := item[i].(int); ok {
								ohlcv.Low = float64(v)
							}
						case "close":
							if v, ok := item[i].(float64); ok {
								unadjClose = v
								ohlcv.Close = v
							} else if v, ok := item[i].(int); ok {
								unadjClose = float64(v)
								ohlcv.Close = float64(v)
							}
						case "vol":
							if v, ok := item[i].(float64); ok {
								ohlcv.Volume = v
							} else if v, ok := item[i].(int); ok {
								ohlcv.Volume = float64(v)
							}
						case "amount":
							if v, ok := item[i].(float64); ok {
								ohlcv.Turnover = v
							} else if v, ok := item[i].(int); ok {
								ohlcv.Turnover = float64(v)
							}
						}
					}
				}

				// 应用前复权调整因子
				if tradeDate != "" && ohlcv.Symbol != "" {
					adjFactor := adjMap[tradeDate]
					if adjFactor == 0 {
						adjFactor = lastAdjFactor
					}
					if adjFactor > 0 && adjFactor != 1.0 {
						ohlcv.Open = ohlcv.Open * adjFactor
						ohlcv.High = ohlcv.High * adjFactor
						ohlcv.Low = ohlcv.Low * adjFactor
						ohlcv.Close = unadjClose * adjFactor
					}
					ohlcvList = append(ohlcvList, ohlcv)
				}
			}

			if len(ohlcvList) > 0 {
				if err := s.storage.SaveOHLCVDailyQFQ(ohlcvList); err != nil {
					logrus.Errorf("Failed to save OHLCV for %s year %d: %v", tsCode, year, err)
					totalErrors++
				} else {
					totalSynced += len(ohlcvList)
					logrus.Infof("Saved %d QFQ records for %s year %d", len(ohlcvList), tsCode, year)
				}
			} else {
				totalSkipped++
			}

			// 速率限制：250ms
			time.Sleep(250 * time.Millisecond)
		}
	}

	logrus.Infof("OHLCV sync completed: %d records synced, %d years skipped, %d errors",
		totalSynced, totalSkipped, totalErrors)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "OHLCV QFQ sync completed",
		Data: map[string]interface{}{
			"total_synced":  totalSynced,
			"total_skipped": totalSkipped,
			"total_errors":  totalErrors,
			"stocks_count":  len(tsCodes),
			"start_year":    req.StartYear,
			"end_year":      req.EndYear,
		},
	})
}

// syncTradeCalendar 同步交易日历
// @Summary 同步交易日历
// @Description 从Tushare同步交易日历到trade_calendar表
// @Tags 同步
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /sync/trade-calendar [post]
func (s *Server) syncTradeCalendar(c *gin.Context) {
	logrus.Info("Starting trade calendar sync")

	// Tushare trade_cal API字段
	fields := []string{"exchange", "cal_date", "is_open", "pre_trade_date"}

	// 同步最近20年的数据
	endDate := time.Now().Format("20060102")
	startDate := fmt.Sprintf("%d0101", time.Now().Year()-20)

	req := &datasource.TradeCalRequest{
		Exchange:  "",
		StartDate: startDate,
		EndDate:   endDate,
		IsOpen:    "",
	}

	resp, err := s.tushareClient.GetTradeCal(req, fields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch trade cal: " + err.Error()})
		return
	}

	if resp == nil || resp.Data == nil || len(resp.Data.Items) == 0 {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "No trade calendar data returned",
			Data:    map[string]interface{}{"count": 0},
		})
		return
	}

	// 解析并保存
	var tradeCalList []models.TradeCal
	for _, item := range resp.Data.Items {
		tc := models.TradeCal{}
		for i, field := range resp.Data.Fields {
			if i < len(item) {
				switch field {
				case "exchange":
					if v, ok := item[i].(string); ok {
						tc.Exchange = v
					}
				case "cal_date":
					if v, ok := item[i].(string); ok {
						tc.CalDate = v
					}
				case "is_open":
					if v, ok := item[i].(string); ok {
						tc.IsOpen = v
					}
				case "pre_trade_date":
					if v, ok := item[i].(string); ok {
						tc.PreTradeDate = v
					}
				}
			}
		}
		if tc.CalDate != "" {
			tradeCalList = append(tradeCalList, tc)
		}
	}

	if err := s.storage.SaveTradeCalendar(tradeCalList); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save trade calendar: " + err.Error()})
		return
	}

	logrus.Infof("Saved %d trade calendar records", len(tradeCalList))

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trade calendar synced successfully",
		Data: map[string]interface{}{
			"count": len(tradeCalList),
		},
	})
}

// getOHLCVStatus 获取OHLCV同步状态
// @Summary 获取OHLCV同步状态
// @Description 查看各股票的OHLCV数据同步情况
// @Tags 同步
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /sync/ohlcv/status [get]
func (s *Server) getOHLCVStatus(c *gin.Context) {
	tsCode := c.Query("ts_code")

	if tsCode != "" {
		count, _ := s.storage.GetOHLCVCountBySymbol(tsCode)
		minDate, maxDate, _ := s.storage.GetExistingDateRangeForSymbol(tsCode)
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "OHLCV status retrieved",
			Data: map[string]interface{}{
				"ts_code":   tsCode,
				"count":     count,
				"min_date":  minDate,
				"max_date":  maxDate,
			},
		})
		return
	}

	// 返回总体统计
	codes, err := s.storage.GetAllStockCodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	stats := make([]map[string]interface{}, 0, min(100, len(codes)))
	for _, code := range codes {
		if len(stats) >= 100 {
			break
		}
		count, _ := s.storage.GetOHLCVCountBySymbol(code)
		minDate, maxDate, _ := s.storage.GetExistingDateRangeForSymbol(code)
		stats = append(stats, map[string]interface{}{
			"ts_code":  code,
			"count":    count,
			"min_date": minDate,
			"max_date": maxDate,
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "OHLCV status retrieved",
		Data: map[string]interface{}{
			"total_stocks": len(codes),
			"sample":       stats,
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
