package appium

import (
	"context"
	"fmt"
	"time"
)

// Wait, element koşulu sağlanana kadar poll eder
// Helper: belirtilen koşul true olana kadar kontrol eder
func (s *Session) Wait(ctx context.Context, condition func() (bool, error), timeout time.Duration, interval time.Duration) error {
	if interval == 0 {
		interval = 500 * time.Millisecond
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		ok, err := condition()
		if err == nil && ok {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("wait timeout after %s", timeout)
}

// WaitForElement, belirtilen element bulunana kadar bekler
func (s *Session) WaitForElement(ctx context.Context, by By, value string, timeout time.Duration) (*Element, error) {
	var lastErr error
	var found *Element
	err := s.Wait(ctx, func() (bool, error) {
		elem, err := s.FindElement(ctx, by, value)
		if err == nil {
			found = elem
			return true, nil
		}
		lastErr = err
		return false, nil
	}, timeout, 1*time.Second)
	if err != nil {
		if lastErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrElementNotFound, lastErr)
		}
		return nil, fmt.Errorf("%w (by=%s, value=%s)", ErrElementNotFound, by, value)
	}
	if found != nil {
		return found, nil
	}
	return s.FindElement(ctx, by, value)
}

// WaitForElementGone, element kaybolana kadar bekler
func (s *Session) WaitForElementGone(ctx context.Context, by By, value string, timeout time.Duration) error {
	return s.Wait(ctx, func() (bool, error) {
		_, err := s.FindElement(ctx, by, value)
		if err != nil {
			return true, nil
		}
		return false, nil
	}, timeout, 1*time.Second)
}

// WaitForText, belirtilen text'i içeren element bulunana kadar bekler
func (s *Session) WaitForText(ctx context.Context, text string, timeout time.Duration) (*Element, error) {
	// Önce Appium'un UiAutomator2 selector'ü ile dene
	selector := fmt.Sprintf(`new UiSelector().textContains("%s")`, text)
	elem, err := s.WaitForElement(ctx, ByAndroidUIAutomator, selector, timeout)
	if err != nil {
		return nil, err
	}
	return elem, nil
}
