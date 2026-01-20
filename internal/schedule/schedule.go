package schedule

import (
	"quant-data-engine/internal/datasource"
	"quant-data-engine/internal/models"
	"quant-data-engine/internal/storage"
	"time"

	"github.com/sirupsen/logrus"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	tushareClient *datasource.TushareClient
	storage       *storage.PostgresStorage
}

// NewScheduler 创建定时任务调度器
func NewScheduler(tushareClient *datasource.TushareClient, storage *storage.PostgresStorage) *Scheduler {
	return &Scheduler{
		tushareClient: tushareClient,
		storage:       storage,
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() {
	// 立即执行一次获取股票列表
	s.fetchStockList()

	// 每30分钟执行一次
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			s.fetchStockList()
		}
	}()

	logrus.Info("Scheduler started, fetching stock list every 30 minutes")
}

// fetchStockList 获取股票列表
func (s *Scheduler) fetchStockList() {
	logrus.Info("Starting to fetch stock list")

	// 调用 Tushare API 获取股票基础信息
	resp, err := s.tushareClient.GetStockBasic(&datasource.StockBasicRequest{
		ListStatus: "L", // 只获取上市的股票
	}, []string{
		"ts_code", "symbol", "name", "area", "industry", "fullname", "enname", "cnspell",
		"market", "exchange", "curr_type", "list_status", "list_date", "delist_date", "is_hs",
		"act_name", "act_ent_type",
	})

	if err != nil {
		logrus.Errorf("Failed to fetch stock list: %v", err)
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

	// 保存到数据库
	if len(stockList) > 0 {
		if err := s.storage.SaveStockBasic(stockList); err != nil {
			logrus.Errorf("Failed to save stock list: %v", err)
			return
		}
		logrus.Infof("Successfully saved %d stocks to database", len(stockList))
	}
}
