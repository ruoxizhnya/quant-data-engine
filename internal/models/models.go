package models

import (
	"time"
)

// 市场数据模型
type MarketData struct {
	ID        string    `json:"id" db:"id"`
	Symbol    string    `json:"symbol" db:"symbol"`
	Price     float64   `json:"price" db:"price"`
	Volume    float64   `json:"volume" db:"volume"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Source    string    `json:"source" db:"source"`
}

// 回测数据模型
type BacktestData struct {
	ID        string    `json:"id" db:"id"`
	Symbol    string    `json:"symbol" db:"symbol"`
	Strategy  string    `json:"strategy" db:"strategy"`
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`
	Results   string    `json:"results" db:"results"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// API响应模型
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 错误响应模型
type ErrorResponse struct {
	Error string `json:"error"`
}

// Parquet数据请求模型
type ParquetDataRequest struct {
	Symbol    string    `form:"symbol" binding:"required"`
	StartDate time.Time `form:"start_date" binding:"required" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" binding:"required" time_format:"2006-01-02"`
}