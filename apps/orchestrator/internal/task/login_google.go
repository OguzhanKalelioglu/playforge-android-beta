package task

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// LoginGoogleTask, warming sonrası yeni emulator'de Google hesabı ekler
// Appium ile Google account setup wizard'ını yürütür
type LoginGoogleTask struct {
	Email    string
	Password string
}

func (t *LoginGoogleTask) Type() string { return "login_google" }

func (t *LoginGoogleTask) Run(ctx context.Context, env Env) error {
	if t.Email == "" || t.Password == "" {
		return fmt.Errorf("email and password required for login task")
	}

	env.Log("login_google started", zap.String("email", maskEmail(t.Email)))

	// 1. Setup wizard'ı geç (varsa)
	if err := env.ReportStep("skip_setup_wizard", func() error {
		// "Skip" butonu varsa tıkla
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)skip")`,
			10*time.Second)
		if err != nil {
			return nil // yoksa skip
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("skip setup: %w", err)
	}

	// 2. Settings > Accounts > Add Google Account
	if err := env.ReportStep("open_add_account", func() error {
		// Settings'i aç
		return env.Session.ActivityStart(ctx, "com.android.settings",
			"com.android.settings.accounts.AddAccountSettings")
	}); err != nil {
		return fmt.Errorf("open add account: %w", err)
	}

	// 3. Google seçeneğini tıkla
	if err := env.ReportStep("select_google", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)google")`,
			15*time.Second)
		if err != nil {
			return err
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("select google: %w", err)
	}

	// 4. Email gir
	if err := env.ReportStep("enter_email", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().resourceId("identifierId")`,
			20*time.Second)
		if err != nil {
			// Fallback: ilk EditText
			elems, err := env.Session.FindElements(ctx,
				"appium:-android uiautomator",
				`new UiSelector().className("android.widget.EditText")`)
			if err != nil || len(elems) == 0 {
				return fmt.Errorf("email input not found: %w", err)
			}
			elem = elems[0]
		}
		if err := elem.Clear(ctx); err != nil {
			return err
		}
		return elem.SendKeys(ctx, t.Email)
	}); err != nil {
		return fmt.Errorf("enter email: %w", err)
	}

	// 5. Next
	if err := env.ReportStep("click_next_after_email", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)next")`,
			10*time.Second)
		if err != nil {
			return err
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("next after email: %w", err)
	}

	env.Anti.Delay(ctx)

	// 6. Password gir
	if err := env.ReportStep("enter_password", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().resourceId("password")`,
			20*time.Second)
		if err != nil {
			elems, err := env.Session.FindElements(ctx,
				"appium:-android uiautomator",
				`new UiSelector().className("android.widget.EditText")`)
			if err != nil || len(elems) == 0 {
				return fmt.Errorf("password input not found: %w", err)
			}
			elem = elems[0]
		}
		if err := elem.Clear(ctx); err != nil {
			return err
		}
		return elem.SendKeys(ctx, t.Password)
	}); err != nil {
		return fmt.Errorf("enter password: %w", err)
	}

	// 7. Next (final)
	if err := env.ReportStep("click_next_after_password", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)next")`,
			10*time.Second)
		if err != nil {
			return err
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("next after password: %w", err)
	}

	env.Anti.Delay(ctx)

	// 8. I agree / Accept
	if err := env.ReportStep("accept_tos", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)(i agree|accept|more|kabul)")`,
			15*time.Second)
		if err != nil {
			return nil // optional
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("accept tos: %w", err)
	}

	env.Log("login_google completed")
	return nil
}

func maskEmail(email string) string {
	for i, c := range email {
		if c == '@' {
			return email[:i/2] + "***" + email[i:]
		}
	}
	return email
}
