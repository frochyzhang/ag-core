package hello

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/frochyzhang/ag-core/ag/ag_netty/client"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	httpSrv *http.Server
	suite   *client.NettyOptionSuite
	logger  *slog.Logger
}
type Option func(s *Server)

func NewHelloServer(
	suite *client.NettyOptionSuite,
	logger *slog.Logger,
) *Server {
	s := &Server{
		suite:  suite,
		logger: logger,
	}
	return s
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", "0.0.0.0", 8888)
	slog.Info("hello server start", "addr", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			fmt.Printf("%s:%s\n", k, v)
		}

		body := r.Body
		buf := make([]byte, 2048)
		n, _ := body.Read(buf)

		clientWithSuite := client.NewNettyClientWithSuite(s.suite, s.logger)
		err2 := clientWithSuite.Connect()
		if err2 != nil {
			slog.Error("netty client connect failed", "error", err2)
			http.Error(w, "connect failed", http.StatusInternalServerError)
			return
		}
		defer clientWithSuite.Close()

		_, err2 = clientWithSuite.SendAndGet(buf[:n])
		if err2 != nil {
			slog.Error("netty client sendAndGet failed", "error", err2)
			http.Error(w, "send failed", http.StatusInternalServerError)
			return
		}

		bbuf := buf[:n]
		var bmap map[string]any
		err := json.Unmarshal(bbuf, &bmap)
		if err != nil {
			slog.Error("unmarshal", "err", err)
		} else {
			bijson, _ := json.MarshalIndent(bmap, " ", " ")
			slog.Info(fmt.Sprintf("%s\n", bijson))
		}

		slog.Info("hello world", "addr", addr, "time", time.Now().Format("2006-01-02 15:04:05.000"))
		fmt.Fprintf(w, "Hello, World!")
	})

	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Info("hello server", "err", err)
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	slog.Info("Shutting down server...")
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	slog.Info("Server exiting")
	return nil
}
