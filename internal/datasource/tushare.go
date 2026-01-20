package datasource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"quant-data-engine/internal/config"

	"github.com/sirupsen/logrus"
)

// TushareClientInterface Tushare 客户端接口
type TushareClientInterface interface {
	GetStockBasic(req *StockBasicRequest, fields []string) (*TushareResponse, error)
	GetTradeCal(req *TradeCalRequest, fields []string) (*TushareResponse, error)
	GetNewShare(req *NewShareRequest, fields []string) (*TushareResponse, error)
	GetStockCompany(req *StockCompanyRequest, fields []string) (*TushareResponse, error)
	GetStkManagers(req *StkManagersRequest, fields []string) (*TushareResponse, error)
	GetStkRewards(req *StkRewardsRequest, fields []string) (*TushareResponse, error)
	GetDaily(req *DailyRequest, fields []string) (*TushareResponse, error)
}

// TushareClient Tushare API客户端
type TushareClient struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewTushareClient 创建Tushare API客户端
func NewTushareClient() *TushareClient {
	cfg := config.AppConfig
	return &TushareClient{
		apiURL: "http://api.tushare.pro",
		apiKey: cfg.TushareAPIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TushareRequest Tushare API请求参数
type TushareRequest struct {
	Token   string                 `json:"token"`
	APIName string                 `json:"api_name"`
	Params  map[string]interface{} `json:"params"`
	Fields  string                 `json:"fields"`
}

// TushareResponse Tushare API响应
type TushareResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    *DataResult `json:"data"`
}

// DataResult 数据结果
type DataResult struct {
	Fields []string        `json:"fields"`
	Items  [][]interface{} `json:"items"`
}

// StockBasicRequest 股票基础信息请求参数
type StockBasicRequest struct {
	TSCode     string `json:"ts_code,omitempty"`
	Name       string `json:"name,omitempty"`
	Market     string `json:"market,omitempty"`
	ListStatus string `json:"list_status,omitempty"`
	Exchange   string `json:"exchange,omitempty"`
	IsHS       string `json:"is_hs,omitempty"`
}

// TradeCalRequest 交易日历请求参数
type TradeCalRequest struct {
	Exchange  string `json:"exchange,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	IsOpen    string `json:"is_open,omitempty"`
}

// NewShareRequest 新股上市列表请求参数
type NewShareRequest struct {
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// StockCompanyRequest 上市公司基础信息请求参数
type StockCompanyRequest struct {
	TSCode   string `json:"ts_code,omitempty"`
	Exchange string `json:"exchange,omitempty"`
}

// StkManagersRequest 上市公司管理层请求参数
type StkManagersRequest struct {
	TSCode    string `json:"ts_code,omitempty"`
	AnnDate   string `json:"ann_date,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// StkRewardsRequest 管理层薪酬和持股请求参数
type StkRewardsRequest struct {
	TSCode  string `json:"ts_code,omitempty"`
	EndDate string `json:"end_date,omitempty"`
}

// DailyRequest A股日线行情请求参数
type DailyRequest struct {
	TSCode    string `json:"ts_code,omitempty"`
	TradeDate string `json:"trade_date,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// GetStockBasic 获取股票基础信息
func (c *TushareClient) GetStockBasic(req *StockBasicRequest, fields []string) (*TushareResponse, error) {
	params := make(map[string]interface{})
	if req.TSCode != "" {
		params["ts_code"] = req.TSCode
	}
	if req.Name != "" {
		params["name"] = req.Name
	}
	if req.Market != "" {
		params["market"] = req.Market
	}
	if req.ListStatus != "" {
		params["list_status"] = req.ListStatus
	}
	if req.Exchange != "" {
		params["exchange"] = req.Exchange
	}
	if req.IsHS != "" {
		params["is_hs"] = req.IsHS
	}

	return c.callAPI("stock_basic", params, fields)
}

// GetTradeCal 获取交易日历
func (c *TushareClient) GetTradeCal(req *TradeCalRequest, fields []string) (*TushareResponse, error) {
	params := make(map[string]interface{})
	if req.Exchange != "" {
		params["exchange"] = req.Exchange
	}
	if req.StartDate != "" {
		params["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		params["end_date"] = req.EndDate
	}
	if req.IsOpen != "" {
		params["is_open"] = req.IsOpen
	}

	return c.callAPI("trade_cal", params, fields)
}

// GetNewShare 获取新股上市列表
func (c *TushareClient) GetNewShare(req *NewShareRequest, fields []string) (*TushareResponse, error) {
	params := make(map[string]interface{})
	if req.StartDate != "" {
		params["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		params["end_date"] = req.EndDate
	}

	return c.callAPI("new_share", params, fields)
}

// GetStockCompany 获取上市公司基础信息
func (c *TushareClient) GetStockCompany(req *StockCompanyRequest, fields []string) (*TushareResponse, error) {
	params := make(map[string]interface{})
	if req.TSCode != "" {
		params["ts_code"] = req.TSCode
	}
	if req.Exchange != "" {
		params["exchange"] = req.Exchange
	}

	return c.callAPI("stock_company", params, fields)
}

// GetStkManagers 获取上市公司管理层
func (c *TushareClient) GetStkManagers(req *StkManagersRequest, fields []string) (*TushareResponse, error) {
	params := make(map[string]interface{})
	if req.TSCode != "" {
		params["ts_code"] = req.TSCode
	}
	if req.AnnDate != "" {
		params["ann_date"] = req.AnnDate
	}
	if req.StartDate != "" {
		params["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		params["end_date"] = req.EndDate
	}

	return c.callAPI("stk_managers", params, fields)
}

// GetStkRewards 获取管理层薪酬和持股
func (c *TushareClient) GetStkRewards(req *StkRewardsRequest, fields []string) (*TushareResponse, error) {
	params := make(map[string]interface{})
	if req.TSCode != "" {
		params["ts_code"] = req.TSCode
	}
	if req.EndDate != "" {
		params["end_date"] = req.EndDate
	}

	return c.callAPI("stk_rewards", params, fields)
}

// GetDaily 获取A股日线行情
func (c *TushareClient) GetDaily(req *DailyRequest, fields []string) (*TushareResponse, error) {
	params := make(map[string]interface{})
	if req.TSCode != "" {
		params["ts_code"] = req.TSCode
	}
	if req.TradeDate != "" {
		params["trade_date"] = req.TradeDate
	}
	if req.StartDate != "" {
		params["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		params["end_date"] = req.EndDate
	}

	return c.callAPI("daily", params, fields)
}

// callAPI 调用Tushare API
func (c *TushareClient) callAPI(apiName string, params map[string]interface{}, fields []string) (*TushareResponse, error) {
	// 将fields数组转换为逗号分隔的字符串
	fieldsStr := strings.Join(fields, ",")

	request := &TushareRequest{
		Token:   c.apiKey,
		APIName: apiName,
		Params:  params,
		Fields:  fieldsStr,
	}

	// 输出请求详细信息
	logrus.Debugf("Calling Tushare API method: %s", apiName)
	logrus.Debugf("API URL: %s", c.apiURL)
	logrus.Debugf("Request params: %+v", params)
	logrus.Debugf("Request fields: %+v", fields)
	logrus.Debugf("Request fields string: %s", fieldsStr)

	jsonData, err := json.Marshal(request)
	if err != nil {
		logrus.Errorf("Failed to marshal request: %v", err)
		return nil, err
	}

	logrus.Debugf("Request JSON: %s", string(jsonData))

	resp, err := c.httpClient.Post(c.apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.Errorf("Failed to call Tushare API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	logrus.Debugf("Response status: %s", resp.Status)
	logrus.Debugf("Response headers: %+v", resp.Header)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Failed to read response body: %v", err)
		return nil, err
	}

	logrus.Debugf("Response body: %s", string(body))

	var tushareResp TushareResponse
	if err := json.Unmarshal(body, &tushareResp); err != nil {
		logrus.Errorf("Failed to unmarshal response: %v", err)
		return nil, err
	}

	logrus.Debugf("Parsed response: %+v", tushareResp)

	if tushareResp.Code != 0 {
		logrus.Errorf("Tushare API error: code=%d, message=%s", tushareResp.Code, tushareResp.Message)
		return nil, fmt.Errorf("Tushare API error: %s", tushareResp.Message)
	}

	logrus.Debugf("Tushare API call successful")
	return &tushareResp, nil
}
