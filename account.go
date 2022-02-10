package dvfapi

import (
	"net/http"
	"strconv"
	"time"
)

type GetFeeRateRespones struct {
	Address   string `json:"address"`
	Timestamp int64  `json:"timestamp"`
	Fees      struct {
		Maker int `json:"maker"`
		Taker int `json:"taker"`
	} `json:"fees"`
}

func (p *Client) GetFeeRate() (result *GetFeeRateRespones, err error) {
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
