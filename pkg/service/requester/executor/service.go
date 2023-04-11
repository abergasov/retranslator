package executor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/abergasov/retranslator/pkg/logger"
	"github.com/abergasov/retranslator/pkg/model"
	"github.com/jellydator/ttlcache/v3"
)

type Service struct {
	log           logger.AppLogger
	clients       map[string]*http.Client
	cacheRequests *ttlcache.Cache[string, *http.Request]
	clientsMU     *sync.RWMutex
}

func NewService(log logger.AppLogger) *Service {
	ex := &Service{
		log:       log,
		clients:   make(map[string]*http.Client),
		clientsMU: &sync.RWMutex{},
		cacheRequests: ttlcache.New[string, *http.Request](
			ttlcache.WithTTL[string, *http.Request](30 * time.Minute),
		),
	}
	go ex.cacheRequests.Start()
	return ex
}

func (s *Service) Execute(request *model.Request) (*model.Response, error) {
	switch request.Method {
	case http.MethodGet:
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.CurlLikeBrowser(ctx, request)
	default:
		return nil, fmt.Errorf("method %s is not supported", request.Method)
	}
}
