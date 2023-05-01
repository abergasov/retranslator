package client

import (
	"context"
	"sync"
	"time"

	"github.com/abergasov/retranslator/pkg/logger"
	"github.com/abergasov/retranslator/pkg/model"
	v1 "github.com/abergasov/retranslator/pkg/retranslator"
	"github.com/abergasov/retranslator/pkg/service/counter"
	"github.com/abergasov/retranslator/pkg/service/requester/orchestrator"
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
	requestCounter  *counter.Service
	counterTicker   *time.Ticker
}

func NewRelay(log logger.AppLogger, host string, service *orchestrator.Service, requestCounter *counter.Service) *Service {
	srv := &Service{
		wg:             &sync.WaitGroup{},
		log:            log,
		targetHost:     host,
		responses:      make(chan *model.Response, 1_000),
		orchestra:      service,
		requestCounts:  map[int32]int{},
		requestCounter: requestCounter,
	}
	return srv
}

func (r *Service) Start() {
	r.log.Info("starting relay")
	r.counterTicker = time.NewTicker(1 * time.Second)
	go r.logRequests()
	r.ctx, r.cancel = context.WithCancel(context.Background())
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
	defer conn.Close()
	client := v1.NewCommandStreamClient(conn)

	stream, err := client.ListenCommands(r.ctx)
	if err != nil {
		r.log.Error("unable to listen commands", err)
		return
	}
	r.wg.Add(1)
	go r.handleCommand(stream)
	go r.sendResponse(stream)
	r.wg.Wait()
}

func (r *Service) Stop() {
	r.log.Info("stopping relay")
	r.counterTicker.Stop()
	r.cancel()
	r.orchestra.Stop()
	r.wg.Wait()
	r.log.Info("relay stopped")
}
