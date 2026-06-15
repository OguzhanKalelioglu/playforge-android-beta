package appium

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Direction, scroll yönü
type Direction string

const (
	DirUp    Direction = "up"
	DirDown  Direction = "down"
	DirLeft  Direction = "left"
	DirRight Direction = "right"
)

// Tap, (x,y) koordinatına dokunur
func (s *Session) Tap(ctx context.Context, x, y int) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	return s.w3cAction(ctx, "actions", []map[string]interface{}{
		{
			"type": "pointer",
			"id":   "finger1",
			"parameters": map[string]string{"pointerType": "touch"},
			"actions": []map[string]interface{}{
				{"type": "pointerMove", "duration": 0, "x": x, "y": y},
				{"type": "pointerDown", "button": 0},
				{"type": "pause", "duration": 80},
				{"type": "pointerUp", "button": 0},
			},
		},
	})
}

// Swipe, (x1,y1) → (x2,y2) arasında sürükleme yapar
// durMs: toplam sürükleme süresi (ms)
func (s *Session) Swipe(ctx context.Context, x1, y1, x2, y2, durMs int) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	if durMs <= 0 {
		durMs = 300
	}
	return s.w3cAction(ctx, "actions", []map[string]interface{}{
		{
			"type": "pointer",
			"id":   "finger1",
			"parameters": map[string]string{"pointerType": "touch"},
			"actions": []map[string]interface{}{
				{"type": "pointerMove", "duration": 0, "x": x1, "y": y1},
				{"type": "pointerDown", "button": 0},
				{"type": "pointerMove", "duration": durMs, "x": x2, "y": y2},
				{"type": "pointerUp", "button": 0},
			},
		},
	})
}

// Scroll, belirtilen yönde ekranı kaydırır
// Default: ekranın ortasından (540, 1200) yukarı/aşağı
func (s *Session) Scroll(ctx context.Context, dir Direction) error {
	x1, y1, x2, y2 := 540, 1200, 540, 400 // default: orta-dikey, yukarı
	switch dir {
	case DirUp:
		x1, y1, x2, y2 = 540, 1500, 540, 400
	case DirDown:
		x1, y1, x2, y2 = 540, 400, 540, 1500
	case DirLeft:
		x1, y1, x2, y2 = 900, 1200, 200, 1200
	case DirRight:
		x1, y1, x2, y2 = 200, 1200, 900, 1200
	}
	return s.Swipe(ctx, x1, y1, x2, y2, 400)
}

// LongPress, (x,y)'ye durMs ms basılı tutar
func (s *Session) LongPress(ctx context.Context, x, y, durMs int) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	if durMs <= 0 {
		durMs = 1000
	}
	return s.w3cAction(ctx, "actions", []map[string]interface{}{
		{
			"type": "pointer",
			"id":   "finger1",
			"parameters": map[string]string{"pointerType": "touch"},
			"actions": []map[string]interface{}{
				{"type": "pointerMove", "duration": 0, "x": x, "y": y},
				{"type": "pointerDown", "button": 0},
				{"type": "pause", "duration": durMs},
				{"type": "pointerUp", "button": 0},
			},
		},
	})
}

// Back, sistem geri tuşuna basar
func (s *Session) Back(ctx context.Context) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	return s.client.do(ctx, http.MethodPost, s.path("/back"), nil, nil)
}

// Home, ana ekrana döner
func (s *Session) Home(ctx context.Context) error {
	if err := s.checkActive(); err != nil {
		return err
	}
	return s.w3cAction(ctx, "actions", []map[string]interface{}{
		{
			"type": "key",
			"id":   "home",
			"actions": []map[string]interface{}{
				{"type": "keyDown", "value": 3},   // KEYCODE_HOME = 3
				{"type": "keyUp", "value": 3},
			},
		},
	})
}

// TapElement, elementin merkezine tap yapar
func (e *Element) Tap(ctx context.Context) error {
	if err := e.session.checkActive(); err != nil {
		return err
	}
	if err := e.clickRaw(ctx); err != nil {
		return err
	}
	return nil
}

func (e *Element) clickRaw(ctx context.Context) error {
	return e.session.client.do(ctx, http.MethodPost, e.path("/click"), nil, nil)
}

// SendKeys, elemente text gönderir
func (e *Element) SendKeys(ctx context.Context, text string) error {
	if err := e.session.checkActive(); err != nil {
		return err
	}
	return e.session.client.do(ctx, http.MethodPost, e.path("/value"), map[string]string{
		"text": text,
	}, nil)
}

// Clear, input element'in içeriğini temizler
func (e *Element) Clear(ctx context.Context) error {
	if err := e.session.checkActive(); err != nil {
		return err
	}
	return e.session.client.do(ctx, http.MethodPost, e.path("/clear"), nil, nil)
}

// w3cAction, W3C WebDriver "actions" endpoint'ine istek gönderir
func (s *Session) w3cAction(ctx context.Context, action string, actions []map[string]interface{}) error {
	req := map[string]interface{}{
		"actions": actions,
	}
	return s.client.do(ctx, http.MethodPost, s.path("/"+action), req, nil)
}

// Wait, belirtilen süre boyunca uyur (gesture arası)
func Wait(durMs int) {
	time.Sleep(time.Duration(durMs) * time.Millisecond)
}

// String, debug amaçlı string representation
func (s *Session) String() string {
	active := s.IsActive()
	return fmt.Sprintf("AppiumSession{id=%s, active=%v}", s.sessionID, active)
}
