package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

func (s *Service) getRequest(ctx context.Context, targetURL, cookie string) (*http.Request, error) {
	key := targetURL + cookie
	item := s.cacheRequests.Get(key)
	if item != nil {
		return item.Value().WithContext(ctx), nil
	}
	req, err := http.NewRequest(http.MethodGet, targetURL, http.NoBody)
	if err != nil {
		return &http.Request{}, fmt.Errorf("failed to get url: %w", err)
	}
	s.cacheRequests.Set(key, req, ttlcache.DefaultTTL)
	return req.WithContext(ctx), err
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
			TLSClientConfig: &tls.Config{
				CurvePreferences: []tls.CurveID{tls.CurveP256, tls.CurveP384, tls.CurveP521, tls.X25519},
				CipherSuites: []uint16{
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
			},
			MaxIdleConns:        30,
			MaxIdleConnsPerHost: 30,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	// client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)
	s.clients[key] = client
	return client
}
