package emulator

import (
	"fmt"
	"sync"
	"time"
)

type Status string

const (
	StatusIdle     Status = "idle"
	StatusBusy     Status = "busy"
	StatusBooting  Status = "booting"
	StatusOffline  Status = "offline"
)

type Emulator struct {
	Index   int
	Serial  string
	Status  Status
	TesterID string
	LastUsed time.Time
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
		serial := fmt.Sprintf("emulator-%d", 5554+2*i)
		p.emulators[serial] = &Emulator{
			Index:  i,
			Serial: serial,
			Status: StatusIdle,
		}
	}
	return p
}

func (p *Pool) Acquire() (*Emulator, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, e := range p.emulators {
		if e.Status == StatusIdle {
			e.Status = StatusBusy
			e.LastUsed = time.Now()
			return e, nil
		}
	}
	return nil, fmt.Errorf("no available emulator (all %d busy)", p.count)
}

func (p *Pool) Release(serial string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.emulators[serial]
	if !ok {
		return fmt.Errorf("emulator %s not found", serial)
	}
	e.Status = StatusIdle
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
		out = append(out, e)
	}
	return out
}
