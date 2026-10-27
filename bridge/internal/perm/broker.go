// Package perm turns agent tool requests into user decisions. Nothing is allowed
// unless the user interface answers; timeouts, disconnects and unknown ids deny.
package perm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Request struct {
	ID          string `json:"request"`
	Chat        string `json:"chat"`
	Tool        string `json:"tool"`
	Description string `json:"description,omitempty"`
	Input       any    `json:"input,omitempty"`
}

type Decision struct {
	Allow  bool
	Reason string
}

type Broker struct {
	mu      sync.Mutex
	pending map[string]chan Decision
}

func New() *Broker { return &Broker{pending: map[string]chan Decision{}} }

// Ask registers the request, calls notify so the UI can show it, and waits.
func (b *Broker) Ask(ctx context.Context, r Request, notify func(Request)) Decision {
	r.ID = newID()
	ch := make(chan Decision, 1)
	b.mu.Lock()
	b.pending[r.ID] = ch
	b.mu.Unlock()
	defer b.forget(r.ID)

	notify(r)
	select {
	case d := <-ch:
		return d
	case <-ctx.Done():
		return Decision{Reason: "request cancelled"}
	}
}

// Decide delivers the user's answer. It reports false for an unknown or already answered id.
func (b *Broker) Decide(id string, d Decision) bool {
	b.mu.Lock()
	ch, ok := b.pending[id]
	delete(b.pending, id)
	b.mu.Unlock()
	if ok {
		ch <- d
	}
	return ok
}

func (b *Broker) forget(id string) {
	b.mu.Lock()
	delete(b.pending, id)
	b.mu.Unlock()
}

func newID() string {
	var raw [12]byte
	_, _ = rand.Read(raw[:])
	return hex.EncodeToString(raw[:])
}
