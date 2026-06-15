package appium

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Screenshot, ekran görüntüsünü alır ve PNG bytes olarak döner
func (s *Session) Screenshot(ctx context.Context) ([]byte, error) {
	if err := s.checkActive(); err != nil {
		return nil, err
	}
	return s.screenshot(ctx)
}

// ScreenshotToFile, ekran görüntüsünü dosyaya kaydeder
// Returns: dosya path'i (relative)
func (s *Session) ScreenshotToFile(ctx context.Context, dir, prefix string) (string, error) {
	data, err := s.Screenshot(ctx)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	filename := fmt.Sprintf("%s_%s_%d.png", prefix, s.sessionID[:min(8, len(s.sessionID))], time.Now().UnixMilli())
	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("write screenshot: %w", err)
	}
	return fullPath, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
