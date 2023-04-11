package executor

import (
	"fmt"
	"net/http"
	"time"

	cloudflarebp "github.com/DaRealFreak/cloudflare-bp-go"
	"github.com/jellydator/ttlcache/v3"
)

func (s *Service) getRequest(targetURL, cookie string) (http.Request, error) {
	key := targetURL + cookie
	item := s.cacheRequests.Get(key)
	if item != nil {
		return *item.Value(), nil
	}
	req, err := http.NewRequest(http.MethodGet, targetURL, http.NoBody)
	if err != nil {
		return http.Request{}, fmt.Errorf("failed to get url: %w", err)
	}
	s.cacheRequests.Set(key, req, ttlcache.DefaultTTL)
	return *req, err
}

func (s *Service) getClient(targetURL, cookie string) *http.Client {
	key := fmt.Sprintf("%s-%s", targetURL, cookie)
	s.clientsMU.RLock()
	client, ok := s.clients[key]
	s.clientsMU.RUnlock()
	if ok {
		return client
	}
	s.clientsMU.Lock()
	defer s.clientsMU.Unlock()

	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        30,
			MaxIdleConnsPerHost: 30,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)
	s.clients[key] = client
	return client
}
