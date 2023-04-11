package server

import (
	"errors"

	"github.com/abergasov/retranslator/pkg/model"
	v1 "github.com/abergasov/retranslator/pkg/retranslator"
)

func (s *Service) ProxyRequest(requestID, method, url string, headers map[string]string, body []byte, omitBody, omitHeaders bool) (<-chan *model.Response, error) {
	s.responseMapperMU.Lock()
	if _, ok := s.responseMapper[requestID]; !ok {
		s.responseMapper[requestID] = make(chan *model.Response, 1_000)
	}
	res := s.responseMapper[requestID]
	s.responseMapperMU.Unlock()
	request := &v1.Request{
		RequestID:   requestID,
		Headers:     headers,
		Method:      method,
		Url:         url,
		Body:        body,
		OmitHeaders: omitHeaders,
		OmitBody:    omitBody,
	}
	s.regusteredMU.Lock()
	defer s.regusteredMU.Unlock()

	for e := s.regustered.Front(); e != nil; e = e.Next() {
		pr := e.Value.(v1.CommandStream_ListenCommandsServer)
		if err := pr.Send(request); err != nil {
			// probably some issues, try next one
			continue
		}
		front := s.regustered.Front()
		s.regustered.MoveToBack(front)
		return res, nil
	}
	return nil, errors.New("no streams available")
}
