package appium

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Session, bir WebDriver session'ı temsil eder
// Tüm UI işlemleri (tap, swipe, vs.) bu session üzerinden yapılır
type Session struct {
	client    *Client
	sessionID string
	active    bool
	mu        sync.RWMutex

	createdAt time.Time
}

// ID, session ID'sini döndürür
func (s *Session) ID() string {
	return s.sessionID
}

// IsActive, session'ın hâlâ aktif olup olmadığını kontrol eder
func (s *Session) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// setActive, internal state değiştirici
func (s *Session) setActive(v bool) {
	s.mu.Lock()
	s.active = v
	s.mu.Unlock()
}

// Quit, session'ı temiz bir şekilde sonlandırır
// Hata olsa bile birden fazla kez çağrılabilir (idempotent)
func (s *Session) Quit(ctx context.Context) error {
	if !s.IsActive() {
		return nil
	}

	err := s.client.do(ctx, http.MethodDelete, "/session/"+s.sessionID, nil, nil)
	s.setActive(false)

	if err != nil {
		// W3C error: invalid session id zaten kapanmış demek
		var w3c *W3CError
		if errors.As(err, &w3c) && w3c.IsSessionLost() {
			return nil
		}
		return fmt.Errorf("quit session: %w", err)
	}
	return nil
}

// path, session-prefixed URL path döndürür
func (s *Session) path(suffix string) string {
	return "/session/" + s.sessionID + suffix
}

// screenshot, mevcut ekran görüntüsünü PNG bytes olarak alır
func (s *Session) screenshot(ctx context.Context) ([]byte, error) {
	var resp struct {
		Value string `json:"value"` // base64 encoded PNG
	}
	if err := s.client.do(ctx, http.MethodGet, s.path("/screenshot"), nil, &resp); err != nil {
		return nil, fmt.Errorf("screenshot: %w", err)
	}
	if resp.Value == "" {
		return nil, fmt.Errorf("empty screenshot value")
	}
	return base64Decode(resp.Value)
}

// checkActive, her UI komutu öncesi session'ın aktif olduğunu doğrular
func (s *Session) checkActive() error {
	if !s.IsActive() {
		return ErrSessionNotActive
	}
	return nil
}
