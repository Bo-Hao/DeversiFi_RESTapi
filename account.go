package dvfapi

import (
	"net/http"
	"strconv"
	"time"
)

type GetFeeRateResponse struct {
	Address   string `json:"address"`
	Timestamp int64  `json:"timestamp"`
	Fees      struct {
		Maker int `json:"maker"`
		Taker int `json:"taker"`
	} `json:"fees"`
}

func (p *Client) GetFeeRate() (result *GetFeeRateResponse, err error) {
	nonce := time.Now().Unix()
	s, err := p.sign()
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	params["nonce"] = strconv.FormatInt(nonce, 10)
	params["signature"] = s
	params["token"] = "ETH"
	res, err := p.sendRequest(http.MethodGet, "/v1/trading/r/feeRate", nil, &params)
	if err != nil {
		return nil, err
	}
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type GetBalanceResponse struct {
	Balance       int    `json:"balance"`
	ActiveBalance int    `json:"activeBalance"`
	ID            string `json:"_id"`
	EthAddress    string `json:"ethAddress"`
	Token         string `json:"token"`
}

func (p *Client) GetBalance() (result *GetBalanceResponse, err error) {
	s, err := p.sign()
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	nonce := time.Now().Unix()
	params["nonce"] = strconv.FormatInt(nonce, 10)
	params["signature"] = s
	params["token"] = "ETH"

	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	res, err := p.sendRequest(http.MethodPost, "/v1/trading/r/getBalance", jsonBody, nil)
	if err != nil {
		return nil, err
	}
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
