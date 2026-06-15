package adb

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Client struct {
	Host string
	Port int
}

func NewClient(host string, port int) *Client {
	return &Client{Host: host, Port: port}
}

func (c *Client) StartADBServer() error {
	cmd := exec.Command("adb", "start-server")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb start-server: %w (out: %s)", err, string(out))
	}
	return nil
}

func (c *Client) Connect(serial string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.Command("adb", "-s", addr, "connect", serial)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb connect %s: %w (out: %s)", serial, err, string(out))
	}
	return nil
}

func (c *Client) Devices() ([]string, error) {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.Command("adb", "-s", addr, "devices")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("adb devices: %w (out: %s)", err, string(out))
	}
	return parseDevices(string(out)), nil
}

func (c *Client) Shell(ctx context.Context, serial, command string) (string, error) {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "shell", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("adb shell: %w (out: %s)", err, string(out))
	}
	return string(out), nil
}

func (c *Client) ShellTimeout(ctx context.Context, serial, command string, timeout time.Duration) (string, error) {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "shell", command)
	timer, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd = exec.CommandContext(timer, "adb", "-s", addr, "-s", serial, "shell", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("adb shell: %w (out: %s)", err, string(out))
	}
	return string(out), nil
}

func (c *Client) Push(ctx context.Context, serial, local, remote string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "push", local, remote)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb push: %w (out: %s)", err, string(out))
	}
	return nil
}

func (c *Client) Pull(ctx context.Context, serial, remote, local string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "pull", remote, local)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb pull: %w (out: %s)", err, string(out))
	}
	return nil
}

func (c *Client) Screenshot(ctx context.Context, serial, localPath string) error {
	remotePath := "/sdcard/screenshot.png"
	if _, err := c.Shell(ctx, serial, "screencap -p "+remotePath); err != nil {
		return err
	}
	if err := c.Pull(ctx, serial, remotePath, localPath); err != nil {
		return err
	}
	_, _ = c.Shell(ctx, serial, "rm "+remotePath)
	return nil
}

func (c *Client) Install(ctx context.Context, serial, apkPath string) (string, error) {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "install", "-r", "-t", apkPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("adb install: %w (out: %s)", err, string(out))
	}
	return string(out), nil
}

func (c *Client) Uninstall(ctx context.Context, serial, packageName string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "uninstall", packageName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb uninstall: %w (out: %s)", err, string(out))
	}
	return nil
}

func (c *Client) WipeData(ctx context.Context, serial string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "shell", "pm", "clear", "--user", "0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wipe data: %w (out: %s)", err, string(out))
	}
	return nil
}

func (c *Client) Reboot(ctx context.Context, serial string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "reboot")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb reboot: %w (out: %s)", err, string(out))
	}
	return nil
}

func (c *Client) AddGoogleAccount(ctx context.Context, serial, email, password string) error {
	if _, err := c.Shell(ctx, serial, "settings put secure user_setup_complete 1"); err != nil {
		return err
	}

	if _, err := c.Shell(ctx, serial, "am force-stop com.google.android.gms"); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	if _, err := c.Shell(ctx, serial, "am start -a android.settings.ADD_ACCOUNT_SETTINGS"); err != nil {
		return err
	}

	_ = email
	_ = password
	return fmt.Errorf("google account add requires UI automation (appium) - use task runner")
}

func (c *Client) IsBooted(ctx context.Context, serial string) (bool, error) {
	out, err := c.Shell(ctx, serial, "getprop sys.boot_completed")
	if err != nil {
		return false, err
	}
	return trimWhitespace(out) == "1", nil
}

func (c *Client) Logcat(ctx context.Context, serial string, lines int) (string, error) {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "logcat", "-d", "-t", strconv.Itoa(lines))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("logcat: %w (out: %s)", err, string(out))
	}
	return string(out), nil
}

func (c *Client) StreamLogcat(ctx context.Context, serial string, w io.Writer) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "logcat", "-v", "time")
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func SerialFor(index int) string {
	port := 5554 + 2*index
	return "emulator-" + strconv.Itoa(port)
}

func parseDevices(output string) []string {
	var devices []string
	for _, line := range splitLines(output) {
		if line == "" || startsWith(line, "List of devices") || startsWith(line, "*") {
			continue
		}
		if endsWith(line, "device") {
			parts := splitByWhitespace(line)
			if len(parts) > 0 {
				devices = append(devices, parts[0])
			}
		}
	}
	return devices
}
