package server

import (
	"context"
	"errors"
	"github.com/frochyzhang/ag-core/ag/ag_error"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func serverErrMidleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		err = next(ctx, req, resp)
		if err != nil {
			var abe ag_error.BizStatusErrorIface
			// 如果是业务异常则转换成kitex的BizStatusError
			if errors.As(err, &abe) {
				// 提取rpcinfo
				ri := rpcinfo.GetRPCInfo(ctx)
				if setter, ok := ri.Invocation().(rpcinfo.InvocationSetter); ok {
					kbe := kerrors.NewBizStatusErrorWithExtra(abe.BizCode(), abe.BizMessage(), abe.BizExtra())
					setter.SetBizStatusErr(kbe)
					return nil
				}
			}
		}

		return err
	}
}

func NewAgBizErrorMiddlewareOption() *server.Option {
	opt := server.WithMiddleware(serverErrMidleware)
	return &opt
}
