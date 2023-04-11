package model

type Request struct {
	RequestID   string            `json:"request_id"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Body        []byte            `json:"body"`
	Headers     map[string]string `json:"headers"`
	OmitBody    bool              `json:"omit_body"`
	OmitHeaders bool              `json:"omit_headers"`
	NewRequest  bool              `json:"new_request"`
}
