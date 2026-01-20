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

// 股票基础信息模型
type StockBasic struct {
	TSCode     string    `json:"ts_code" db:"ts_code"`
	Symbol     string    `json:"symbol" db:"symbol"`
	Name       string    `json:"name" db:"name"`
	Area       string    `json:"area" db:"area"`
	Industry   string    `json:"industry" db:"industry"`
	Fullname   string    `json:"fullname" db:"fullname"`
	Enname     string    `json:"enname" db:"enname"`
	Cnspell    string    `json:"cnspell" db:"cnspell"`
	Market     string    `json:"market" db:"market"`
	Exchange   string    `json:"exchange" db:"exchange"`
	CurrType   string    `json:"curr_type" db:"curr_type"`
	ListStatus string    `json:"list_status" db:"list_status"`
	ListDate   string    `json:"list_date" db:"list_date"`
	DelistDate string    `json:"delist_date" db:"delist_date"`
	IsHS       string    `json:"is_hs" db:"is_hs"`
	ActName    string    `json:"act_name" db:"act_name"`
	ActEntType string    `json:"act_ent_type" db:"act_ent_type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// 交易日历模型
type TradeCal struct {
	Exchange     string    `json:"exchange" db:"exchange"`
	CalDate      string    `json:"cal_date" db:"cal_date"`
	IsOpen       string    `json:"is_open" db:"is_open"`
	PreTradeDate string    `json:"pre_trade_date" db:"pre_trade_date"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// 新股上市列表模型
type NewShare struct {
	TSCode       string    `json:"ts_code" db:"ts_code"`
	SubCode      string    `json:"sub_code" db:"sub_code"`
	Name         string    `json:"name" db:"name"`
	IPODate      string    `json:"ipo_date" db:"ipo_date"`
	IssueDate    string    `json:"issue_date" db:"issue_date"`
	Amount       float64   `json:"amount" db:"amount"`
	MarketAmount float64   `json:"market_amount" db:"market_amount"`
	Price        float64   `json:"price" db:"price"`
	PE           float64   `json:"pe" db:"pe"`
	LimitAmount  float64   `json:"limit_amount" db:"limit_amount"`
	Funds        float64   `json:"funds" db:"funds"`
	Ballot       float64   `json:"ballot" db:"ballot"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// 上市公司基础信息模型
type StockCompany struct {
	TSCode        string    `json:"ts_code" db:"ts_code"`
	ComName       string    `json:"com_name" db:"com_name"`
	ComID         string    `json:"com_id" db:"com_id"`
	Exchange      string    `json:"exchange" db:"exchange"`
	Chairman      string    `json:"chairman" db:"chairman"`
	Manager       string    `json:"manager" db:"manager"`
	Secretary     string    `json:"secretary" db:"secretary"`
	RegCapital    float64   `json:"reg_capital" db:"reg_capital"`
	SetupDate     string    `json:"setup_date" db:"setup_date"`
	Province      string    `json:"province" db:"province"`
	City          string    `json:"city" db:"city"`
	Introduction  string    `json:"introduction" db:"introduction"`
	Website       string    `json:"website" db:"website"`
	Email         string    `json:"email" db:"email"`
	Office        string    `json:"office" db:"office"`
	Employees     int       `json:"employees" db:"employees"`
	MainBusiness  string    `json:"main_business" db:"main_business"`
	BusinessScope string    `json:"business_scope" db:"business_scope"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// 上市公司管理层模型
type StkManagers struct {
	TSCode    string    `json:"ts_code" db:"ts_code"`
	AnnDate   string    `json:"ann_date" db:"ann_date"`
	Name      string    `json:"name" db:"name"`
	Gender    string    `json:"gender" db:"gender"`
	Lev       string    `json:"lev" db:"lev"`
	Title     string    `json:"title" db:"title"`
	Edu       string    `json:"edu" db:"edu"`
	National  string    `json:"national" db:"national"`
	Birthday  string    `json:"birthday" db:"birthday"`
	BeginDate string    `json:"begin_date" db:"begin_date"`
	EndDate   string    `json:"end_date" db:"end_date"`
	Resume    string    `json:"resume" db:"resume"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// 管理层薪酬和持股模型
type StkRewards struct {
	TSCode    string    `json:"ts_code" db:"ts_code"`
	AnnDate   string    `json:"ann_date" db:"ann_date"`
	EndDate   string    `json:"end_date" db:"end_date"`
	Name      string    `json:"name" db:"name"`
	Title     string    `json:"title" db:"title"`
	Reward    float64   `json:"reward" db:"reward"`
	HoldVol   float64   `json:"hold_vol" db:"hold_vol"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// A股日线行情模型
type Daily struct {
	TSCode    string    `json:"ts_code" db:"ts_code"`
	TradeDate string    `json:"trade_date" db:"trade_date"`
	Open      float64   `json:"open" db:"open"`
	High      float64   `json:"high" db:"high"`
	Low       float64   `json:"low" db:"low"`
	Close     float64   `json:"close" db:"close"`
	PreClose  float64   `json:"pre_close" db:"pre_close"`
	Change    float64   `json:"change" db:"change"`
	PctChg    float64   `json:"pct_chg" db:"pct_chg"`
	Vol       float64   `json:"vol" db:"vol"`
	Amount    float64   `json:"amount" db:"amount"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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
