package orchestrator

import (
	"context"

	"github.com/abergasov/retranslator/pkg/logger"
	"github.com/abergasov/retranslator/pkg/model"
	"github.com/abergasov/retranslator/pkg/service/requester/executor"
)

type Service struct {
	log       logger.AppLogger
	responser chan *model.Response
	service   *executor.Service
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewService(log logger.AppLogger, service *executor.Service) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		log:       log,
		ctx:       ctx,
		cancel:    cancel,
		service:   service,
		responser: make(chan *model.Response, 1_000),
	}
}

func (s *Service) GetResponder() <-chan *model.Response {
	return s.responser
}

func (s *Service) Stop() {
	s.log.Info("stopping orchestrator")
	s.cancel()
}

func (s *Service) ProcessRequest(request *model.Request) {
	response, err := s.service.Execute(request)
	if err != nil {
		s.log.Error("unable to execute request", err)
		return
	}
	s.responser <- response
}
