package dvfapi

import (
	"net/http"
)

type GetConfigResponse struct {
	Dvf struct {
		StarkExVersion                         string  `json:"starkExVersion"`
		DefaultFeeRate                         float64 `json:"defaultFeeRate"`
		DefaultFeeRateSwsap                    float64 `json:"defaultFeeRateSwap"`
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
}

type TokenRegistry struct {
	Decimals             int     `json:"decimals"`
	Quantization         int64   `json:"quantization"`
	MinOrderSize         float64 `json:"minOrderSize"`
	TransferFee          float64 `json:"transferFee"`
	StarkTokenID         string  `json:"starkTokenId"`
	TokenAddressPerChain struct {
		Ethereum string `json:"ETHEREUM"`
	} `json:"tokenAddressPerChain"`
	FastWithdrawalRequiredGas int    `json:"fastWithdrawalRequiredGas"`
	CoingeckoID               string `json:"coingeckoId"`
	DeployedAtBlock           int    `json:"deployedAtBlock"`
}

type AmmPools struct {
	Tokens             []string `json:"tokens"`
	LpToken            string   `json:"lpToken"`
	SwapFees           float64  `json:"swapFees"`
	PoolFee            float64  `json:"poolFee"`
	OperatorFee        float64  `json:"operatorFee"`
	WeeklyMiningReward int      `json:"weeklyMiningReward"`
	Enabled            bool     `json:"enabled"`
}

func (p *Client) GetConfig() (result *GetConfigResponse, err error) {
	r := GetConfigResponse{}
	tokenRegistry := make(map[string]TokenRegistry)
	ammPools := make(map[string]AmmPools)
	r.TokenRegistry = tokenRegistry
	r.AmmPools = ammPools
	res, err := p.sendRequest(http.MethodPost, "/v1/trading/r/getConf", nil, nil)
	if err != nil {
		return nil, err
	}

	err = decode(res, &r)
	if err != nil {
		return nil, err
	}
	result = &r
	return result, nil
}
