package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/testerscommunity/api/internal/model"
)

// AsynqClient, Asynq client'ı sarmalayan helper
// Enqueue method'ları type-safe payload'larla çalışır
type AsynqClient struct {
	client *asynq.Client
}

func NewAsynqClient(c *asynq.Client) *AsynqClient {
	return &AsynqClient{client: c}
}

// EnqueueTestStart, yeni test başlatma task'ı ekler
func (a *AsynqClient) EnqueueTestStart(ctx context.Context, p model.TestStartPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload, err := p.Encode()
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	task := asynq.NewTask(string(model.TaskTypeTestStart), payload)
	opts = append(opts, asynq.TaskID(model.JobID(p.TestID, model.TaskTypeTestStart, 0)))
	return a.client.Enqueue(task, opts...)
}

// EnqueueLoginGoogle, Google hesabı ekleme task'ı (TestStart'tan önce)
// Asynq scheduler'dan değil, doğrudan orchestrator tarafından tetiklenir
func (a *AsynqClient) EnqueueLoginGoogle(ctx context.Context, p model.LoginGooglePayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload, err := p.Encode()
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	task := asynq.NewTask(string(model.TaskTypeLoginGoogle), payload)
	opts = append(opts, asynq.TaskID(model.JobID(p.TestID, model.TaskTypeLoginGoogle, 0)))
	return a.client.Enqueue(task, opts...)
}

// EnqueueDailyEngagement, günlük engagement task'ı
// Scheduler tarafından çağrılır (14 günlük plan)
func (a *AsynqClient) EnqueueDailyEngagement(ctx context.Context, p model.DailyEngagementPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload, err := p.Encode()
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	task := asynq.NewTask(string(model.TaskTypeDailyEngagement), payload)
	opts = append(opts, asynq.TaskID(model.JobID(p.TestID, model.TaskTypeDailyEngagement, p.Day)))
	return a.client.Enqueue(task, opts...)
}

// EnqueueWriteReview, review yazma task'ı (gün 14)
func (a *AsynqClient) EnqueueWriteReview(ctx context.Context, p model.WriteReviewPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload, err := p.Encode()
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	task := asynq.NewTask(string(model.TaskTypeWriteReview), payload)
	opts = append(opts, asynq.TaskID(model.JobID(p.TestID, model.TaskTypeWriteReview, 14)))
	return a.client.Enqueue(task, opts...)
}

// EnqueueHealthcheck, günlük 23:00 sağlık kontrolü
func (a *AsynqClient) EnqueueHealthcheck(ctx context.Context, p model.HealthcheckPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload, err := p.Encode()
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	task := asynq.NewTask(string(model.TaskTypeHealthcheck), payload)
	opts = append(opts, asynq.TaskID(model.JobID(p.TestID, model.TaskTypeHealthcheck, -1)))
	return a.client.Enqueue(task, opts...)
}
