package chain

import "github.com/Mobile-Web3/backend/pkg/cosmos"

type Service struct {
	client *cosmos.ChainClient
}

func NewService(client *cosmos.ChainClient) *Service {
	return &Service{
		client: client,
	}
}

type ShortResponse struct {
	ID          string `json:"chainId"`
	Name        string `json:"chainName"`
	PrettyName  string `json:"prettyName"`
	Prefix      string `json:"bech32Prefix"`
	Slip44      uint32 `json:"slip44"`
	Description string `json:"description"`
	Base        string `json:"base"`
	Symbol      string `json:"symbol"`
	Display     string `json:"display"`
	LogoPngURL  string `json:"logoPngUrl"`
	LogoSvgURL  string `json:"logoSvgUrl"`
}

func (s *Service) GetChains() []ShortResponse {
	chains := s.client.GetAllChains()
	var result []ShortResponse
	for _, chain := range chains {
		result = append(result, ShortResponse{
			ID:          chain.ID,
			Name:        chain.Name,
			PrettyName:  chain.PrettyName,
			Prefix:      chain.Prefix,
			Slip44:      chain.Slip44,
			Description: chain.Asset.Description,
			Base:        chain.Asset.Base,
			Symbol:      chain.Asset.Symbol,
			Display:     chain.Asset.Display,
			LogoPngURL:  chain.Asset.Logo.Png,
			LogoSvgURL:  chain.Asset.Logo.Svg,
		})
	}

	return result
}
