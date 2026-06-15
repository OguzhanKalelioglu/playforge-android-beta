package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/model"
)

// RunnerClient, orchestrator HTTP API'sine bağlanır
// Worker handler'ları bu client üzerinden task çalıştırır
type RunnerClient struct {
	baseURL string
	token   string
	logger  *zap.Logger
	http    *http.Client
}

func NewRunnerClient(baseURL, token string, logger *zap.Logger) *RunnerClient {
	return &RunnerClient{
		baseURL: baseURL,
		token:   token,
		logger:  logger,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *RunnerClient) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("X-API-Token", c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("runner returned %d", resp.StatusCode)
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// StartTest, orchestrator'a test başlatma isteği gönderir
func (c *RunnerClient) StartTest(ctx context.Context, p model.TestStartPayload) error {
	var resp map[string]interface{}
	return c.do(ctx, http.MethodPost, "/v1/tasks/test_start/start", p, &resp)
}

// StartLoginGoogle, Google hesabı ekleme isteği
func (c *RunnerClient) StartLoginGoogle(ctx context.Context, p model.LoginGooglePayload) error {
	var resp map[string]interface{}
	return c.do(ctx, http.MethodPost, "/v1/tasks/login_google/start", p, &resp)
}

// StartEngagement, günlük engagement isteği
func (c *RunnerClient) StartEngagement(ctx context.Context, p model.DailyEngagementPayload) error {
	var resp map[string]interface{}
	return c.do(ctx, http.MethodPost, "/v1/tasks/engage/start", p, &resp)
}

// StartReview, review yazma isteği
func (c *RunnerClient) StartReview(ctx context.Context, p model.WriteReviewPayload) error {
	var resp map[string]interface{}
	return c.do(ctx, http.MethodPost, "/v1/tasks/review/start", p, &resp)
}

// Health, runner'a ulaşılabilir mi?
func (c *RunnerClient) Health(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/health", nil, nil)
}
