package chain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	DefaultLowGasPrice     = 0.01
	DefaultAverageGasPrice = 0.025
	DefaultHighGasPrice    = 0.04
)

var (
	ErrBaseDenomNotFound = errors.New("base denom not found")
	ErrInvalidAmount     = errors.New("invalid amount")
)

type FeeToken struct {
	Denom           string  `json:"denom"`
	MinGasPrice     float64 `json:"fixed_min_gas_price"`
	LowGasPrice     float64 `json:"low_gas_price"`
	AverageGasPrice float64 `json:"average_gas_price"`
	HighGasPrice    float64 `json:"high_gas_price"`
}

type Fee struct {
	FeeTokens []FeeToken `json:"fee_tokens"`
}

type Rpc struct {
	Address  string `json:"address"`
	Provider string `json:"provider"`
}

type Api struct {
	Rpc []Rpc `json:"rpc"`
}

type DenomUnit struct {
	Denom    string `json:"denom"`
	Exponent int    `json:"exponent"`
}

type Logo struct {
	Png string `json:"png"`
	Svg string `json:"svg"`
}

type Asset struct {
	Description string      `json:"description"`
	Base        string      `json:"base"`
	Symbol      string      `json:"symbol"`
	Display     string      `json:"display"`
	Logo        Logo        `json:"logo_URIs"`
	DenomUnits  []DenomUnit `json:"denom_units"`
}

type Chain struct {
	ID              string   `json:"chain_id"`
	Name            string   `json:"chain_name"`
	PrettyName      string   `json:"pretty_name"`
	Prefix          string   `json:"bech32_prefix"`
	Slip44          uint32   `json:"slip44"`
	LowGasPrice     float64  `json:"lowGasPrice"`
	AverageGasPrice float64  `json:"averageGasPrice"`
	HighGasPrice    float64  `json:"highGasPrice"`
	KeyAlgos        []string `json:"key_algos"`
	Fees            Fee      `json:"fees"`
	Api             Api      `json:"apis"`
	Asset           Asset    `json:"asset,omitempty"`
}

func GetBaseDenom(base string, display string, denoms []DenomUnit) (denom string, exponent int, err error) {
	for _, unit := range denoms {
		if unit.Denom == base {
			denom = unit.Denom
			for _, displayUnit := range denoms {
				if displayUnit.Denom == display {
					exponent = displayUnit.Exponent
					break
				}
			}
			return
		}
	}

	err = ErrBaseDenomNotFound
	return
}

func FromBaseToDisplay(amount string, exponent int) string {
	if amount == "0" {
		return amount
	}
	result := ""
	index := len(amount) - 1
	count := 0
	for {
		if index == -1 {
			if count == exponent {
				result = "0." + result
				return result
			}
			if count > exponent {
				return result
			}

			count++
			result = "0" + result
			continue
		}

		if count == exponent {
			result = "." + result
		}

		result = string(amount[index]) + result
		index--
		count++
	}
}

func FromDisplayToBase(amount string, denom string, exponent int) (string, error) {
	var sb strings.Builder
	if !strings.Contains(amount, ".") {
		sb.WriteString(amount)
		for i := 0; i < exponent; i++ {
			sb.WriteString("0")
		}
		sb.WriteString(denom)
		return sb.String(), nil
	}

	amountValues := strings.Split(amount, ".")

	if len(amountValues) > 2 {
		return "", ErrInvalidAmount
	}

	if len(amountValues[0]) > 0 && rune(amountValues[0][0]) != '0' {
		for _, symbol := range amountValues[0] {
			sb.WriteRune(symbol)
		}
		for i := 0; i < exponent; i++ {
			if len(amountValues[1]) < i+1 {
				sb.WriteString("0")
				continue
			}
			sb.WriteRune(rune(amountValues[1][i]))
		}
		sb.WriteString(denom)
		return sb.String(), nil
	}

	for i := 0; i < exponent; i++ {
		if len(amountValues[1]) < i+1 {
			sb.WriteString("0")
			continue
		}
		if rune(amountValues[1][i]) == '0' {
			continue
		}
		sb.WriteRune(rune(amountValues[1][i]))
	}
	if sb.Len() == 0 {
		sb.WriteString("0")
	}
	sb.WriteString(denom)
	return sb.String(), nil
}

func GetGasPrices(base string, feeTokens []FeeToken) (lowGasPrice float64, averageGasPrice float64, highGasPrice float64) {
	for _, feeToken := range feeTokens {
		if feeToken.Denom == base {
			if feeToken.LowGasPrice <= 0 && feeToken.AverageGasPrice <= 0 && feeToken.HighGasPrice <= 0 {
				lowGasPrice = DefaultLowGasPrice + feeToken.MinGasPrice
				averageGasPrice = DefaultAverageGasPrice + feeToken.MinGasPrice
				highGasPrice = DefaultHighGasPrice + feeToken.MinGasPrice
				return
			}

			lowGasPrice = feeToken.LowGasPrice + feeToken.MinGasPrice
			averageGasPrice = feeToken.AverageGasPrice + feeToken.MinGasPrice
			highGasPrice = feeToken.HighGasPrice + feeToken.MinGasPrice
			return
		}
	}

	return
}

func ValidateRPCUrls(rpc []Rpc) ([]Rpc, error) {
	var result []Rpc
	for _, endpoint := range rpc {
		u, err := url.Parse(endpoint.Address)
		if err != nil {
			return nil, err
		}

		var port string
		if u.Port() == "" {
			switch u.Scheme {
			case "https":
				port = "443"
			case "http":
				port = "80"
			default:
				continue
			}
		} else {
			port = u.Port()
		}

		result = append(result, Rpc{
			Provider: endpoint.Provider,
			Address:  fmt.Sprintf("%s://%s:%s%s", u.Scheme, u.Hostname(), port, u.Path),
		})
	}

	return result, nil
}
