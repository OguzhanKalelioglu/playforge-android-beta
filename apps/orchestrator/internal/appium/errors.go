package appium

import (
	"errors"
	"fmt"
)

// Typed errors — task'lar bu error'ları inspect edebilir
var (
	ErrSessionLost      = errors.New("appium session lost")
	ErrSessionNotActive = errors.New("appium session not active")
	ErrElementNotFound  = errors.New("element not found")
	ErrTimeout          = errors.New("operation timed out")
	ErrServerUnreachable = errors.New("appium server unreachable")
)

// W3CError, Appium server'dan dönen structured error
type W3CError struct {
	Code    string                 `json:"error"`
	Message string                 `json:"message"`
	Stacktrace string              `json:"stacktrace,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	HTTPStatus int                 `json:"-"`
}

func (e *W3CError) Error() string {
	return fmt.Sprintf("appium error [%s] %s: %s", e.Code, e.HTTPStatus, e.Message)
}

func (e *W3CError) IsSessionLost() bool {
	return e.Code == "invalid session id" || e.Code == "no such driver"
}
