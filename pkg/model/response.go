package model

type Response struct {
	RequestID  string            `json:"request_id"`
	StatusCode int32             `json:"status_code"`
	Body       []byte            `json:"body"`
	Headers    map[string]string `json:"headers"`
}
