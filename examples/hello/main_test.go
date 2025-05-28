package hello

import (
	"ag-core/ag/ag_app"
	"ag-core/fxs"
	"embed"
	"fmt"
	"testing"

	"go.uber.org/fx"
)

//go:embed app*
var localConfigFile embed.FS

// func main() {
// 	FX_RUN(localConfigFile)
// }

func TestMain(t *testing.T) {
	FX_RUN(localConfigFile)
}

func FX_RUN(fs embed.FS) {
	var fxopts []fx.Option

	fxopts = append(
		fxopts,
		fx.Supply(fs),
		mainFx,
		fx.Invoke(func(s *ag_app.App) {}),
	)

	fxapp := fx.New(
		fxopts...,
	)

	fxapp.Run()

	fmt.Println("========shutdown======")
}

var mainFx = fx.Module("main",
	/** conf **/
	// 初始化配置
	fxs.FxAgConfModule,
	// localconf
	fxs.FxConfLocMode,
	// nacosconf
	// fxs.FxConfNacoMode,

	/** DB **/
	// fxs.FxAicGromdbModule,

	// 根APP
	fxs.FxAppMode,
	fxs.FxLogMode,

	/** BaseServer **/
	// Hello服务
	fxs.FxHelloServerMode,
	// HttpServerBase
	// fxs.FxHttpServerBaseModule,
	// KitexServerBase
	// fxs.FxKitexServerBaseModule,

)
