package client

import (
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/abergasov/retranslator/pkg/model"
	v1 "github.com/abergasov/retranslator/pkg/retranslator"
)

func (r *Service) handleCommand(stream v1.CommandStream_ListenCommandsClient) {
	defer r.wg.Done()
	r.requestCounter.DetectIP()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-stream.Context().Done():
			return
		default:
			resp, err := stream.Recv()
			if err == io.EOF {
				r.log.Info("stream closed")
				return
			}
			if err != nil {
				r.log.Error("unable to receive command", err)
				return
			}
			if err = r.requestCounter.CanRequest(); err != nil {
				r.log.Error("unable to process request", err)
				r.Stop()
				time.Sleep(15 * time.Minute)
				os.Exit(0)
				return
			}
			go r.orchestra.ProcessRequest(&model.Request{
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
}

func (r *Service) sendResponse(stream v1.CommandStream_ListenCommandsClient) {
	for resp := range r.orchestra.GetResponder() {
		select {
		case <-r.ctx.Done():
			return
		case <-stream.Context().Done():
			return
		default:
			r.countRequests(resp.StatusCode)
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
}

func (r *Service) countRequests(statusCode int32) {
	r.requestCountsMU.Lock()
	defer r.requestCountsMU.Unlock()
	r.requestCounts[statusCode]++
}

func (r *Service) logRequests() {
	for range r.counterTicker.C {
		if time.Now().Second() == 0 {
			r.printRequestCounts()
		}
	}
}

func (r *Service) printRequestCounts() {
	r.requestCountsMU.Lock()
	countMap := r.requestCounts
	r.requestCounts = make(map[int32]int)
	r.requestCountsMU.Unlock()

	okCount := 0
	nonOkCount := 0
	for k, v := range countMap {
		if k == http.StatusOK {
			okCount = v
		} else {
			nonOkCount += v
		}
		r.log.Info("request count", zap.Int32("status", k), zap.Int("count", v))
	}

	if nonOkCount+okCount == 0 {
		return
	}
	percentage := (nonOkCount / (nonOkCount + okCount)) * 100
	if percentage > 30 {
		r.Stop()
		r.log.Info("bad requests more than limit", zap.Int("count", nonOkCount), zap.Int("ok_count", okCount), zap.Int("percentage", percentage))
		time.Sleep(15 * time.Minute)
		os.Exit(0)
	}
}
