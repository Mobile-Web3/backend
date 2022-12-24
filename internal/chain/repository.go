package chain

import "context"

type ShortResponse struct {
	ID          string   `json:"chainId"`
	Name        string   `json:"chainName"`
	PrettyName  string   `json:"prettyName"`
	Prefix      string   `json:"bech32Prefix"`
	Slip44      uint32   `json:"slip44"`
	Description string   `json:"description"`
	Base        string   `json:"base"`
	Symbol      string   `json:"symbol"`
	Display     string   `json:"display"`
	LogoPngURL  string   `json:"logoPngUrl"`
	LogoSvgURL  string   `json:"logoSvgUrl"`
	KeyAlgos    []string `json:"keyAlgos"`
}

type Repository interface {
	GetAllChains(ctx context.Context) ([]ShortResponse, error)
	GetAllPrefixes(ctx context.Context) ([]string, error)
	GetChainByPrefix(ctx context.Context, prefix string) (Chain, error)
	UpdateChains(ctx context.Context, chains []Chain) error
}
