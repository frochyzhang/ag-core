package hello

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	httpSrv *http.Server
}
type Option func(s *Server)

// func NewHelloServer(engine *gin.Engine, logger *log.Logger, opts ...Option) *Server {
func NewHelloServer() *Server {
	s := &Server{}
	return s
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", 8888)
	slog.Info("hello server start", "addr", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("hello world", "addr", addr, "time", time.Now().Format("2006-01-02 15:04:05.000"))
		fmt.Fprintf(w, "Hello, World!")
	})

	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// log.Fatalf("listen: %s\n", err)
		// slog.Error("hello server", "err", err)
		slog.Info("hello server", "err", err)
		return err
	}
	return nil
}
func (s *Server) Stop(ctx context.Context) error {
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown: ", err)
	}

	slog.Info("Server exiting")
	return nil
}
