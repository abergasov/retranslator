package client

import (
	"io"

	"github.com/abergasov/retranslator/internal/model"
	v1 "github.com/abergasov/retranslator/pkg/retranslator"
	"go.uber.org/zap"
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
		r.log.Info("received command", zap.String("url", resp.Url), zap.String("method", resp.Method))
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
		r.log.Info("send response to server")
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
