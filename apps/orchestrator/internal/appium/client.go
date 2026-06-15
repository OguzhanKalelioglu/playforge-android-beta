package appium

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client, W3C WebDriver protocol HTTP client
// Appium server'a bağlanır, session yönetir
type Client struct {
	baseURL string
	http    *http.Client
}

type Option func(*Client)

func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.http.Timeout = d }
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) URL() string { return c.baseURL }

// do, HTTP isteği yapar, response'u parse eder
func (c *Client) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("parse base url: %w", err)
	}
	u.Path = path

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrServerUnreachable, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	// Appium 4xx/5xx hata response'ları W3C format'ta
	if resp.StatusCode >= 400 {
		var w3cErr W3CError
		if err := json.Unmarshal(respBody, &w3cErr); err == nil && w3cErr.Code != "" {
			w3cErr.HTTPStatus = resp.StatusCode
			return &w3cErr
		}
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshal response: %w (body: %s)", err, string(respBody))
		}
	}
	return nil
}

// Status, server'ın ayakta olup olmadığını kontrol eder
func (c *Client) Status(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/status", nil, nil)
}

// CreateSession, yeni bir WebDriver session oluşturur
func (c *Client) CreateSession(ctx context.Context, caps Capabilities) (*Session, error) {
	w3cCaps := capsToW3C(caps)
	req := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"alwaysMatch": w3cCaps,
		},
	}

	var resp struct {
		Value struct {
			SessionID string                 `json:"sessionId"`
			Capabilities map[string]interface{} `json:"capabilities"`
		} `json:"value"`
	}

	if err := c.do(ctx, http.MethodPost, "/session", req, &resp); err != nil {
		return nil, err
	}
	if resp.Value.SessionID == "" {
		return nil, fmt.Errorf("empty session id in response")
	}

	return &Session{
		client:    c,
		sessionID: resp.Value.SessionID,
		active:    true,
	}, nil
}

// GetSession, mevcut bir session'a ID ile bağlanır (ör: orchestrator restart sonrası)
func (c *Client) GetSession(ctx context.Context, sessionID string) *Session {
	return &Session{
		client:    c,
		sessionID: sessionID,
		active:    true,
	}
}

func capsToW3C(c Capabilities) map[string]interface{} {
	out := map[string]interface{}{
		"platformName": c.PlatformName,
	}

	// W3C standard
	if c.BrowserName != "" {
		out["browserName"] = c.BrowserName
	}
	if c.AcceptInsecureCerts {
		out["acceptInsecureCerts"] = true
	}

	// appium: prefix'li vendor options
	appium := map[string]interface{}{}
	if c.AutomationName != "" {
		appium["automationName"] = c.AutomationName
	}
	if c.DeviceName != "" {
		appium["deviceName"] = c.DeviceName
	}
	if c.App != "" {
		appium["app"] = c.App
	}
	if c.AppPackage != "" {
		appium["appPackage"] = c.AppPackage
	}
	if c.AppActivity != "" {
		appium["appActivity"] = c.AppActivity
	}
	if c.AppWaitActivity != "" {
		appium["appWaitActivity"] = c.AppWaitActivity
	}
	if c.NoReset {
		appium["noReset"] = true
	}
	if c.FullReset {
		appium["fullReset"] = true
	}
	if c.AutoGrantPermissions {
		appium["autoGrantPermissions"] = true
	}
	if c.NewCommandTimeout > 0 {
		appium["newCommandTimeout"] = c.NewCommandTimeout
	}
	if c.Uiautomator2ServerLaunchTimeout > 0 {
		appium["uiautomator2ServerLaunchTimeout"] = c.Uiautomator2ServerLaunchTimeout
	}
	if c.Uiautomator2ServerInstallTimeout > 0 {
		appium["uiautomator2ServerInstallTimeout"] = c.Uiautomator2ServerInstallTimeout
	}
	if c.SkipServerInstallation {
		appium["skipServerInstallation"] = true
	}
	if c.UDID != "" {
		appium["udid"] = c.UDID
	}
	if c.SystemPort > 0 {
		appium["systemPort"] = c.SystemPort
	}
	if c.ChromeDriverPort > 0 {
		appium["chromedriverPort"] = c.ChromeDriverPort
	}
	if c.Orientation != "" {
		appium["orientation"] = c.Orientation
	}
	if c.Locale != "" {
		appium["locale"] = c.Locale
	}
	if c.TimeZone != "" {
		appium["timezone"] = c.TimeZone
	}

	if len(appium) > 0 {
		out["appium:options"] = appium
	}
	return out
}
