package api

import (
	"context"
	"time"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/log"
)

type Worker struct {
	interval time.Duration
	timer    *time.Timer
	logger   log.Logger
	registry chain.Registry
}

func NewWorker(interval time.Duration, logger log.Logger, registry chain.Registry) *Worker {
	worker := &Worker{
		interval: interval,
		logger:   logger,
		registry: registry,
	}

	return worker
}

func (w *Worker) Start() {
	if w.timer != nil {
		w.timer.Stop()
	}

	w.timer = time.AfterFunc(w.interval, func() {
		w.timer.Stop()
		if err := w.registry.UploadChainInfo(context.Background()); err != nil {
			w.logger.Error(err)
		}
		w.timer.Reset(w.interval)
	})
}

func (w *Worker) Stop() {
	w.timer.Stop()
}
