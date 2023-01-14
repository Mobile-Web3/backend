package api

import (
	"context"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/robfig/cron/v3"
)

type Worker struct {
	logger       log.Logger
	chainService *chain.Service
	scheduler    *cron.Cron
	jobID        cron.EntryID
}

func NewWorker(logger log.Logger, chainService *chain.Service) *Worker {
	worker := &Worker{
		logger:       logger,
		scheduler:    cron.New(),
		chainService: chainService,
	}

	return worker
}

func (w *Worker) Start() error {
	jobID, err := w.scheduler.AddFunc("0 0 * * *", func() {
		_ = w.chainService.UpdateChainInfo(context.Background())
		w.logger.Info("cron worked out")
	})
	if err != nil {
		return err
	}

	w.jobID = jobID
	w.scheduler.Start()
	return nil
}

func (w *Worker) Stop() {
	w.scheduler.Remove(w.jobID)
	w.scheduler.Stop()
}
