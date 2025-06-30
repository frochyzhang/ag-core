{{$ifcName := .Name}}
{{$lowerIfcName := toLower .Name}}
package service

import (
	"context"
	"log"
	"time"

	pb "{{.PbPkg}}"
	mw "github.com/frochyzhang/ag-core/ag/ag_ext"
)

// ===================== 接口定义 =====================

// 代理接口
type {{$ifcName}}Proxy interface {
    pb.{{$ifcName}}Server
	AddMiddleware(m mw.Middleware)
}

// ===================== 代理实现 =====================

type {{toLower .Name}}ProxyImpl struct {
	service      interface{} // 原始服务实例
	middlewares  []mw.Middleware
}

func New{{$ifcName}}Proxy(service *{{$ifcName}}Service) {{$ifcName}}Proxy {
	return &{{$lowerIfcName}}ProxyImpl{
		service:     service,
		middlewares: make([]mw.Middleware, 0),
	}
}

func (p *{{$lowerIfcName}}ProxyImpl) AddMiddleware(m mw.Middleware) {
	p.middlewares = append(p.middlewares, m)
}

// ======== {{.Name}} 代理方法 ========
{{range .Methods}}
func (p *{{$lowerIfcName}}ProxyImpl) {{.Name}}(ctx context.Context, in *pb.{{.Request}}) (*pb.{{.Reply}}, error) {
    start := time.Now()
    methodName := "{{.Name}}"

    // 创建处理链
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        // 最终调用原始服务方法
        s := p.service.(pb.{{$ifcName}}Server)
        return s.{{.Name}}(ctx, req.(*pb.{{.Request}}))
    }

    // 应用中间件
    for i := len(p.middlewares) - 1; i >= 0; i-- {
        mw := p.middlewares[i]
        next := handler
        handler = func(ctx context.Context, req interface{}) (interface{}, error) {
            return mw(methodName, ctx, req, next)
        }
    }

    // 执行调用链
    res, err := handler(ctx, in)
    if err != nil {
        log.Printf("[%s] failed in %v: %v", methodName, time.Since(start), err)
        return nil, err
    }

    log.Printf("[%s] success in %v", methodName, time.Since(start))
    return res.(*pb.{{.Reply}}), nil
}{{end}}