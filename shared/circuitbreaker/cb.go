package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/sony/gobreaker"
)

var breaker *gobreaker.CircuitBreaker

var ErrClientResponse = errors.New("client response")

func Init() {
	cbCfg := config.Get().CircuitBreaker
	settings := gobreaker.Settings{
		Name:     "GlobalCircuitBreaker",
		Interval: cbCfg.Interval, // Thời gian reset lại các thống kê
		Timeout:  cbCfg.Timeout,  // Sau khi mở (Open), đợi 10s rồi thử lại
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(cbCfg.ConsecutiveFailures)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("🔌 Circuit state changed: %s → %s", from.String(), to.String())
		},
	}
	breaker = gobreaker.NewCircuitBreaker(settings)
	log.Println("🛡️ Circuit Breaker (gobreaker) initialized")
}

func Run(name string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	var panicErr error

	result, err := breaker.Execute(func() (interface{}, error) {
		defer func() {
			if r := recover(); r != nil {
				panicErr = fmt.Errorf("panic: %v", r)
				logger.Error(fmt.Sprintf("🔥 Panic in circuit breaker [%s]: %v", name, r))
				logger.Debug(string(debug.Stack()))
			}
		}()
		return fn(context.Background())
	})

	if panicErr != nil {
		// logger.Error(fmt.Sprintf("❌ Circuit Panic Error on [%s]: %v", name, panicErr))
		return nil, panicErr
	}

	switch {
	case errors.Is(err, gobreaker.ErrOpenState):
		logger.Warn("🚫 Circuit Open: blocked call [" + name + "]")
		return nil, err
	case errors.Is(err, ErrClientResponse):
		logger.Warn("⚠️ Client response on [" + name + "]")
		return nil, err
	case err != nil:
		if strings.Contains(err.Error(), "client error") {
			logger.Warn(fmt.Sprintf("⚠️ Client error on [%s]: %v", name, err))
			return nil, err
		}
		// logger.Error(fmt.Sprintf("❌ Circuit Error on [%s]:%v", name, err))
		return nil, err
	default:
		logger.Info("✅ Circuit call success: " + name)
		return result, nil
	}
}
