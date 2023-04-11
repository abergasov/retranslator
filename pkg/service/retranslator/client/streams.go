package client

import (
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/abergasov/retranslator/pkg/model"
	v1 "github.com/abergasov/retranslator/pkg/retranslator"
)

func (r *Service) handleCommand(stream v1.CommandStream_ListenCommandsClient) {
	r.wg.Add(1)
	defer r.wg.Done()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			r.log.Info("stream closed")
			return
		}
		if err != nil {
			r.log.Error("unable to receive command", err)
			return
		}
		//r.log.Info("received command", zap.String("url", resp.Url), zap.String("method", resp.Method))
		r.orchestra.ProcessRequest(&model.Request{
			RequestID:   resp.RequestID,
			Method:      resp.Method,
			URL:         resp.Url,
			Body:        resp.Body,
			Headers:     resp.Headers,
			OmitBody:    resp.OmitBody,
			OmitHeaders: resp.OmitHeaders,
			NewRequest:  resp.NewRequest,
		})
	}
}

func (r *Service) sendResponse(stream v1.CommandStream_ListenCommandsClient) {
	for resp := range r.orchestra.GetResponder() {
		r.countRequests(resp.StatusCode)
		//r.log.Info("send response to server", zap.Int32("status", resp.StatusCode), zap.String("request_id", resp.RequestID))
		if err := stream.Send(&v1.Response{
			RequestID:  resp.RequestID,
			StatusCode: resp.StatusCode,
			Body:       resp.Body,
			Headers:    resp.Headers,
		}); err != nil {
			r.log.Error("unable to send response", err)
		}
	}
}

func (r *Service) countRequests(statusCode int32) {
	r.requestCountsMU.Lock()
	defer r.requestCountsMU.Unlock()
	r.requestCounts[statusCode]++
}

func (r *Service) logRequests() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		r.printRequestCounts()
	}
}

func (r *Service) printRequestCounts() {
	r.requestCountsMU.Lock()
	countMap := r.requestCounts
	r.requestCounts = make(map[int32]int)
	r.requestCountsMU.Unlock()

	for k, v := range countMap {
		r.log.Info("request count", zap.Int32("status", k), zap.Int("count", v))
	}
}
