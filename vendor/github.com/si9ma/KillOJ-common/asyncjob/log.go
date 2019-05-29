package asyncjob

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/utils"

	"github.com/si9ma/KillOJ-common/log"
)

type ZLogger struct{}

func (l ZLogger) Print(args ...interface{}) {
}

// we just need log about receiving message
func (l ZLogger) Printf(s string, args ...interface{}) {
	// TODO There may be a bug here
	if s == "Received new message: %s" {
		log.Bg().Info("received new message", zap.String("message", fmt.Sprintf("%s", args...)))
	}
}

func (l ZLogger) Println(args ...interface{}) {
}

func (l ZLogger) Fatal(args ...interface{}) {
}

func (l ZLogger) Fatalf(s string, args ...interface{}) {
}

func (l ZLogger) Fatalln(args ...interface{}) {
}

func (l ZLogger) Panic(args ...interface{}) {
}

func (l ZLogger) Panicf(s string, args ...interface{}) {
}

func (l ZLogger) Panicln(args ...interface{}) {
}

type DebugLogger struct{}

func (d DebugLogger) Print(args ...interface{}) {
}

func (d DebugLogger) Printf(s string, args ...interface{}) {
	if utils.IsDebug() {
		ZLogger{}.Printf(s, args...)
	}
}

func (d DebugLogger) Println(args ...interface{}) {
}

func (d DebugLogger) Fatal(args ...interface{}) {
}

func (d DebugLogger) Fatalf(s string, args ...interface{}) {
}

func (d DebugLogger) Fatalln(args ...interface{}) {
}

func (d DebugLogger) Panic(args ...interface{}) {
}

func (d DebugLogger) Panicf(s string, args ...interface{}) {
}

func (d DebugLogger) Panicln(args ...interface{}) {
}
