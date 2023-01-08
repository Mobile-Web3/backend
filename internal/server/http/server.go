package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/Mobile-Web3/backend/pkg/log"
)

type Server struct {
	logger log.Logger
	server *http.Server
}

func New(port string, logger log.Logger, handler http.Handler) *Server {
	return &Server{
		logger: logger,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
	}
}

func (s *Server) Start() {
	s.logger.Info("server started")
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error(err)
	}
}

func (s *Server) Stop() {
	s.logger.Info("server shutting down")
	if err := s.server.Shutdown(context.Background()); err != nil {
		s.logger.Error(err)
	}
}
