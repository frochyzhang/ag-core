package ag_ext

import (
	"context"
	"log"
	"time"
)

type HandlerFunc func(ctx context.Context, req interface{}) (interface{}, error)

// Middleware 中间件类型
type Middleware func(
	method string,
	ctx context.Context,
	req interface{},
	next func(context.Context, interface{}) (interface{}, error),
) (interface{}, error)

// 注册单个方法的处理链
func registerHandler(proxy interface{}, methodName string, middlewares []Middleware, handler HandlerFunc) HandlerFunc {
	// 从后向前包装中间件
	wrappedHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		next := wrappedHandler
		wrappedHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
			return mw(methodName, ctx, req, next)
		}
	}

	return wrappedHandler
}

// LoggingMiddleware 示例：日志中间件
func LoggingMiddleware(
	method string,
	ctx context.Context,
	req interface{},
	next func(context.Context, interface{}) (interface{}, error),
) (interface{}, error) {
	start := time.Now()
	log.Printf("[%s] request received", method)

	res, err := next(ctx, req)

	if err != nil {
		log.Printf("[%s] failed in %v: %v", method, time.Since(start), err)
	} else {
		log.Printf("[%s] completed in %v", method, time.Since(start))
	}
	return res, err
}
