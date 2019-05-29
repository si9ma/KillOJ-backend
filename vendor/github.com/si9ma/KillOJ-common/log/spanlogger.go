package log

import (
	"time"

	"github.com/opentracing/opentracing-go"
	tag "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	tlog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// warp span logger and zap logger
// write span log and zap log simultaneously
type spanLogger struct {
	logger *zap.Logger
	span   opentracing.Span
}

func (sl spanLogger) Info(msg string, fields ...zapcore.Field) {
	sl.logToSpan("INFO", msg, fields...)
	sl.logger.Info(msg, fields...)
}

func (sl spanLogger) Error(msg string, fields ...zapcore.Field) {
	sl.logToSpan("ERROR", msg, fields...)
	tag.Error.Set(sl.span, true) // set error tag
	sl.logger.Error(msg, fields...)
}

func (sl spanLogger) Warn(msg string, fields ...zapcore.Field) {
	sl.logToSpan("WARN", msg, fields...)
	sl.logger.Warn(msg, fields...)
}

func (sl spanLogger) Fatal(msg string, fields ...zapcore.Field) {
	sl.logToSpan("FATAL", msg, fields...)
	tag.Error.Set(sl.span, true) // set error tag
	sl.logger.Fatal(msg, fields...)
}

func (sl spanLogger) Debug(msg string, fields ...zapcore.Field) {
	sl.logToSpan("DEBUG", msg, fields...)
	sl.logger.Fatal(msg, fields...)
}

func (sl spanLogger) Panic(msg string, fields ...zapcore.Field) {
	sl.logToSpan("PANIC", msg, fields...)
	tag.Error.Set(sl.span, true) // set error tag
	sl.logger.Fatal(msg, fields...)
}

// write span log
func (sl spanLogger) logToSpan(level, msg string, fields ...zapcore.Field) {
	fa := fieldAdapter(make([]log.Field, 0, 2+len(fields)))
	fa = append(fa, log.String("event", msg))
	fa = append(fa, log.String("level", level))
	for _, field := range fields {
		field.AddTo(&fa)
	}
	sl.span.LogFields(fa...)
}

// convert zapcore Fields to opentracing log Fields
type fieldAdapter []tlog.Field

func (fa *fieldAdapter) AddBool(key string, value bool) {
	*fa = append(*fa, tlog.Bool(key, value))
}

func (fa *fieldAdapter) AddFloat64(key string, value float64) {
	*fa = append(*fa, tlog.Float64(key, value))
}

func (fa *fieldAdapter) AddFloat32(key string, value float32) {
	*fa = append(*fa, tlog.Float64(key, float64(value)))
}

func (fa *fieldAdapter) AddInt(key string, value int) {
	*fa = append(*fa, tlog.Int(key, value))
}

func (fa *fieldAdapter) AddInt64(key string, value int64) {
	*fa = append(*fa, tlog.Int64(key, value))
}

func (fa *fieldAdapter) AddInt32(key string, value int32) {
	*fa = append(*fa, tlog.Int64(key, int64(value)))
}

func (fa *fieldAdapter) AddInt16(key string, value int16) {
	*fa = append(*fa, tlog.Int64(key, int64(value)))
}

func (fa *fieldAdapter) AddInt8(key string, value int8) {
	*fa = append(*fa, tlog.Int64(key, int64(value)))
}

func (fa *fieldAdapter) AddUint(key string, value uint) {
	*fa = append(*fa, tlog.Uint64(key, uint64(value)))
}

func (fa *fieldAdapter) AddUint64(key string, value uint64) {
	*fa = append(*fa, tlog.Uint64(key, value))
}

func (fa *fieldAdapter) AddUint32(key string, value uint32) {
	*fa = append(*fa, tlog.Uint64(key, uint64(value)))
}

func (fa *fieldAdapter) AddUint16(key string, value uint16) {
	*fa = append(*fa, tlog.Uint64(key, uint64(value)))
}

func (fa *fieldAdapter) AddUint8(key string, value uint8) {
	*fa = append(*fa, tlog.Uint64(key, uint64(value)))
}

func (fa *fieldAdapter) AddUintptr(key string, value uintptr)                        {}
func (fa *fieldAdapter) AddArray(key string, marshaler zapcore.ArrayMarshaler) error { return nil }
func (fa *fieldAdapter) AddComplex128(key string, value complex128)                  {}
func (fa *fieldAdapter) AddComplex64(key string, value complex64)                    {}
func (fa *fieldAdapter) AddObject(key string, value zapcore.ObjectMarshaler) error   { return nil }
func (fa *fieldAdapter) AddReflected(key string, value interface{}) error            { return nil }
func (fa *fieldAdapter) OpenNamespace(key string)                                    {}

func (fa *fieldAdapter) AddDuration(key string, value time.Duration) {
	*fa = append(*fa, tlog.String(key, value.String()))
}

func (fa *fieldAdapter) AddTime(key string, value time.Time) {
	*fa = append(*fa, tlog.String(key, value.String()))
}

func (fa *fieldAdapter) AddBinary(key string, value []byte) {
	*fa = append(*fa, tlog.Object(key, value))
}

func (fa *fieldAdapter) AddByteString(key string, value []byte) {
	*fa = append(*fa, tlog.Object(key, value))
}

func (fa *fieldAdapter) AddString(key, value string) {
	if key != "" && value != "" {
		*fa = append(*fa, tlog.String(key, value))
	}
}
