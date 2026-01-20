package main

import (
	"context"
	"os"
	"os/signal"
	"quant-data-engine/internal/api"
	"quant-data-engine/internal/config"
	"quant-data-engine/internal/datasource"
	"quant-data-engine/internal/kafka"
	"quant-data-engine/internal/schedule"
	"quant-data-engine/internal/storage"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// 设置日志级别为Debug
	logrus.SetLevel(logrus.DebugLevel)

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	logrus.Info("Starting Quant Data Engine...")

	// 创建根上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化存储
	db, err := storage.NewPostgresStorage()
	if err != nil {
		logrus.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.Close()

	// 初始化Kafka
	kafkaProducer, err := kafka.NewKafkaProducer()
	if err != nil {
		logrus.Fatalf("Failed to initialize Kafka: %v", err)
	}
	defer kafkaProducer.Close()

	// 初始化数据源
	dataSourceFactory := datasource.NewDataSourceFactory()
	dataSourceFactory.Register("binance", datasource.NewExchangeDataSource("binance", config.AppConfig.ExchangeAPIKey, config.AppConfig.ExchangeAPISecret))
	dataSourceFactory.Register("okx", datasource.NewExchangeDataSource("okx", config.AppConfig.ExchangeAPIKey, config.AppConfig.ExchangeAPISecret))

	// 初始化 Tushare 客户端
	tushareClient := datasource.NewTushareClient()

	// 初始化定时任务调度器
	scheduler := schedule.NewScheduler(tushareClient, db)

	// 启动定时任务
	scheduler.Start()

	// 初始化API服务器
	apiServer := api.NewServer(tushareClient, db)

	// 启动API服务器
	go func() {
		if err := apiServer.Run(config.AppConfig.APIPort); err != nil {
			logrus.Fatalf("Failed to start API server: %v", err)
		}
	}()

	// 启动数据获取和处理
	dataProcessingDone := make(chan struct{})
	go func() {
		defer close(dataProcessingDone)
		startDataProcessing(ctx, dataSourceFactory, db, kafkaProducer)
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down Quant Data Engine...")

	// 取消上下文，通知所有goroutine停止
	cancel()

	// 等待数据处理完成
	select {
	case <-dataProcessingDone:
		logrus.Info("Data processing stopped gracefully")
	case <-time.After(5 * time.Second):
		logrus.Warn("Data processing stopped forcefully after timeout")
	}

	logrus.Info("Quant Data Engine stopped")
}

// startDataProcessing 启动数据处理
func startDataProcessing(ctx context.Context, factory *datasource.DataSourceFactory, db *storage.PostgresStorage, kafkaProducer *kafka.KafkaProducer) {
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	interval := 30 * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logrus.Infof("Starting data processing with interval %v", interval)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Data processing context canceled, exiting...")
			return
		case <-ticker.C:
			processData(factory, db, kafkaProducer, symbols)
		}
	}
}

// processData 处理数据
func processData(factory *datasource.DataSourceFactory, db *storage.PostgresStorage, kafkaProducer *kafka.KafkaProducer, symbols []string) {
	logrus.Info("Processing market data...")

	for _, symbol := range symbols {
		// 从各个数据源获取数据
		for _, sourceName := range []string{"binance", "okx"} {
			source := factory.GetDataSource(sourceName)
			if source == nil {
				logrus.Warnf("DataSource %s not found", sourceName)
				continue
			}

			// 获取市场数据
			data, err := source.GetMarketData(symbol)
			if err != nil {
				logrus.Errorf("Failed to get market data from %s for %s: %v", sourceName, symbol, err)
				continue
			}

			if len(data) == 0 {
				logrus.Infof("No market data received from %s for %s", sourceName, symbol)
				continue
			}

			// 保存到数据库
			if err := db.SaveMarketData(data); err != nil {
				logrus.Errorf("Failed to save market data to database: %v", err)
				continue
			}

			// 发送到Kafka
			if err := kafkaProducer.SendMarketData(data); err != nil {
				logrus.Errorf("Failed to send market data to Kafka: %v", err)
				// 即使Kafka发送失败，也继续处理其他数据
				// 可以考虑添加重试机制或死信队列
			}
		}
	}

	logrus.Info("Market data processing completed")
}
