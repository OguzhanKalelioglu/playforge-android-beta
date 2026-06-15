package appium

import (
	"context"
	"fmt"
	"net/http"
)

// AppActivate, uygulamayı ön plana getirir (varsa)
// Appium'da ActivateApp komutu
func (s *Session) AppActivate(ctx context.Context, packageName string) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	req := map[string]interface{}{
		"appId": packageName,
	}
	return s.client.do(ctx, http.MethodPost, s.path("/app/activate"), req, nil)
}

// AppTerminate, uygulamayı zorla kapatır
// iOS'ta çalışmaz, Android'de force-stop
func (s *Session) AppTerminate(ctx context.Context, packageName string) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	req := map[string]interface{}{
		"appId": packageName,
	}
	return s.client.do(ctx, http.MethodPost, s.path("/app/terminate"), req, nil)
}

// AppIsInstalled, uygulama yüklü mü?
func (s *Session) AppIsInstalled(ctx context.Context, packageName string) (bool, error) {
	if err := s.checkActive(); err != nil {
		return false, err
	}
	req := map[string]interface{}{
		"appId": packageName,
	}
	var resp struct {
		Value bool `json:"value"`
	}
	if err := s.client.do(ctx, http.MethodPost, s.path("/app/is_installed"), req, &resp); err != nil {
		return false, err
	}
	return resp.Value, nil
}

// AppState, uygulama durumunu sorgular
// Return: 0=not installed, 1=not running, 2=background, 3=foreground, 4=shutdown
func (s *Session) AppState(ctx context.Context, packageName string) (int, error) {
	if err := s.checkActive(); err != nil {
		return 0, err
	}
	req := map[string]interface{}{
		"appId": packageName,
	}
	var resp struct {
		Value int `json:"value"`
	}
	if err := s.client.do(ctx, http.MethodPost, s.path("/app/state"), req, &resp); err != nil {
		return 0, err
	}
	return resp.Value, nil
}

// GetClipboard, clipboard içeriğini alır (Android 7+)
func (s *Session) GetClipboard(ctx context.Context) (string, error) {
	if err := s.checkActive(); err != nil {
		return "", err
	}
	var resp struct {
		Value string `json:"value"`
	}
	if err := s.client.do(ctx, http.MethodPost, s.path("/clipboard/get"), nil, &resp); err != nil {
		return "", fmt.Errorf("get clipboard: %w", err)
	}
	return resp.Value, nil
}

// SetClipboard, clipboard'a yazar
func (s *Session) SetClipboard(ctx context.Context, text string) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	req := map[string]interface{}{
		"contentType": "plaintext",
		"content":     text,
	}
	return s.client.do(ctx, http.MethodPost, s.path("/clipboard/set"), req, nil)
}

// ActivityStart, belirtilen activity'yi başlatır
func (s *Session) ActivityStart(ctx context.Context, packageName, activityName string) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	req := map[string]interface{}{
		"appPackage":   packageName,
		"appActivity":  activityName,
	}
	return s.client.do(ctx, http.MethodPost, s.path("/app/launch"), req, nil)
}
