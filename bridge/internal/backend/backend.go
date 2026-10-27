// Package backend adapts agent CLIs and daemons to one session interface.
package backend

import (
	"context"
	"encoding/json"
)

// Event is one raw provider message, forwarded to the app unchanged with a sequence number.
type Event struct {
	Seq int             `json:"seq"`
	Raw json.RawMessage `json:"raw"`
}

type ToolRequest struct {
	Tool        string
	Description string
	Input       map[string]any
}

type Decision struct {
	Allow  bool
	Reason string
}

// Asker is implemented by the permission broker; it blocks until the user answers.
type Asker func(ctx context.Context, r ToolRequest) Decision

type Options struct {
	Cwd    string
	Model  string
	Mode   string
	Resume string
	Ask    Asker
}

type Session interface {
	Send(text string) error
	Interrupt() error
	SetMode(mode string) error
	SetModel(model string) error
	Events() <-chan Event
	Close() error
}
