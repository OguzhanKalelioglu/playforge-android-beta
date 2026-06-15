package health

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Monitor struct {
	ADBHost   string
	ADBPort   int
	BootTimeout time.Duration
	PollInterval time.Duration
}

func NewMonitor(adbHost string, adbPort int) *Monitor {
	return &Monitor{
		ADBHost:      adbHost,
		ADBPort:      adbPort,
		BootTimeout:  5 * time.Minute,
		PollInterval: 10 * time.Second,
	}
}

type Result struct {
	BootCompleted bool
	BootedAt      time.Time
	AndroidVersion string
	SDKLevel      string
	Manufacturer  string
	Model         string
	Uptime        time.Duration
	Error         string
}

func (m *Monitor) Check(ctx context.Context, serial string) (*Result, error) {
	res := &Result{}

	if err := m.connect(ctx, serial); err != nil {
		res.Error = "connect: " + err.Error()
		return res, err
	}

	out, err := m.shell(ctx, serial, "getprop sys.boot_completed")
	if err != nil {
		res.Error = "boot_completed: " + err.Error()
		return res, err
	}
	bootCompleted := strings.TrimSpace(out) == "1"
	res.BootCompleted = bootCompleted

	if !bootCompleted {
		return res, fmt.Errorf("not booted yet (sys.boot_completed=%s)", strings.TrimSpace(out))
	}

	if v, err := m.shell(ctx, serial, "getprop ro.build.version.release"); err == nil {
		res.AndroidVersion = strings.TrimSpace(v)
	}
	if v, err := m.shell(ctx, serial, "getprop ro.build.version.sdk"); err == nil {
		res.SDKLevel = strings.TrimSpace(v)
	}
	if v, err := m.shell(ctx, serial, "getprop ro.product.manufacturer"); err == nil {
		res.Manufacturer = strings.TrimSpace(v)
	}
	if v, err := m.shell(ctx, serial, "getprop ro.product.model"); err == nil {
		res.Model = strings.TrimSpace(v)
	}

	if t, err := m.shell(ctx, serial, "uptime"); err == nil {
		if d, perr := time.ParseDuration(strings.TrimSpace(t) + "s"); perr == nil {
			res.Uptime = d
		}
	}

	return res, nil
}

func (m *Monitor) WaitForBoot(ctx context.Context, serial string) (*Result, error) {
	deadline := time.Now().Add(m.BootTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		res, err := m.Check(ctx, serial)
		if err == nil && res.BootCompleted {
			res.BootedAt = time.Now()
			return res, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.PollInterval):
		}
	}
	return nil, fmt.Errorf("boot timeout after %s for %s", m.BootTimeout, serial)
}

func (m *Monitor) connect(ctx context.Context, serial string) error {
	addr := fmt.Sprintf("%s:%d", m.ADBHost, m.ADBPort)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "connect", serial)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb connect: %w", err)
	}
	return nil
}

func (m *Monitor) shell(ctx context.Context, serial, command string) (string, error) {
	addr := fmt.Sprintf("%s:%d", m.ADBHost, m.ADBPort)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "shell", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("adb shell: %w (out: %s)", err, string(out))
	}
	return string(out), nil
}
