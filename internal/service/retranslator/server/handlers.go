package server

import (
	"errors"
	"retranslator/internal/model"
	v1 "retranslator/pkg/retranslator"
)

func (s *Service) ProxyRequest(requestID, method, url string, body []byte, omitBody, omitHeaders bool) (<-chan *model.Response, error) {
	s.responseMapperMU.Lock()
	if _, ok := s.responseMapper[requestID]; !ok {
		s.responseMapper[requestID] = make(chan *model.Response, 1_000)
	}
	res := s.responseMapper[requestID]
	s.responseMapperMU.Unlock()
	request := &v1.Request{
		RequestID:   requestID,
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
