package adb

import (
	"context"
	"fmt"
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

func (c *Client) Shell(ctx context.Context, serial, command string) (string, error) {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	cmd := exec.CommandContext(ctx, "adb", "-s", addr, "-s", serial, "shell", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("adb shell: %w (out: %s)", err, string(out))
	}
	return string(out), nil
}

func (c *Client) StartADBServer() error {
	cmd := exec.Command("adb", "start-server")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb start-server: %w (out: %s)", err, string(out))
	}
	return nil
}

func EmulatorSerial(index int) string {
	port := 5554 + 2*index
	return "emulator-" + strconv.Itoa(port)
}

func WaitForDevice(ctx context.Context, serial string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		cmd := exec.CommandContext(ctx2, "adb", "shell", "getprop", "sys.boot_completed")
		out, err := cmd.CombinedOutput()
		cancel()
		if err == nil {
			s := string(out)
			s = trimWhitespace(s)
			if s == "1" {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("device %s not ready after %s", serial, timeout)
}
