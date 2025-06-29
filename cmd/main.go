package main

import (
	"embed"
	"fmt"
	"github.com/frochyzhang/ag-core/ag/ag_app"
	"github.com/frochyzhang/ag-core/fxs"
	"go.uber.org/fx"
	"runtime/pprof"
)

//go:embed app*
var localConfigFile embed.FS

func main() {
	threadProfile := pprof.Lookup("threadcreate")
	fmt.Printf(" beforeClient threads counts: %d\n", threadProfile.Count())
	var fxopts []fx.Option

	fxopts = append(
		fxopts,
		fx.Supply(localConfigFile),
		mainFx,
		fx.Invoke(func(s *ag_app.App) {}),
	)

	fxapp := fx.New(
		fxopts...,
	)

	fxapp.Run()

	fmt.Println("========shutdown======")
	fmt.Printf(" afterClient threads counts: %d\n", threadProfile.Count())
}

var mainFx = fx.Module("main",
	/** conf **/
	// 初始化配置
	fxs.FxAgConfModule,
	// localconf
	fxs.FxConfLocMode,
	// nacosconf
	fxs.FxConfNacoMode,
	// nettyClient
	fxs.FxNettyClientBaseModule,

	/** DB **/
	// fxs.FxAicGromdbModule,

	// 根APP
	fxs.FxAppMode,
	fxs.FxLogMode,

	/** BaseServer **/
	// Hello服务
	fxs.FxHelloServerMode,
	// HttpServerBase
	//fxs.FxHertzWithRegistryServerBaseModule,
	// KitexServerBase
	//fxs.FxKitexServerBaseModule,
	//fxs.FxNettyServerBaseModule,
)
