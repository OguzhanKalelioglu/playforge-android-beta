package scheduler

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/model"
)

// Registrar, Asynq Scheduler'ı 14 günlük plana göre kurar
// Her test için:
//   - Gün 0:  LoginGoogle + TestStart (09:00 ± 2h jitter)
//   - Gün 1-13: DailyEngagement (5 hesap/gün rotasyon, 3 session)
//   - Gün 14: WriteReview
//   - Her gün 23:00: Healthcheck
type Registrar struct {
	scheduler *asynq.Scheduler
	location  *time.Location
	logger    *zap.Logger
}

func NewRegistrar(s *asynq.Scheduler, logger *zap.Logger) *Registrar {
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		loc = time.UTC
	}
	return &Registrar{
		scheduler: s,
		location:  loc,
		logger:    logger,
	}
}

// Register14DayPlan, yeni test için 14 günlük plan kaydeder
// startTime: test başlangıç anı (UTC), jitter bu saat etrafında
// Not: Asynq Scheduler ProcessIn ile göreceli zaman kullanır
func (r *Registrar) Register14DayPlan(testID, packageName string, startTime time.Time) error {
	r.logger.Info("registering 14-day plan",
		zap.String("test_id", testID),
		zap.Time("start", startTime))

	// Test oluşturulduğu an itibariyle 1 saniye sonra test_start job
	// Scheduler ProcessIn ile relatif çalışır
	// Gün 0: hemen
	if err := r.registerTestStart(testID, packageName, startTime); err != nil {
		return err
	}

	// Gün 1-13: günlük engagement
	for day := 1; day <= 13; day++ {
		jobTime := startTime.Add(time.Duration(day) * 24 * time.Hour)
		if err := r.registerDailyEngagement(testID, packageName, day, jobTime); err != nil {
			r.logger.Warn("engagement day registration failed",
				zap.Int("day", day), zap.Error(err))
		}
	}

	// Gün 14: review
	reviewTime := startTime.Add(14 * 24 * time.Hour)
	if err := r.registerReview(testID, packageName, reviewTime); err != nil {
		return fmt.Errorf("review registration: %w", err)
	}

	r.logger.Info("14-day plan registered", zap.String("test_id", testID))
	return nil
}

func (r *Registrar) registerTestStart(testID, packageName string, at time.Time) error {
	// ProcessIn: scheduled time'dan ne kadar sonra çalışsın
	in := time.Until(at)
	if in < 0 {
		in = time.Second // geçmişse hemen çalıştır
	}

	payload, _ := model.TestStartPayload{
		Payload: model.Payload{
			TestID:      testID,
			PackageName: packageName,
		},
	}.Encode()

	jobID := model.JobID(testID, model.TaskTypeTestStart, 0)
	task := asynq.NewTask(string(model.TaskTypeTestStart), payload)

	_, err := r.scheduler.Register(jobID, task, asynq.ProcessIn(in))
	if err != nil {
		// Job zaten varsa skip
		r.logger.Debug("test_start already registered",
			zap.String("test_id", testID))
	}
	return err
}

func (r *Registrar) registerDailyEngagement(testID, packageName string, day int, at time.Time) error {
	in := time.Until(at)
	if in < 0 {
		in = time.Hour
	}

	payload, _ := model.DailyEngagementPayload{
		Payload: model.Payload{
			TestID:      testID,
			PackageName: packageName,
		},
		Day: day,
	}.Encode()

	jobID := model.JobID(testID, model.TaskTypeDailyEngagement, day)
	task := asynq.NewTask(string(model.TaskTypeDailyEngagement), payload)

	_, err := r.scheduler.Register(jobID, task, asynq.ProcessIn(in))
	if err != nil {
		r.logger.Debug("engagement already registered",
			zap.String("test_id", testID), zap.Int("day", day))
	}
	return err
}

func (r *Registrar) registerReview(testID, packageName string, at time.Time) error {
	in := time.Until(at)
	if in < 0 {
		in = time.Hour
	}

	payload, _ := model.WriteReviewPayload{
		Payload: model.Payload{
			TestID:      testID,
			PackageName: packageName,
		},
		Stars:    5,
		Comment:  "Güzel uygulama, tavsiye ederim.",
		Language: "tr",
	}.Encode()

	jobID := model.JobID(testID, model.TaskTypeWriteReview, 14)
	task := asynq.NewTask(string(model.TaskTypeWriteReview), payload)

	_, err := r.scheduler.Register(jobID, task, asynq.ProcessIn(in))
	return err
}
