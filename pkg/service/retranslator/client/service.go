package client

import (
	"context"
	"sync"
	"time"

	"github.com/abergasov/retranslator/pkg/logger"
	"github.com/abergasov/retranslator/pkg/model"
	v1 "github.com/abergasov/retranslator/pkg/retranslator"
	"github.com/abergasov/retranslator/pkg/service/requester/orchestrator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Service struct {
	wg              *sync.WaitGroup
	targetHost      string
	log             logger.AppLogger
	ctx             context.Context
	cancel          context.CancelFunc
	responses       chan *model.Response
	orchestra       *orchestrator.Service
	requestCounts   map[int32]int
	requestCountsMU sync.Mutex
}

func NewRelay(log logger.AppLogger, host string, service *orchestrator.Service) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	srv := &Service{
		wg:            &sync.WaitGroup{},
		log:           log.With(zap.String("host", host)),
		targetHost:    host,
		ctx:           ctx,
		cancel:        cancel,
		responses:     make(chan *model.Response, 1_000),
		orchestra:     service,
		requestCounts: map[int32]int{},
	}
	go srv.logRequests()
	return srv
}

func (r *Service) Start() {
	r.log.Info("starting relay")
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			default:
				r.processConnection()
				time.Sleep(5 * time.Second) // probably it broken connection
			}
		}
	}()
}

func (r *Service) GetResponder() chan<- *model.Response {
	return r.responses
}

func (r *Service) processConnection() {
	r.log.Info("connecting to target host")
	conn, err := grpc.Dial(r.targetHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		r.log.Error("unable to connect to target host", err)
		return
	}
	client := v1.NewCommandStreamClient(conn)

	stream, err := client.ListenCommands(r.ctx)
	if err != nil {
		r.log.Error("unable to listen commands", err)
		return
	}
	go r.handleCommand(stream)
	go r.sendResponse(stream)
	r.wg.Wait()
}

func (r *Service) Stop() {
	r.cancel()
	r.orchestra.Stop()
	r.log.Info("relay stopped")
	r.wg.Wait()
}
