package kafka

import (
	"encoding/json"
	"fmt"
	"quant-data-engine/internal/config"
	"quant-data-engine/internal/models"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

// KafkaProducer Kafka生产者
type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

// NewKafkaProducer 创建Kafka生产者
func NewKafkaProducer() (*KafkaProducer, error) {
	cfg := config.AppConfig

	// 配置Kafka生产者
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
		"client.id":         "quant-data-engine",
		"acks":              "all",
		"retries":           3,
		"retry.backoff.ms":  1000,
		"linger.ms":         100,
		"batch.size":        16384,
		"compression.type":  "gzip",
	})
	if err != nil {
		logrus.Errorf("Failed to create Kafka producer: %v", err)
		return nil, err
	}

	// 启动消息发送结果处理
	go handleDeliveryReports(producer)

	logrus.Info("Connected to Kafka successfully")
	return &KafkaProducer{
		producer: producer,
		topic:    cfg.KafkaTopic,
	}, nil
}

// handleDeliveryReports 处理消息发送结果
func handleDeliveryReports(producer *kafka.Producer) {
	for e := range producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				logrus.Errorf("Delivery failed: %v", ev.TopicPartition.Error)
			} else {
				logrus.Debugf("Delivered message to %v", ev.TopicPartition)
			}
		}
	}
}

// SendMarketData 发送市场数据到Kafka
func (p *KafkaProducer) SendMarketData(data []models.MarketData) error {
	if len(data) == 0 {
		return nil
	}

	// 验证数据
	for i, d := range data {
		if err := validateMarketDataForKafka(d); err != nil {
			return fmt.Errorf("invalid market data at index %d: %w", i, err)
		}
	}

	// 用于跟踪发送失败的消息
	var failedMessages int

	for _, d := range data {
		// 将数据转换为JSON
		jsonData, err := json.Marshal(d)
		if err != nil {
			logrus.Errorf("Failed to marshal market data: %v", err)
			failedMessages++
			continue
		}

		// 创建消息
		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
			Value:          jsonData,
			Key:            []byte(d.Symbol),
			Headers: []kafka.Header{
				{Key: "source", Value: []byte(d.Source)},
				{Key: "timestamp", Value: []byte(d.Timestamp.Format(time.RFC3339))},
			},
		}

		// 发送消息
		if err := p.producer.Produce(message, nil); err != nil {
			logrus.Errorf("Failed to produce message: %v", err)
			failedMessages++
			continue
		}
	}

	// 等待所有消息发送完成
	remaining := p.producer.Flush(10 * 1000)
	if remaining > 0 {
		logrus.Warnf("Failed to send %d messages to Kafka", remaining)
		failedMessages += remaining
	}

	if failedMessages > 0 {
		return fmt.Errorf("failed to send %d out of %d messages to Kafka", failedMessages, len(data))
	}

	logrus.Infof("Sent %d market data messages to Kafka", len(data))
	return nil
}

// validateMarketDataForKafka 验证Kafka消息数据
func validateMarketDataForKafka(data models.MarketData) error {
	if data.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if data.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	if data.Source == "" {
		return fmt.Errorf("source is required")
	}
	return nil
}

// SendBacktestData 发送回测数据到Kafka
func (p *KafkaProducer) SendBacktestData(data models.BacktestData) error {
	// 将数据转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("Failed to marshal backtest data: %v", err)
		return fmt.Errorf("failed to marshal backtest data: %w", err)
	}

	// 创建消息
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          jsonData,
		Key:            []byte(data.Symbol),
		Headers: []kafka.Header{
			{Key: "type", Value: []byte("backtest")},
			{Key: "strategy", Value: []byte(data.Strategy)},
			{Key: "timestamp", Value: []byte(data.Timestamp.Format(time.RFC3339))},
		},
	}

	// 发送消息
	if err := p.producer.Produce(message, nil); err != nil {
		logrus.Errorf("Failed to produce backtest data message: %v", err)
		return fmt.Errorf("failed to produce backtest data message: %w", err)
	}

	// 等待消息发送完成
	p.producer.Flush(5 * 1000)

	logrus.Infof("Sent backtest data message for symbol %s to Kafka", data.Symbol)
	return nil
}

// Close 关闭Kafka生产者
func (p *KafkaProducer) Close() {
	if p.producer != nil {
		p.producer.Close()
		logrus.Info("Kafka producer closed")
	}
}
