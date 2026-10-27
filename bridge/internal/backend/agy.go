package backend

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// ErrUnsupported is returned for controls a backend cannot change mid-session.
var ErrUnsupported = errors.New("not supported by this backend")

// AgyHeadless runs the official agy binary in print mode with stream-json input and output.
// Headless agy cannot ask the user for tool approval: unapproved tools are soft-denied
// by agy itself, so this backend never grants anything.
type AgyHeadless struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	events chan Event
	cancel context.CancelFunc
	wmu    sync.Mutex
	seq    int
}

func StartAgy(parent context.Context, binary string, o Options) (*AgyHeadless, error) {
	if o.Mode != "" && o.Mode != "plan" {
		return nil, ErrModeNotAllowed
	}
	args := []string{"-p", "--input-format", "stream-json", "--output-format", "stream-json"}
	if o.Model != "" {
		args = append(args, "--model", o.Model)
	}
	if o.Mode != "" {
		args = append(args, "--mode", o.Mode)
	}
	if o.Resume != "" {
		args = append(args, "--conversation", o.Resume)
	}

	ctx, cancel := context.WithCancel(parent)
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = o.Cwd
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	cmd.WaitDelay = 3 * time.Second // escalate to kill if SIGTERM is ignored
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	a := &AgyHeadless{cmd: cmd, stdin: stdin, events: make(chan Event, 64), cancel: cancel}
	go a.read(ctx, stdout)
	return a, nil
}

func (a *AgyHeadless) read(ctx context.Context, r io.Reader) {
	defer close(a.events)
	defer func() { a.cancel(); _ = a.cmd.Wait() }()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := append([]byte(nil), sc.Bytes()...)
		if !json.Valid(line) {
			continue
		}
		a.seq++
		select {
		case a.events <- Event{Seq: a.seq, Raw: line}:
		case <-ctx.Done():
			return
		}
	}
}

func (a *AgyHeadless) Send(text string) error {
	b, err := json.Marshal(map[string]any{"event": "user", "message": map[string]any{"content": text}})
	if err != nil {
		return err
	}
	a.wmu.Lock()
	defer a.wmu.Unlock()
	_, err = a.stdin.Write(append(b, '\n'))
	return err
}

func (a *AgyHeadless) Interrupt() error      { return a.cmd.Process.Signal(os.Interrupt) }
func (a *AgyHeadless) SetMode(string) error  { return ErrUnsupported }
func (a *AgyHeadless) SetModel(string) error { return ErrUnsupported }
func (a *AgyHeadless) Events() <-chan Event  { return a.events }
func (a *AgyHeadless) Close() error          { _ = a.stdin.Close(); a.cancel(); return nil }

var _ Session = (*AgyHeadless)(nil)
