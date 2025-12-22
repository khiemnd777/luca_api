package logger

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"go.uber.org/zap"
)

var (
	useZap      = false
	zapLogger   *zap.Logger
	infoLogger  = log.New(os.Stdout, "INFO  ", log.LstdFlags)
	errorLogger = log.New(os.Stderr, "ERROR ", log.LstdFlags)
	warnLogger  = log.New(os.Stdout, "WARN  ", log.LstdFlags)
	debugLogger = log.New(os.Stdout, "DEBUG ", log.LstdFlags)
)

func Init() {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "" {
		env = "development"
	}

	if env != "development" {
		useZap = true

		cfg := zap.NewProductionConfig()

		// TẮT stacktrace mặc định của zap (vì bạn đã log tay)
		cfg.DisableStacktrace = true

		var err error
		zapLogger, err = cfg.Build(
			zap.AddCaller(),
			zap.AddCallerSkip(2),
		)
		if err != nil {
			panic(err)
		}
	}
}

// --- Logging Methods ---

func Info(msg string, fields ...any) {
	if useZap {
		zapLogger.Info(msg, convert(fields...)...)
	} else {
		infoLogger.Println(format(msg, fields...))
	}
}

func Warn(msg string, fields ...any) {
	if useZap {
		zapLogger.Warn(msg, convert(fields...)...)
	} else {
		warnLogger.Println(format(msg, fields...))
	}
}

func Debug(msg string, fields ...any) {
	if useZap {
		zapLogger.Debug(msg, convert(fields...)...)
	} else {
		debugLogger.Println(format(msg, fields...))
	}
}

func Error(msg string, fields ...any) {
	stack := debug.Stack()

	if useZap {
		zapLogger.Error(msg,
			append(convert(fields...), zap.String("stacktrace", string(stack)))...,
		)
	} else {
		errorLogger.Println(format(msg, fields...))
		errorLogger.Println("Stacktrace:\n" + string(stack))
	}
}

// --- Helpers ---

func format(msg string, fields ...any) string {
	if len(fields)%2 != 0 {
		fields = append(fields, "(missing)")
	}
	var sb strings.Builder
	sb.WriteString(msg)
	for i := 0; i < len(fields); i += 2 {
		key, _ := fields[i].(string)
		val := fields[i+1]
		sb.WriteString(fmt.Sprintf(" | %s=%v", key, val))
	}
	return sb.String()
}

func convert(fields ...any) []zap.Field {
	out := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		out = append(out, zap.Any(key, fields[i+1]))
	}
	return out
}
