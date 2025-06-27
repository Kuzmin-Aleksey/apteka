package server

import (
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"server/internal/config"
	"time"
)

type Logger interface {
	Println(v ...any)
}

type HttpServer struct {
	server *http.Server
	l      Logger
	cfg    *config.HttpConfig
}

func CreateHttpServer(handler http.Handler, l Logger, cfg *config.HttpConfig) *HttpServer {
	return &HttpServer{
		server: &http.Server{
			Addr:         cfg.Address,
			Handler:      handler,
			ReadTimeout:  time.Duration(cfg.ReadTimeoutSec) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeoutSec) * time.Second,
			ErrorLog:     log.New(os.Stderr, "HTTP server: ", log.LstdFlags),
		},
		l:   l,
		cfg: cfg,
	}
}

func (s *HttpServer) ListenAndServe() error {
	if len(s.cfg.SSLKeyPath) != 0 && len(s.cfg.SSLCertPath) != 0 {
		s.l.Println("Starting HTTPS server on", s.server.Addr)
		if err := s.server.ListenAndServeTLS(s.cfg.SSLCertPath, s.cfg.SSLKeyPath); err != nil {
			return err
		}
		return nil
	}
	s.l.Println("Starting HTTP server on", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
