package kitex

import (
	"context"
	"fmt"
	"sort"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/server"
)

type order interface {
	Order() int
}

type IAgKitexServerMiddleware interface {
	Before(ctx context.Context, req, resp interface{}) (context.Context, error)
	After(ctx context.Context, req, resp interface{}, e error) (err error)
}

type AgKitexServerMiddleware struct {
	Middlewares []IAgKitexServerMiddleware
}

// RegistKitexServerMiddlewareOption 创建一个带有有序中间件链的 Kitex server.Option.
func RegistKitexServerMiddlewareOption(mids *AgKitexServerMiddleware) *server.Option {
	var middlewares []IAgKitexServerMiddleware
	if mids == nil || len(mids.Middlewares) == 0 {
		middlewares = make([]IAgKitexServerMiddleware, 0)
	} else {
		middlewares = make([]IAgKitexServerMiddleware, len(mids.Middlewares))
		copy(middlewares, mids.Middlewares)
	}

	// Sort middlewares by their Order() value
	sort.Slice(middlewares, func(i, j int) bool {
		var i_o, j_o int = 0, 0
		im := middlewares[i]
		jm := middlewares[j]
		if oi, ok := im.(order); ok {
			i_o = oi.Order()

		}
		if oj, ok := jm.(order); ok {
			j_o = oj.Order()
		}
		return i_o < j_o
	})

	opt := server.WithMiddleware(
		func(next endpoint.Endpoint) endpoint.Endpoint {
			return func(ctx context.Context, req, resp interface{}) (rerr error) {
				tctx := ctx
				// treq := req
				// tresp := resp

				for _, mid := range middlewares {
					var err error
					if tctx, err = mid.Before(tctx, req, resp); err != nil {
						// Before 方法出错，直接返回
						return err
					}
				}

				// Call next middleware/handler
				rerr = next(tctx, req, resp)

				for i := len(middlewares) - 1; i >= 0; i-- {
					aferr := middlewares[i].After(ctx, req, resp, rerr)
					if aferr != nil && aferr != rerr {
						rerr = fmt.Errorf("%w >>> %w", aferr, rerr)
					}
				}

				return
			}
		})

	return &opt
}

/*
func RegistKitexServerMiddleware() *server.Option {
	opt := server.WithMiddleware(func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {

			// TODO 提取设置上下文信息

			err = next(ctx, req, resp)
			// TODO 对服务端异常做必要解析判断，以供客户端识别
			if err != nil {
				slog.Debug(fmt.Sprintf("kitex errMidelware find error: %v", err))
				// err = nil
			}

			return
		}
	})

	return &opt

}
*/
