package ag_app

import (
	"context"
	"fmt"
	"github.com/frochyzhang/ag-core/ag/ag_server"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	AppNameKey string = "appName"
)

type App struct {
	name    string
	Servers []ag_server.Server
	Logger  *slog.Logger
	cancel  context.CancelFunc
	ctx     context.Context
	// startTimeout time.Duration // 服务启动超时时间
}

type Option func(a *App)

func NewApp(opts ...Option) (*App, func()) {
	a := &App{}
	for _, opt := range opts {
		opt(a)
	}
	cleanup := func() {
		fmt.Printf("wire cleanup!!!!!!")
		time.Sleep(time.Second)
	}
	return a, cleanup
}

func WithServer(servers ...ag_server.Server) Option {
	return func(a *App) {
		a.Servers = servers
	}
}

func WithName(name string) Option {
	return func(a *App) {
		a.name = name
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(a *App) {
		a.Logger = logger
	}
}
func (a *App) Run(ctx context.Context) error {

	// Start
	a.Start(ctx)

	// Wait
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signals:
		// Received termination signal
		a.Logger.Info("Received termination signal")
	case <-ctx.Done():
		// Context canceled
		//log.Println("Context canceled")
		a.Logger.Info("Context canceled")
	}

	a.Stop(ctx)

	return nil
}

// func (a *App) Start(ctx context.Context) error {
// 	vctx := context.WithValue(ctx, AppNameKey, a.name)

// 	cctx, cancel := context.WithCancel(vctx)
// 	a.cancel = cancel
// 	a.ctx = cctx

// 	g, gctx := errgroup.WithContext(a.ctx)
// 	var startErrs []error

// 	for _, srv := range a.servers {
// 		srv := srv // capture range variable
// 		g.Go(func() error {
// 			err := srv.Start(gctx)
// 			if err != nil {
// 				a.Logger.Error("Server start failed", "error", err)
// 				startErrs = append(startErrs, err)
// 				return err
// 			}
// 			return nil
// 		})
// 	}

// 	if err := g.Wait(); err != nil { // TODO 此处的g.Wait()会一直阻塞，无法返回
// 		rerr := fmt.Errorf("one or more servers failed to start: %v", errors.Join(startErrs...))
// 		slog.Error("", "err", rerr)
// 		return rerr
// 	}
// 	slog.Info("servers started")

// 	return nil
// }

func (a *App) Start(ctx context.Context) error {
	vctx := context.WithValue(ctx, AppNameKey, a.name)

	cctx, cancel := context.WithCancel(vctx)
	a.cancel = cancel
	a.ctx = cctx

	for _, srv := range a.Servers {
		srv := srv // capture range variable
		go func() {
			err := srv.Start(a.ctx)
			if err != nil {
				a.Logger.Error("Server start failed", "error", err)
				log.Fatal(err) // TODO 服务启动失败暂强制退出
			}
		}()
	}
	// TODO 如何优雅的捕获返回 servers 启动的错误，以最终确定APP是否启动成功

	slog.Info("servers started")
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	defer a.cancel()

	var rerr error
	// Gracefully stop the servers
	for _, srv := range a.Servers {
		err := srv.Stop(a.ctx) // TODO 此处的 ctx 应该用哪个
		if err != nil {
			rerr = fmt.Errorf("%w,%w", rerr, err)
			//log.Printf("Server stop err: %v", err)
			// a.Logger.Info(fmt.Sprintf("Server stop err: %v", err))
			slog.Info(fmt.Sprintf("Server stop err: %v", err))
		}
	}

	return rerr
}
