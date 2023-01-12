package api

import (
	"context"
	"time"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
)

type Worker struct {
	interval     time.Duration
	timer        *time.Timer
	chainService *chain.Service
}

func NewWorker(interval time.Duration, chainService *chain.Service) *Worker {
	worker := &Worker{
		interval:     interval,
		chainService: chainService,
	}

	return worker
}

func (w *Worker) Start() {
	if w.timer != nil {
		w.timer.Stop()
	}

	w.timer = time.AfterFunc(w.interval, func() {
		w.timer.Stop()
		_ = w.chainService.UpdateChainInfo(context.Background())
		w.timer.Reset(w.interval)
	})
}

func (w *Worker) Stop() {
	w.timer.Stop()
}
