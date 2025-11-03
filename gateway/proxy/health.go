package proxy

import (
	"net/http"
	"time"
)

func StartHealthCheck(lb *LoadBalancer, interval time.Duration) {
	go func() {
		for {
			for i, target := range lb.targets {
				resp, err := http.Get(target.String() + "/health")
				if err != nil || resp.StatusCode != http.StatusOK {
					lb.alive[i] = false
				} else {
					lb.alive[i] = true
					resp.Body.Close()
				}
			}
			time.Sleep(interval)
		}
	}()
}
