package cosmos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	registry "github.com/strangelove-ventures/lens/client/chain_registry"
)

var (
	ErrBaseDenomNotFound = errors.New("base denom not found")
	ErrInvalidAmount     = errors.New("invalid amount")
)

var excludingDirs = []string{".git", ".github", "_IBC", "_non-cosmos", "testnets", "thorchain"}

type DenomUnit struct {
	Denom    string `json:"denom"`
	Exponent int    `json:"exponent"`
}

type Asset struct {
	Description string      `json:"description"`
	Base        string      `json:"base"`
	Symbol      string      `json:"symbol"`
	Display     string      `json:"display"`
	DenomUnits  []DenomUnit `json:"denom_units"`
}

type Rpc struct {
	Address  string `json:"address"`
	Provider string `json:"provider"`
}

type Api struct {
	Rpc []Rpc `json:"rpc"`
}

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

type Chain struct {
	ID         string             `json:"chain_id"`
	Name       string             `json:"chain_name"`
	PrettyName string             `json:"pretty_name"`
	Prefix     string             `json:"bech32_prefix"`
	Slip44     uint32             `json:"slip44"`
	Fees       Fee                `json:"fees"`
	Api        Api                `json:"apis"`
	Info       registry.ChainInfo `json:"-"`
	Asset      Asset              `json:"-"`

	isConnectionInit bool
	activeRPC        string
	rpcMutex         sync.RWMutex
	rpcLifetime      time.Duration
}

func (c *Chain) invalidateRPC() {
	c.rpcMutex.Lock()
	c.isConnectionInit = false
	c.rpcMutex.Unlock()
}

func (c *Chain) getRpc(ctx context.Context) (string, error) {
	c.rpcMutex.RLock()
	if c.isConnectionInit {
		c.rpcMutex.RUnlock()
		return c.activeRPC, nil
	}
	c.rpcMutex.RUnlock()
	c.rpcMutex.Lock()
	rpc, err := c.Info.GetRandomRPCEndpoint(ctx)
	if err != nil {
		c.rpcMutex.Unlock()
		return "", err
	}
	c.activeRPC = rpc
	c.isConnectionInit = true
	c.rpcMutex.Unlock()
	time.AfterFunc(c.rpcLifetime, c.invalidateRPC)
	return rpc, nil
}

func (c *Chain) GetLowGasPrice() string {
	for _, feeToken := range c.Fees.FeeTokens {
		if feeToken.Denom == c.Asset.Base {
			if feeToken.LowGasPrice == 0 {
				return "0.01" + feeToken.Denom
			}
			return fmt.Sprintf("%f%s", feeToken.LowGasPrice, feeToken.Denom)
		}
	}
	return "0.01" + c.Asset.Base
}

func (c *Chain) GetBaseDenom() (denom string, exponent int, err error) {
	for _, unit := range c.Asset.DenomUnits {
		if unit.Denom == c.Asset.Base {
			denom = unit.Denom
			for _, displayUnit := range c.Asset.DenomUnits {
				if displayUnit.Denom == c.Asset.Display {
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

func (c *Chain) FromDisplayToBase(amount string) (string, error) {
	denom, exponent, err := c.GetBaseDenom()
	if err != nil {
		return "", err
	}

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
		if rune(amountValues[1][i]) == '0' {
			continue
		}
		if len(amountValues[1]) < i+1 {
			sb.WriteString("0")
			continue
		}
		sb.WriteRune(rune(amountValues[1][i]))
	}
	sb.WriteString(denom)
	return sb.String(), nil
}
