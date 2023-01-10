package chain

import "context"

type Service struct {
	registry   Registry
	repository Repository
}

func NewService(registry Registry, repository Repository) *Service {
	return &Service{
		registry:   registry,
		repository: repository,
	}
}

func (s *Service) UpdateChainInfo(ctx context.Context) error {
	chains, err := s.registry.UploadChainInfo(ctx)
	if err != nil {
		return err
	}

	return s.repository.UpdateChains(ctx, chains)
}
