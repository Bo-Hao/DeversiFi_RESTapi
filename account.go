package dvfapi

import (
	"fmt"
	"io/ioutil"
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

func (p *Client) GetFeeRate(token string) (result *GetFeeRateResponse, err error) {
	nonce := time.Now().Unix()
	s, err := p.sign(strconv.FormatInt(nonce, 10))
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	params["nonce"] = strconv.FormatInt(nonce, 10)
	params["signature"] = s
	params["token"] = token
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

type GetBalanceResponse []struct {
	Balance       int    `json:"balance"`
	ActiveBalance int    `json:"activeBalance"`
	ID            string `json:"_id"`
	EthAddress    string `json:"ethAddress"`
	Token         string `json:"token"`
}

func (p *Client) GetBalance(token string) (result *GetBalanceResponse, err error) {
	nonce := time.Now().Add(time.Second).Unix()
	nonceStr := strconv.FormatInt(nonce, 10)
	s, err := p.sign(nonceStr)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	params["nonce"] = nonceStr
	params["signature"] = s
	params["token"] = token
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

type GetUserConfigResponse struct {
	Dvf struct {
		StarkExVersion                         string  `json:"starkExVersion"`
		DefaultFeeRate                         float64 `json:"defaultFeeRate"`
		DefaultFeeRateSwap                     float64 `json:"defaultFeeRateSwap"`
		DeversifiAddress                       string  `json:"deversifiAddress"`
		StarkExContractAddress                 string  `json:"starkExContractAddress"`
		WithdrawalBalanceReaderContractAddress string  `json:"withdrawalBalanceReaderContractAddress"`
		StarkExTransferRegistryContractAddress string  `json:"starkExTransferRegistryContractAddress"`
		RegistrationAndDepositInterfaceAddress string  `json:"registrationAndDepositInterfaceAddress"`
		AMMfactoryAddress                      string  `json:"AMMfactoryAddress"`
		AMMrouterAddress                       string  `json:"AMMrouterAddress"`
		BridgeConfigPerChain                   struct {
			MaticPos struct {
				ContractAddress       string  `json:"contractAddress"`
				WithdrawalFeeRatio    int     `json:"withdrawalFeeRatio"`
				MaxWithdrawalRatio    float64 `json:"maxWithdrawalRatio"`
				MaxTotalUsdInContract int     `json:"maxTotalUsdInContract"`
			} `json:"MATIC_POS"`
		} `json:"bridgeConfigPerChain"`
		ExchangeSymbols      []string `json:"exchangeSymbols"`
		DlmMarkets           []string `json:"dlmMarkets"`
		TempStarkVaultID     int      `json:"tempStarkVaultId"`
		MinDepositUSDT       int      `json:"minDepositUSDT"`
		AuthVersion          int      `json:"authVersion"`
		DisableLP            bool     `json:"disableLP"`
		DeversifiStarkKeyHex string   `json:"deversifiStarkKeyHex"`
	} `json:"DVF"`
	TokenBalancesHistory []string `json:"tokenBalancesHistory"`
	TradingRewards       struct {
		WeeklyAmount    int      `json:"weeklyAmount"`
		RewardToken     string   `json:"rewardToken"`
		ExcludedMarkets []string `json:"excludedMarkets"`
	} `json:"tradingRewards"`
	AmmPools      map[string]AmmPools      `json:"ammPools"`
	TokenRegistry map[string]TokenRegistry `json:"tokenRegistry"`
	IsRegistered  bool                     `json:"isRegistered"`
	EthAddress    string                   `json:"ethAddress"`
}

// Returns the DeversiFi application and user configuration details.
func (p *Client) GetUserConfig() (result *GetUserConfigResponse, err error) {
	nonce := time.Now().Add(time.Second).Unix()
	nonceStr := strconv.FormatInt(nonce, 10)
	s, err := p.sign(nonceStr)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	params["nonce"] = nonceStr
	params["signature"] = s
	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	res, err := p.sendRequest(http.MethodPost, "/v1/trading/r/getUserConf", jsonBody, nil)
	if err != nil {
		return nil, err
	}

	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type RegisterResponse struct {
	Dvf struct {
		StarkExVersion                         string  `json:"starkExVersion"`
		DefaultFeeRate                         float64 `json:"defaultFeeRate"`
		DefaultFeeRateSwap                     float64 `json:"defaultFeeRateSwap"`
		DeversifiAddress                       string  `json:"deversifiAddress"`
		StarkExContractAddress                 string  `json:"starkExContractAddress"`
		WithdrawalBalanceReaderContractAddress string  `json:"withdrawalBalanceReaderContractAddress"`
		StarkExTransferRegistryContractAddress string  `json:"starkExTransferRegistryContractAddress"`
		RegistrationAndDepositInterfaceAddress string  `json:"registrationAndDepositInterfaceAddress"`
		AMMfactoryAddress                      string  `json:"AMMfactoryAddress"`
		AMMrouterAddress                       string  `json:"AMMrouterAddress"`
		BridgeConfigPerChain                   struct {
			MaticPos struct {
				ContractAddress       string  `json:"contractAddress"`
				WithdrawalFeeRatio    int     `json:"withdrawalFeeRatio"`
				MaxWithdrawalRatio    float64 `json:"maxWithdrawalRatio"`
				MaxTotalUsdInContract int     `json:"maxTotalUsdInContract"`
			} `json:"MATIC_POS"`
		} `json:"bridgeConfigPerChain"`
		ExchangeSymbols      []string `json:"exchangeSymbols"`
		DlmMarkets           []string `json:"dlmMarkets"`
		TempStarkVaultID     int      `json:"tempStarkVaultId"`
		MinDepositUSDT       int      `json:"minDepositUSDT"`
		AuthVersion          int      `json:"authVersion"`
		DisableLP            bool     `json:"disableLP"`
		DeversifiStarkKeyHex string   `json:"deversifiStarkKeyHex"`
	} `json:"DVF"`
	TokenBalancesHistory []string `json:"tokenBalancesHistory"`
	TradingRewards       struct {
		WeeklyAmount    int      `json:"weeklyAmount"`
		RewardToken     string   `json:"rewardToken"`
		ExcludedMarkets []string `json:"excludedMarkets"`
	} `json:"tradingRewards"`
	AmmPools      map[string]AmmPools      `json:"ammPools"`
	TokenRegistry map[string]TokenRegistry `json:"tokenRegistry"`
	IsRegistered  bool                     `json:"isRegistered"`
	EthAddress    string                   `json:"ethAddress"`
}

// This method is used to register a Stark key that corresponds to an Ethereum public address. This will return deversifi Signature or DeversiFi application and user configuration details.
func (p *Client) Register() (result *RegisterResponse, err error) {
	nonce := time.Now().Add(time.Second).Unix()
	nonceStr := strconv.FormatInt(nonce, 10)
	s, publ, err := p.signAndPublicKey(nonceStr)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	params["starkKey"] = publ
	params["nonce"] = nonceStr
	params["signature"] = s
	fmt.Println(publ)
	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	res, err := p.sendRequest(http.MethodPost, "/v1/trading/w/register", jsonBody, nil)
	if err != nil {
		return nil, err
	}

	b, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(b))
	err = decode(res, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type GetUserBalancesResponse struct {
}

// This is used to retrieve the total and active balances of a user per token. Active balance is the balance that is currently available. Total balance (specified as balance) is the sum of all the balances including those locked for trading.
func (p *Client) GetUserBalances() (result *GetUserBalancesResponse, err error) {
	nonce := time.Now().Add(time.Second).Unix()
	nonceStr := strconv.FormatInt(nonce, 10)
	s, err := p.sign(nonceStr)
	if err != nil {
		return nil, err
	}
	params := make(map[string]interface{})

	params["nonce"] = nonceStr
	params["signature"] = s
	params["fields"] = []string{"balance", "updatedAt"}

	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	res, err := p.sendRequest(http.MethodPost, "/v1/trading/r/getBalanceForUser/"+p.subaccount, jsonBody, nil)
	if err != nil {
		return nil, err
	}

	b, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(b))
	/* err = decode(res, &result)
	if err != nil {
		return nil, err
	} */
	return result, nil
}
