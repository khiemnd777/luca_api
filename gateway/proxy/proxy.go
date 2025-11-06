package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type LoadBalancer struct {
	targets []*url.URL
	alive   []bool
	counter uint64
}

// NewLoadBalancer initializes the load balancer
func NewLoadBalancer(targets []string) (*LoadBalancer, error) {
	var urls []*url.URL
	for _, t := range targets {
		u, err := url.Parse(t)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	alive := make([]bool, len(urls))
	for i := range alive {
		alive[i] = true
	}
	return &LoadBalancer{targets: urls, alive: alive}, nil
}

func (lb *LoadBalancer) NextTarget() *url.URL {
	for i := 0; i < len(lb.targets); i++ {
		index := (lb.counter + uint64(i)) % uint64(len(lb.targets))
		if lb.alive[index] {
			lb.counter = index + 1
			return lb.targets[index]
		}
	}
	// fallback nếu tất cả đều chết
	return lb.targets[0]
}

func singleJoiningSlash(a, b string) string {
	aSlash := strings.HasSuffix(a, "/")
	bSlash := strings.HasPrefix(b, "/")
	switch {
	case aSlash && bSlash:
		return a + b[1:]
	case !aSlash && !bSlash:
		return a + "/" + b
	default:
		return a + b
	}
}

// RegisterReverseProxy mounts a reverse proxy at given route with load balancing
func RegisterReverseProxy(app *fiber.App, route string, targets []string) error {
	lb, err := NewLoadBalancer(targets)
	if err != nil {
		return err
	}

	// StartHealthCheck(lb, 10*time.Second)

	app.All(route+"/*", func(c *fiber.Ctx) error {
		target := lb.NextTarget()
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = singleJoiningSlash(target.Path, c.Params("*"))
			req.Host = target.Host

			// Clean hop-by-hop headers
			if strings.ToLower(c.Get("Upgrade")) != "websocket" {
				hopHeaders := []string{
					"Connection", "Keep-Alive", "Proxy-Authenticate",
					"Proxy-Authorization", "Te", "Trailer",
					"Transfer-Encoding", "Upgrade",
				}
				for _, h := range hopHeaders {
					delete(req.Header, h)
				}
			}

			// Forward headers
			req.Header.Set("Authorization", c.Get("Authorization"))
			req.Header.Set("X-Internal-Token", utils.GetInternalToken())

			// 🪵 Log proxy action
			logger.Debug(fmt.Sprintf("[Gateway] Proxy %s → %s", c.OriginalURL(), target.String()))
		}

		fasthttpadaptor.NewFastHTTPHandler(proxy)(c.Context())
		return nil
	})

	return nil
}
