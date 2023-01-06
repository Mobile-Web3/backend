package chain

import "context"

type Registry interface {
	UploadChainInfo(ctx context.Context) error
}
