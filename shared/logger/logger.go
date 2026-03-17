package logger

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	ServiceName  string
	Component    string
	Environment  string
	Level        string
	RedactFields []string
}

var (
	mu        sync.RWMutex
	zapLogger *zap.Logger
	options   = defaultOptions()
)

func defaultOptions() Options {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" {
		env = "development"
	}

	level := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if level == "" {
		if env == "development" {
			level = "debug"
		} else {
			level = "info"
		}
	}

	serviceName := strings.TrimSpace(os.Getenv("APP_NAME"))
	if serviceName == "" {
		serviceName = "luca_api"
	}

	return Options{
		ServiceName: serviceName,
		Component:   "app",
		Environment: env,
		Level:       level,
		RedactFields: []string{
			"authorization",
			"cookie",
			"password",
			"secret",
			"token",
			"otp",
		},
	}
}

func Init() {
	mu.Lock()
	defer mu.Unlock()
	rebuildLoggerLocked()
}

func Configure(next Options) {
	mu.Lock()
	defer mu.Unlock()

	if next.ServiceName != "" {
		options.ServiceName = next.ServiceName
	}
	if next.Component != "" {
		options.Component = next.Component
	}
	if next.Environment != "" {
		options.Environment = next.Environment
	}
	if next.Level != "" {
		options.Level = next.Level
	}
	if next.RedactFields != nil {
		options.RedactFields = append([]string(nil), next.RedactFields...)
	}

	rebuildLoggerLocked()
}

func SetComponent(component string) {
	Configure(Options{Component: component})
}

func Info(msg string, fields ...any) {
	logWithContext(context.Background(), zapcore.InfoLevel, msg, fields...)
}

func InfoContext(ctx context.Context, msg string, fields ...any) {
	logWithContext(ctx, zapcore.InfoLevel, msg, fields...)
}

func Warn(msg string, fields ...any) {
	logWithContext(context.Background(), zapcore.WarnLevel, msg, fields...)
}

func WarnContext(ctx context.Context, msg string, fields ...any) {
	logWithContext(ctx, zapcore.WarnLevel, msg, fields...)
}

func Debug(msg string, fields ...any) {
	logWithContext(context.Background(), zapcore.DebugLevel, msg, fields...)
}

func DebugContext(ctx context.Context, msg string, fields ...any) {
	logWithContext(ctx, zapcore.DebugLevel, msg, fields...)
}

func Error(msg string, fields ...any) {
	logErrorWithContext(context.Background(), msg, fields...)
}

func ErrorContext(ctx context.Context, msg string, fields ...any) {
	logErrorWithContext(ctx, msg, fields...)
}

func logErrorWithContext(ctx context.Context, msg string, fields ...any) {
	normalized := normalizeFields(fields...)
	if _, ok := normalized["stacktrace"]; !ok {
		normalized["stacktrace"] = string(debug.Stack())
	}
	logWithContext(ctx, zapcore.ErrorLevel, msg, normalized)
}

func logWithContext(ctx context.Context, level zapcore.Level, msg string, fields ...any) {
	entryFields := mergeFields(
		map[string]any{
			"service": options.ServiceName,
			"module":  options.Component,
			"env":     options.Environment,
		},
		fieldsFromContext(ctx),
		normalizeFields(fields...),
	)

	entryFields = redactFields(entryFields, options.RedactFields)

	zl := getLogger()
	switch level {
	case zapcore.DebugLevel:
		zl.Debug(msg, toZapFields(entryFields)...)
	case zapcore.WarnLevel:
		zl.Warn(msg, toZapFields(entryFields)...)
	case zapcore.ErrorLevel:
		zl.Error(msg, toZapFields(entryFields)...)
	default:
		zl.Info(msg, toZapFields(entryFields)...)
	}
}

func rebuildLoggerLocked() {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(options.Level)))); err != nil {
		level.SetLevel(zap.InfoLevel)
	}

	cfg := zap.Config{
		Level:             level,
		Development:       options.Environment == "development",
		Encoding:          "json",
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "source",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	var err error
	zapLogger, err = cfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	)
	if err != nil {
		panic(err)
	}
}

func getLogger() *zap.Logger {
	mu.RLock()
	current := zapLogger
	mu.RUnlock()
	if current != nil {
		return current
	}

	Init()

	mu.RLock()
	defer mu.RUnlock()
	return zapLogger
}

func normalizeFields(fields ...any) map[string]any {
	if len(fields) == 1 {
		if prebuilt, ok := fields[0].(map[string]any); ok {
			return cloneMap(prebuilt)
		}
	}

	out := make(map[string]any)
	extraIndex := 0
	for i := 0; i < len(fields); i++ {
		if err, ok := fields[i].(error); ok {
			out["error"] = err.Error()
			continue
		}

		key, ok := fields[i].(string)
		if !ok {
			out[fmt.Sprintf("arg_%d", extraIndex)] = fields[i]
			extraIndex++
			continue
		}

		if i+1 >= len(fields) {
			out[key] = "(missing)"
			continue
		}

		if err, ok := fields[i+1].(error); ok {
			out[key] = err.Error()
		} else {
			out[key] = fields[i+1]
		}
		i++
	}

	return out
}

func mergeFields(parts ...map[string]any) map[string]any {
	out := make(map[string]any)
	for _, part := range parts {
		for k, v := range part {
			out[k] = v
		}
	}
	return out
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func toZapFields(fields map[string]any) []zap.Field {
	out := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		out = append(out, zap.Any(k, v))
	}
	return out
}
