package taskrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ActivitySink, task'lardan gelen step event'lerini toplayıp
// API'ye batch olarak yollar. Network hatasında disk fallback kullanır.
type ActivitySink struct {
	apiURL      string
	apiToken    string
	fallbackDir string
	logger      *zap.Logger

	ch         chan activityEvent
	httpClient *http.Client

	mu            sync.Mutex
	closed        bool
	failedBuffer  []activityEvent
	droppedCount  int64
	totalEmitted  int64
	totalUploaded int64
}

type activityEvent struct {
	EventType      string                 `json:"event_type"`
	TestID         string                 `json:"test_id"`
	AssignmentID   string                 `json:"assignment_id"`
	StepName       string                 `json:"step_name,omitempty"`
	Status         string                 `json:"status"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	ScreenshotPath string                 `json:"screenshot_path,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	OccurredAt     time.Time              `json:"occurred_at"`
}

func NewActivitySink(apiURL, apiToken, fallbackDir string, logger *zap.Logger) *ActivitySink {
	s := &ActivitySink{
		apiURL:      apiURL,
		apiToken:    apiToken,
		fallbackDir: fallbackDir,
		logger:      logger,
		ch:          make(chan activityEvent, 256),
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
	_ = os.MkdirAll(fallbackDir, 0755)
	return s
}

// Emit, non-blocking olarak event ekler
// Backpressure durumunda disk fallback'e yazar
func (s *ActivitySink) Emit(event activityEvent) {
	s.mu.Lock()
	s.totalEmitted++
	s.mu.Unlock()

	select {
	case s.ch <- event:
		// queued
	default:
		// Buffer dolu → disk fallback
		s.mu.Lock()
		s.droppedCount++
		s.failedBuffer = append(s.failedBuffer, event)
		s.mu.Unlock()
	}
}

// Start, batch upload worker'ını başlatır
// Her 2 saniye veya 50 event birikince flush eder
func (s *ActivitySink) Start(ctx context.Context) {
	go s.batchLoop(ctx)
	go s.replayFallback(ctx) // Disk fallback replay (5dk'da bir)
}

func (s *ActivitySink) batchLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	batch := make([]activityEvent, 0, 50)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := s.upload(ctx, batch); err != nil {
			s.logger.Warn("activity upload failed, writing to disk",
				zap.Error(err),
				zap.Int("batch_size", len(batch)))
			s.mu.Lock()
			s.failedBuffer = append(s.failedBuffer, batch...)
			s.mu.Unlock()
			s.writeToDisk(batch)
		} else {
			s.mu.Lock()
			s.totalUploaded += int64(len(batch))
			s.mu.Unlock()
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case ev := <-s.ch:
			batch = append(batch, ev)
			if len(batch) >= 50 {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (s *ActivitySink) upload(ctx context.Context, batch []activityEvent) error {
	if s.apiURL == "" {
		// API yok, direkt disk'e yaz (local mode)
		return fmt.Errorf("api_url not configured")
	}

	body := map[string]interface{}{
		"events": batch,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.apiToken != "" {
		req.Header.Set("X-Activity-Token", s.apiToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upload status %d", resp.StatusCode)
	}
	return nil
}

func (s *ActivitySink) writeToDisk(batch []activityEvent) {
	filename := fmt.Sprintf("failed_%d.jsonl", time.Now().Unix())
	fullPath := filepath.Join(s.fallbackDir, filename)
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Error("disk fallback write failed", zap.Error(err))
		return
	}
	defer f.Close()

	for _, ev := range batch {
		if data, err := json.Marshal(ev); err == nil {
			f.Write(data)
			f.Write([]byte("\n"))
		}
	}
}

func (s *ActivitySink) replayFallback(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.replayOnce(ctx)
		}
	}
}

func (s *ActivitySink) replayOnce(ctx context.Context) {
	s.mu.Lock()
	// In-memory buffer'ı önce dene
	if len(s.failedBuffer) > 0 {
		batch := s.failedBuffer
		s.failedBuffer = nil
		s.mu.Unlock()

		if err := s.upload(ctx, batch); err != nil {
			s.logger.Warn("in-memory replay failed", zap.Error(err))
			s.mu.Lock()
			s.failedBuffer = append(batch, s.failedBuffer...)
			s.mu.Unlock()
		}
		return
	}
	s.mu.Unlock()

	// Disk'ten oku
	files, err := filepath.Glob(filepath.Join(s.fallbackDir, "failed_*.jsonl"))
	if err != nil || len(files) == 0 {
		return
	}

	// En eski dosyayı dene
	oldest := files[0]
	events := s.readFromDisk(oldest)
	if len(events) == 0 {
		return
	}

	if err := s.upload(ctx, events); err != nil {
		s.logger.Warn("disk replay failed", zap.String("file", oldest), zap.Error(err))
		return
	}
	os.Remove(oldest)
	s.logger.Info("replayed activity events from disk", zap.Int("count", len(events)))
}

func (s *ActivitySink) readFromDisk(path string) []activityEvent {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	events := make([]activityEvent, 0)
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var ev activityEvent
		if err := json.Unmarshal(line, &ev); err == nil {
			events = append(events, ev)
		}
	}
	return events
}

// Stats, monitoring için sayaçlar
func (s *ActivitySink) Stats() (emitted, uploaded, dropped int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalEmitted, s.totalUploaded, s.droppedCount
}

// Close, channel'ı kapatır (graceful shutdown)
func (s *ActivitySink) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	close(s.ch)
	s.mu.Unlock()
}
