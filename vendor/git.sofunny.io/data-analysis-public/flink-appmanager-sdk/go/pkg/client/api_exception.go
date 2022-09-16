package client

import (
	"fmt"
)

type ApiException struct {
	Model
	Message    string            `json:"message,omitempty"`
	Reason     string            `json:"reason,omitempty"`
	StatusCode int64             `json:"statusCode,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
}

// ExceptionFormat 格式化返回的异常
func (a *ApiException) ExceptionFormat() string {
	if em, ok := a.Context["exceptionMessage"]; ok {
		return fmt.Sprintf("%s: %s", a.Message, em)
	}
	return a.Message
}
