package log

import (
	"context"

	"github.com/si9ma/KillOJ-common/utils"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogEncoding string

const Json LogEncoding = "json"
const Console LogEncoding = "console"

var zapLogger *zap.Logger

// auto init
func init() {
	_ = Init([]string{}, Json)
}

// manual init
func Init(outputPaths []string, encode LogEncoding) (err error) {
	// write to stdout when output path is empty
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	// on debug mode, write log to stdout at the same time
	if utils.IsDebug() && !utils.ContainsString(outputPaths, "stdout") {
		outputPaths = append(outputPaths, "stdout")
	}

	// config format
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	if encode == Console {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg := zap.NewProductionConfig()
	cfg.Encoding = string(encode)
	cfg.EncoderConfig = encoderCfg
	cfg.OutputPaths = outputPaths // set output paths
	if zapLogger, err = cfg.Build(); err != nil {
		// if build fail , set zapLogger as a default logger
		// todo There may be a bug here
		zapLogger, _ = zap.NewProduction()
	}
	return
}

// get the normal Logger without write span log
func Bg() Logger {
	return logger{logger: zapLogger}
}

// if the context contains an span,
// return a spanLogger
func For(ctx context.Context) Logger {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		return spanLogger{span: span, logger: zapLogger}
	}

	return Bg()
}
