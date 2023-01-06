package http

import (
	"context"
	"errors"
	"log"
	"net/http"
)

type Server struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	server      *http.Server
}

func New(port string, handler http.Handler, infoLogger *log.Logger, errorLogger *log.Logger) *Server {
	return &Server{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
	}
}

func (s *Server) Start() {
	s.infoLogger.Println("server started")
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.errorLogger.Println(err)
	}
}

func (s *Server) Stop() {
	s.infoLogger.Println("server shutting down")
	if err := s.server.Shutdown(context.Background()); err != nil {
		s.errorLogger.Println(err)
	}
}
