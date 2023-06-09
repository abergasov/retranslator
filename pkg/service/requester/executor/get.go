package executor

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/abergasov/retranslator/pkg/model"
)

func (s *Service) CurlLikeBrowser(ctx context.Context, request *model.Request) (*model.Response, error) {
	req, err := s.getRequest(ctx, request.URL, request.Headers["cookie"])
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	client := s.getClient(request.URL, request.Headers["cookie"])
	for header, headerVal := range request.Headers {
		if req.Header.Get(header) != headerVal {
			req.Header.Set(header, headerVal)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %w", err)
	}
	defer resp.Body.Close()
	result := &model.Response{
		RequestID:  request.RequestID,
		StatusCode: int32(resp.StatusCode),
	}
	if !request.OmitHeaders {
		result.Headers = make(map[string]string)
		for k, v := range resp.Header {
			result.Headers[k] = v[0]
		}
	}
	if request.OmitBody {
		return result, nil
	}
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
	default:
		reader = resp.Body
	}
	defer reader.Close()
	result.Body, err = io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode == 431 {
		respStr := string(result.Body)
		var res string
		if len(respStr) > 165 {
			res = respStr[0:165]
		} else {
			res = respStr
		}
		return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, res)
	}
	return result, nil
}
