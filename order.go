package dvfapi

import (
	"bytes"
	"net/http"
	"strconv"
	"time"
)

type GetOrderResponse struct {
	ID          string  `json:"_id"`
	Symbol      string  `json:"symbol"`
	Amount      float64 `json:"amount"`
	Price       int     `json:"price"`
	TotalFilled int     `json:"totalFilled"`
	Pending     bool    `json:"pending"`
	Canceled    bool    `json:"canceled"`
	Active      bool    `json:"active"`
}

// This is endpoint is used to retrieve the details for a specific order using the order ID.
func (p *Client) GetOrder(orderId string) (result []*GetOrderResponse, err error) {
	nonce := time.Now().Add(time.Second).Unix()
	nonceStr := strconv.FormatInt(nonce, 10)
	s, err := p.sign(nonceStr)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	params["orderId"] = orderId
	//params["cid"] = ""
	params["nonce"] = nonceStr
	params["signature"] = s

	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	res, err := p.sendRequest(http.MethodPost, "/v1/trading/r/getOrder", jsonBody, nil)
	if err != nil {
		return nil, err
	}

	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type GetAllOrdersResponse []struct {
	ID           string    `json:"_id"`
	User         string    `json:"user"`
	Symbol       string    `json:"symbol"`
	Amount       float64   `json:"amount"`
	TotalFilled  float64   `json:"totalFilled"`
	Price        int       `json:"price"`
	AveragePrice float64   `json:"averagePrice"`
	FeeRate      string    `json:"feeRate"`
	TokenBuy     string    `json:"tokenBuy"`
	TotalBought  string    `json:"totalBought"`
	TokenSell    string    `json:"tokenSell"`
	TotalSold    string    `json:"totalSold"`
	Active       bool      `json:"active"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	ActivatedAt  time.Time `json:"activatedAt"`
}

// This endpoints allows to retrieve details on all open orders.
func (p *Client) GetAllOrders(base, quote string) (result *GetAllOrdersResponse, err error) {
	var buffer bytes.Buffer
	buffer.WriteString(base)
	buffer.WriteString(":")
	buffer.WriteString(quote)
	nonce := time.Now().Add(time.Second).Unix()
	nonceStr := strconv.FormatInt(nonce, 10)
	s, err := p.sign(nonceStr)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	params["nonce"] = nonceStr
	params["signature"] = s
	params["symbol"] = buffer.String()
	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	res, err := p.sendRequest(http.MethodPost, "/v1/trading/r/openOrders", jsonBody, nil)
	if err != nil {
		return nil, err
	}

	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type CancelOrderResponse struct {
	OrderId  string `json:"orderId"`
	Canceled bool   `json:"canceled"`
}

// This endpoint allows to cancel a specific order.
func (p *Client) CancelOrder(orderId string) (result *CancelOrderResponse, err error) {
	nonce := time.Now().Add(time.Second).Unix()
	nonceStr := strconv.FormatInt(nonce, 10)
	s, err := p.sign(nonceStr)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)

	params["orderId"] = orderId
	//params["cid"] = ""
	params["nonce"] = nonceStr
	params["signature"] = s

	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	res, err := p.sendRequest(http.MethodPost, "/v1/trading/w/cancelOrder", jsonBody, nil)
	if err != nil {
		return nil, err
	}

	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SubmitOrderResponse struct {
}

// func (p *Client) SubmitOrder() (result *SubmitOrderResponse, err error) {
// 	s, err := p.sign()
// 	if err != nil {
// 		return nil, err
// 	}
// 	params := make(map[string]interface{})
// 	meta := make(map[string]interface{})
// 	starkOrder := make(map[string]interface{})

// 	nonce := time.Now().Unix()

// 	params["cid"] = ""
// 	params["type"] = ""
// 	params["symbol"] = ""
// 	params["amount"] = ""
// 	params["price"] = ""

// 	starkOrder["vaultIdSell"] = ""
// 	starkOrder["vaultIdBuy"] = ""
// 	starkOrder["amountSell"] = ""
// 	starkOrder["amountBuy"] = ""
// 	starkOrder["tokenSell"] = ""
// 	starkOrder["tokenBuy"] = ""
// 	starkOrder["nonce"] = strconv.FormatInt(nonce, 10)
// 	starkOrder["expirationTimestamp"] = ""

// 	meta["starkOrder"] = starkOrder
// 	meta["starkMessage"] = ""
// 	meta["ethAddress"] = ""
// 	meta["starkPublicKey"] = ""
// 	meta["starkSignature"] = ""

// 	params["nonce"] = strconv.FormatInt(nonce, 10)
// 	params["signature"] = s

// 	jsonBody, err := json.Marshal(params)
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, err := p.sendRequest(http.MethodPost, "/v1/trading/w/cancelOrder", jsonBody, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = decode(res, &result)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return result, nil
// }
