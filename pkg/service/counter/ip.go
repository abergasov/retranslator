package counter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (s *Service) observeIP() {
	s.DetectIP()
	for range time.NewTicker(5 * time.Minute).C {
		s.DetectIP()
	}
}

func (s *Service) DetectIP() {
	ip, err := s.getPublicIP()
	if err != nil {
		s.log.Error("unable to get public IP", err)
		return
	}
	s.currentIPMU.Lock()
	s.currentIP = ip
	s.currentIPMU.Unlock()
}

func (s *Service) getPublicIP() (string, error) {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return "", fmt.Errorf("unable to get public IP: %w", err)
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %w", err)
	}

	type IP struct {
		Query string
	}
	var ip IP
	if err = json.Unmarshal(body, &ip); err != nil {
		return "", fmt.Errorf("unable to unmarshal response body: %w", err)
	}
	s.log.Info("got current client ip", zap.String("ip", ip.Query))
	return ip.Query, nil
}
