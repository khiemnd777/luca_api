package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/khiemnd777/andy_api/shared/circuitbreaker"
	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/logger"
)

// How to use:
// app.GetHttpClient().CallXxx

type HttpClient struct {
	client *http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: http.DefaultClient,
	}
}

var (
	once       sync.Once
	httpClient *HttpClient
)

func GetHttpClient() *HttpClient {
	once.Do(func() {
		httpClient = NewHttpClient()
	})
	return httpClient
}

func (c *HttpClient) CallRequestWithPort(ctx context.Context, method, module string, port int, path, token, internalToken string, body, out any, opts ...RetryOptions) error {
	srvCfg := config.Get().Server
	url := fmt.Sprintf("http://%s:%d%s", srvCfg.Host, port, path)
	reqBody, err := marshalBody(body)
	if err != nil {
		return err
	}

	retry := getRetryOptions(opts)
	var lastErr error

	for i := 0; i < retry.MaxAttempts; i++ {
		_, err := circuitbreaker.Run(fmt.Sprintf("%s:%s", module, path), func(_ context.Context) (interface{}, error) {
			log.Printf("URL: %s", url)
			req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
			if err != nil {
				return nil, err
			}

			setHeaders(req, token, internalToken)

			resp, err := c.client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 500 {
				return nil, parseErrorResponse(resp)
			}

			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				err := parseErrorResponse(resp)
				return nil, fmt.Errorf("client error: %w", err)
			}

			if out != nil {
				return nil, json.NewDecoder(resp.Body).Decode(out)
			}

			return nil, nil
		})

		if err == nil || !retry.ShouldRetry(err) {
			return err
		}

		lastErr = err
		logger.Warn(fmt.Sprintf("🔁 Retry [%s %s] #%d failed: %v", method, url, i+1, err))
		time.Sleep(retry.Delay)
	}

	logger.Error(fmt.Sprintf("❌ Request failed after retries: %s %s", method, url))
	return lastErr
}

func (c *HttpClient) CallRequest(ctx context.Context, method, module, path, token, internalToken string, body, out any, opts ...RetryOptions) error {
	srvCfg := config.Get().Server
	return c.CallRequestWithPort(ctx, method, module, srvCfg.Port, path, token, internalToken, body, out, opts...)
}

func (c *HttpClient) CallGet(ctx context.Context, module, path, token, internalToken string, out any, opts ...RetryOptions) error {
	return c.CallRequest(ctx, http.MethodGet, module, path, token, internalToken, nil, out, opts...)
}

func (c *HttpClient) CallPost(ctx context.Context, module, path, token, internalToken string, body, out any, opts ...RetryOptions) error {
	return c.CallRequest(ctx, http.MethodPost, module, path, token, internalToken, body, out, opts...)
}

func (c *HttpClient) CallPostAsync(ctx context.Context, module, path, token, internalToken string, body, out any, opts ...RetryOptions) {
	go func() {
		if err := c.CallPost(ctx, module, path, token, internalToken, body, out, opts...); err != nil {
			logger.Warn("❌ CallPostAsync failed: " + err.Error())
		}
	}()
}

// --- Helpers ---

func marshalBody(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func setHeaders(req *http.Request, token, internalToken string) {
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if internalToken != "" {
		req.Header.Set("X-Internal-Token", internalToken)
	}
}

func parseErrorResponse(resp *http.Response) error {
	if resp == nil || resp.Body == nil {
		return errors.New("empty response body")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body failed: %w", err)
	}
	return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
}

func getRetryOptions(opts []RetryOptions) RetryOptions {
	defaultRetry := RetryOptions{
		MaxAttempts: 3,
		Delay:       200 * time.Millisecond,
		ShouldRetry: func(err error) bool {
			if err == nil {
				return false
			}
			if strings.Contains(err.Error(), "client error:") {
				return false
			}
			return true
		},
	}
	if len(opts) > 0 {
		merged := opts[0]
		if merged.ShouldRetry == nil {
			merged.ShouldRetry = defaultRetry.ShouldRetry
		}
		if merged.MaxAttempts <= 0 {
			merged.MaxAttempts = defaultRetry.MaxAttempts
		}
		if merged.Delay == 0 {
			merged.Delay = defaultRetry.Delay
		}
		return merged
	}
	return defaultRetry
}
