package emulator

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Status string

const (
	StatusUnknown Status = "unknown"
	StatusStopped Status = "stopped"
	StatusBooting Status = "booting"
	StatusReady   Status = "ready"
	StatusBusy    Status = "busy"
	StatusWiping  Status = "wiping"
	StatusError   Status = "error"
)

type Emulator struct {
	Index       int       `json:"index"`
	Serial      string    `json:"serial"`
	ContainerID string    `json:"container_id,omitempty"`
	Status      Status    `json:"status"`
	TesterID    string    `json:"tester_id,omitempty"`
	LastUsed    time.Time `json:"last_used,omitempty"`
	BootedAt    time.Time `json:"booted_at,omitempty"`
	LastCheck   time.Time `json:"last_check,omitempty"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
	BootCount   int       `json:"boot_count"`
}

type Pool struct {
	mu        sync.RWMutex
	emulators map[string]*Emulator
	count     int
}

func NewPool(count int) *Pool {
	p := &Pool{
		emulators: make(map[string]*Emulator),
		count:     count,
	}
	for i := 0; i < count; i++ {
		serial := SerialFor(i)
		p.emulators[serial] = &Emulator{
			Index:  i,
			Serial: serial,
			Status: StatusStopped,
		}
	}
	return p
}

func SerialFor(index int) string {
	return fmt.Sprintf("emulator-%d", 5554+2*index)
}

func (p *Pool) Acquire() (*Emulator, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, e := range p.emulators {
		if e.Status == StatusReady {
			e.Status = StatusBusy
			e.LastUsed = time.Now()
			return e, nil
		}
	}
	return nil, fmt.Errorf("no available emulator (all %d busy or not ready)", p.count)
}

// AcquireForTest, ready durumda bir emulator alır ve busy yapar (atomic CAS)
// testID/assignmentID ile işaretlenir, aynı anda iki farklı test'in
// aynı emulator'ü almasını engeller.
// Not: ready yoksa anında hata döner. Blocking için AcquireForTestBlocking kullan.
func (p *Pool) AcquireForTest(testID, assignmentID string) (*Emulator, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, e := range p.emulators {
		if e.Status == StatusReady {
			e.Status = StatusBusy
			e.TesterID = assignmentID // assignmentID geçici olarak tutulur
			e.LastUsed = time.Now()
			out := *e
			return &out, nil
		}
	}
	return nil, fmt.Errorf("no available emulator (all %d busy or not ready)", p.count)
}

// AcquireForTestBlocking, ready duruma gelene kadar bekler (max 10dk default)
// Tek mutex critical section içinde CAS yapar
func (p *Pool) AcquireForTestBlocking(ctx context.Context, testID, assignmentID string, pollInterval time.Duration) (*Emulator, error) {
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		em, err := p.AcquireForTest(testID, assignmentID)
		if err == nil {
			return em, nil
		}

		// Wait and retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func (p *Pool) Release(serial string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.emulators[serial]
	if !ok {
		return fmt.Errorf("emulator %s not found", serial)
	}
	e.Status = StatusReady
	e.TesterID = ""
	return nil
}

func (p *Pool) Assign(serial, testerID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.emulators[serial]
	if !ok {
		return fmt.Errorf("emulator %s not found", serial)
	}
	e.TesterID = testerID
	e.Status = StatusBusy
	e.LastUsed = time.Now()
	return nil
}

func (p *Pool) SetStatus(serial string, status Status, errMsg string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.emulators[serial]
	if !ok {
		return fmt.Errorf("emulator %s not found", serial)
	}
	prev := e.Status
	e.Status = status
	e.LastCheck = time.Now()
	e.ErrorMsg = errMsg

	if status == StatusReady && prev != StatusReady {
		e.BootedAt = time.Now()
		e.BootCount++
	}
	if status == StatusError && errMsg != "" {
		e.ErrorMsg = errMsg
	}
	return nil
}

func (p *Pool) SetContainerID(serial, containerID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.emulators[serial]
	if !ok {
		return fmt.Errorf("emulator %s not found", serial)
	}
	e.ContainerID = containerID
	return nil
}

func (p *Pool) Get(serial string) (*Emulator, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	e, ok := p.emulators[serial]
	if !ok {
		return nil, fmt.Errorf("emulator %s not found", serial)
	}
	out := *e
	return &out, nil
}

func (p *Pool) Status() map[string]Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	s := make(map[string]Status, len(p.emulators))
	for k, v := range p.emulators {
		s[k] = v.Status
	}
	return s
}

func (p *Pool) List() []*Emulator {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]*Emulator, 0, len(p.emulators))
	for _, e := range p.emulators {
		cp := *e
		out = append(out, &cp)
	}
	return out
}

func (p *Pool) Count() int {
	return p.count
}

func (p *Pool) Counts() map[Status]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	counts := map[Status]int{}
	for _, e := range p.emulators {
		counts[e.Status]++
	}
	return counts
}

func (p *Pool) Ready() []*Emulator {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]*Emulator, 0)
	for _, e := range p.emulators {
		if e.Status == StatusReady {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out
}
