package ag_ext

import (
	"context"
	"log"
	"time"
)

// Middleware 中间件类型
type Middleware func(
	method string,
	ctx context.Context,
	req interface{},
	next func(context.Context, interface{}) (interface{}, error),
) (interface{}, error)

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
