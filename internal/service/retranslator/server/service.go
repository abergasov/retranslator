package server

import (
	"container/list"
	"context"
	"sync"

	"github.com/abergasov/retranslator/internal/model"
	v1 "github.com/abergasov/retranslator/pkg/retranslator"
	"google.golang.org/grpc"
)

type Service struct {
	ctx              context.Context
	cancel           context.CancelFunc
	responseMapper   map[string]chan *model.Response
	responseMapperMU sync.RWMutex
	regusteredMU     sync.RWMutex
	regustered       *list.List
	v1.UnimplementedCommandStreamServer
}

func NewService() *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		ctx:            ctx,
		cancel:         cancel,
		regustered:     list.New(),
		responseMapper: make(map[string]chan *model.Response),
	}
}

func (s *Service) ListenCommands(server v1.CommandStream_ListenCommandsServer) error {
	s.regusteredMU.Lock()
	el := s.regustered.PushFront(server)
	s.regusteredMU.Unlock()
	defer func() {
		s.regusteredMU.Lock()
		s.regustered.Remove(el)
		s.regusteredMU.Unlock()
	}()

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			response, err := server.Recv()
			if err != nil {
				return err
			}
			s.responseMapperMU.RLock()
			if ch, ok := s.responseMapper[response.RequestID]; ok {
				ch <- &model.Response{
					Body:       response.Body,
					Headers:    response.Headers,
					StatusCode: response.StatusCode,
					RequestID:  response.RequestID,
				}
			}
			s.responseMapperMU.RUnlock()
		}
	}
}

func (s *Service) Start(reg grpc.ServiceRegistrar) {
	v1.RegisterCommandStreamServer(reg, s)
}

func (s *Service) Stop() {
	s.cancel()
}
