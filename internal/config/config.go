package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBMaxConns int

	// Kafka配置
	KafkaBrokers string
	KafkaTopic   string
	KafkaRetries int

	// API配置
	APIPort string
	APITimeout int

	// 数据源配置
	ExchangeAPIKey    string
	ExchangeAPISecret string
	DataSourceTimeout int

	// 数据处理配置
	ProcessingInterval int
	MaxSymbols         int

	// 日志配置
	LogLevel string
}

var AppConfig *Config

func LoadConfig() error {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		// 数据库配置
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "quant_data"),
		DBMaxConns: getEnvAsInt("DB_MAX_CONNS", 10),

		// Kafka配置
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "quant_data"),
		KafkaRetries: getEnvAsInt("KAFKA_RETRIES", 3),

		// API配置
		APIPort: getEnv("API_PORT", "8080"),
		APITimeout: getEnvAsInt("API_TIMEOUT", 30),

		// 数据源配置
		ExchangeAPIKey:    getEnv("EXCHANGE_API_KEY", ""),
		ExchangeAPISecret: getEnv("EXCHANGE_API_SECRET", ""),
		DataSourceTimeout: getEnvAsInt("DATA_SOURCE_TIMEOUT", 10),

		// 数据处理配置
		ProcessingInterval: getEnvAsInt("PROCESSING_INTERVAL", 30),
		MaxSymbols:         getEnvAsInt("MAX_SYMBOLS", 10),

		// 日志配置
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(AppConfig.LogLevel)
	if err != nil {
		logrus.Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}