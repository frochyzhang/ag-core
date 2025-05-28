package ag_log

import (
	"ag-core/ag/ag_conf"
	"os"
	"time"

	"github.com/spf13/cast"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const ctxLoggerKey = "zapLogger"

func NewZapLog(conf ag_conf.IConfigurableEnvironment) *zap.Logger {
	// log address "out.log" User-defined
	lp := conf.GetProperty("log.log_file_name")
	lv := conf.GetProperty("log.log_level")
	var level zapcore.Level
	//debug<info<warn<error<fatal<panic
	switch lv {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}
	ws := []zapcore.WriteSyncer{}
	ws = append(ws, zapcore.AddSync(os.Stdout))
	// 多套hook 对应多个zaplogger -->sloger [rpclogger,tradeinfologger,heartbealogger]
	if lp != "" {
		hook := lumberjack.Logger{
			Filename: lp, // Log file path
			// MaxSize:    ,                      // Maximum size unit for each log file: M
			// MaxBackups: conf.GetProperty("log.max_backups"), // The maximum number of backups that can be saved for log files
			// MaxAge:     conf.GetProperty("log.max_age"),     // Maximum number of days the file can be saved
			// Compress:   conf.GetProperty("log.compress"),    // Compression or not
		}
		maxsize, err := cast.ToIntE(conf.GetProperty("log.max_size"))
		if err != nil {
			panic(err)
		}
		hook.MaxSize = maxsize
		maxBackUps, err := cast.ToIntE(conf.GetProperty("log.max_backups"))
		if err != nil {
			panic(err)
		}
		hook.MaxBackups = maxBackUps
		compress, err := cast.ToBoolE(conf.GetProperty("log.compress"))
		if err != nil {
			panic(err)
		}
		hook.Compress = compress
		maxAge, err := cast.ToIntE(conf.GetProperty("log.max_age"))
		if err != nil {
			panic(err)
		}
		hook.MaxAge = maxAge
		ws = append(ws, zapcore.AddSync(&hook))
	}

	var encoder zapcore.Encoder
	if conf.GetProperty("log.encoding") == "console" {
		encoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "Logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     timeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
			//EncodeCaller: zapcore.ShortCallerEncoder,
		})
	} else {
		encoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})
	}
	core := zapcore.NewCore(
		encoder,
		//zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // Print to console and file
		zapcore.NewMultiWriteSyncer(ws...), // Print to console and file
		level,
	)
	if conf.GetProperty("app.profile") != "prod" {
		// 放入多个
		return zap.New(zapcore.NewTee(core), zap.Development(), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	}
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	//enc.AppendString(t.Format("2006-01-02 15:04:05"))
	enc.AppendString(t.Format("2006-01-02 15:04:05.000000000"))
}
