package container

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Manager struct {
	ComposePath string
	ProjectName string
	ServicePrefix string
}

func NewManager(composePath, projectName, servicePrefix string) *Manager {
	return &Manager{
		ComposePath:   composePath,
		ProjectName:   projectName,
		ServicePrefix: servicePrefix,
	}
}

func (m *Manager) Up(ctx context.Context, service string) error {
	args := []string{"compose", "-p", m.ProjectName, "-f", m.ComposePath, "up", "-d", "--remove-orphans"}
	if service != "" {
		args = append(args, service)
	}
	return m.run(ctx, 5*time.Minute, args...)
}

func (m *Manager) UpAll(ctx context.Context) error {
	return m.Up(ctx, "")
}

func (m *Manager) Down(ctx context.Context) error {
	args := []string{"compose", "-p", m.ProjectName, "-f", m.ComposePath, "down"}
	return m.run(ctx, 2*time.Minute, args...)
}

func (m *Manager) Restart(ctx context.Context, service string) error {
	args := []string{"compose", "-p", m.ProjectName, "-f", m.ComposePath, "restart", service}
	return m.run(ctx, 2*time.Minute, args...)
}

func (m *Manager) Stop(ctx context.Context, service string) error {
	args := []string{"compose", "-p", m.ProjectName, "-f", m.ComposePath, "stop", service}
	return m.run(ctx, 2*time.Minute, args...)
}

func (m *Manager) Start(ctx context.Context, service string) error {
	args := []string{"compose", "-p", m.ProjectName, "-f", m.ComposePath, "start", service}
	return m.run(ctx, 2*time.Minute, args...)
}

func (m *Manager) PS(ctx context.Context) ([]ContainerInfo, error) {
	args := []string{"compose", "-p", m.ProjectName, "-f", m.ComposePath, "ps", "-a", "--format", "json"}
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker ps: %w (out: %s)", err, string(out))
	}
	return parseContainerList(string(out))
}

func (m *Manager) ContainerID(ctx context.Context, service string) (string, error) {
	args := []string{"compose", "-p", m.ProjectName, "-f", m.ComposePath, "ps", "-q", service}
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("docker ps -q: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) Logs(ctx context.Context, service string, tail int) (string, error) {
	args := []string{"compose", "-p", m.ProjectName, "-f", m.ComposePath, "logs", "--tail", fmt.Sprintf("%d", tail), service}
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker logs: %w (out: %s)", err, string(out))
	}
	return string(out), nil
}

func (m *Manager) run(ctx context.Context, timeout time.Duration, args ...string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker %s: %w (out: %s)", strings.Join(args, " "), err, string(output))
	}
	return nil
}

func ServiceFor(index int, prefix string) string {
	if prefix == "" {
		prefix = "emulator"
	}
	return fmt.Sprintf("%s-%02d", prefix, index+1)
}
