package backend

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/DeepuExplainsAi/pocketAiHarness/bridge/internal/guard"
)

// Claude runs the official claude CLI in stream-json mode, one process per chat.
// Tool permission requests arrive as control_request messages and are answered
// with control_response, so every decision comes from the user interface.
type Claude struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	events chan Event
	opts   Options
	cancel context.CancelFunc
	ctx    context.Context

	wmu    sync.Mutex
	nextID atomic.Int64
	seq    int
}

// ErrModeNotAllowed is returned for permission modes that would skip the user's approval.
var ErrModeNotAllowed = errors.New("only the manual permission mode is allowed")

func StartClaude(parent context.Context, binary string, o Options) (*Claude, error) {
	if o.Mode != "" && o.Mode != "manual" {
		return nil, ErrModeNotAllowed
	}
	args := []string{
		"-p", "--input-format", "stream-json", "--output-format", "stream-json",
		"--verbose", "--include-partial-messages", "--permission-prompt-tool", "stdio",
		// User, project and local settings can hold allow rules and hooks that skip the prompt.
		"--setting-sources", "",
		"--permission-mode", "manual",
	}
	if o.Model != "" {
		args = append(args, "--model", o.Model)
	}
	if o.Resume != "" {
		args = append(args, "--resume", o.Resume)
	}

	ctx, cancel := context.WithCancel(parent)
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = o.Cwd
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
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	c := &Claude{cmd: cmd, stdin: stdin, events: make(chan Event, 64), opts: o, cancel: cancel, ctx: ctx}
	go c.read(stdout)
	return c, nil
}

func (c *Claude) read(r io.Reader) {
	defer close(c.events)
	defer func() { c.cancel(); _ = c.cmd.Wait() }() // cancel first: pending approvals end with a deny

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := append([]byte(nil), sc.Bytes()...)
		var head struct {
			Type    string `json:"type"`
			Request *struct {
				Subtype  string         `json:"subtype"`
				ToolName string         `json:"tool_name"`
				Desc     string         `json:"description"`
				Input    map[string]any `json:"input"`
			} `json:"request"`
			RequestID string `json:"request_id"`
		}
		if json.Unmarshal(line, &head) != nil {
			continue
		}
		switch {
		case head.Type == "control_request" && head.Request != nil && head.Request.Subtype == "can_use_tool":
			go c.answer(head.RequestID, ToolRequest{Tool: head.Request.ToolName, Description: head.Request.Desc, Input: head.Request.Input})
		case head.Type == "control_response", head.Type == "keep_alive":
			// transport detail; not shown to the user
		default:
			c.seq++
			select {
			case c.events <- Event{Seq: c.seq, Raw: line}:
			case <-c.ctx.Done():
				return
			}
		}
	}
	if err := sc.Err(); err != nil {
		c.seq++
		raw, _ := json.Marshal(map[string]any{"type": "bridge_error", "message": "provider output could not be read: " + err.Error()})
		select {
		case c.events <- Event{Seq: c.seq, Raw: raw}:
		case <-c.ctx.Done():
		}
	}
}

func (c *Claude) answer(requestID string, r ToolRequest) {
	d := c.decide(r)
	body := map[string]any{"behavior": "deny", "message": d.Reason}
	if d.Allow {
		body = map[string]any{"behavior": "allow", "updatedInput": r.Input}
	}
	_ = c.write(map[string]any{"type": "control_response", "response": map[string]any{
		"subtype": "success", "request_id": requestID, "response": body,
	}})
}

func (c *Claude) decide(r ToolRequest) Decision {
	if p := outsidePath(c.opts.Cwd, r.Input); p != "" {
		return Decision{Reason: "path is outside the workspace: " + p}
	}
	if c.opts.Ask == nil {
		return Decision{Reason: "no approval channel"}
	}
	return c.opts.Ask(c.ctx, r)
}

// outsidePath returns the first path in the tool input that escapes the workspace.
// It looks at the known file keys. Shell commands and URLs are not inspected: the user
// sees them on the approval card, and containment of a command needs an OS sandbox.
func outsidePath(root string, input map[string]any) string {
	for _, key := range []string{"file_path", "path", "notebook_path"} {
		v, present := input[key]
		if !present {
			continue
		}
		s, ok := v.(string)
		if !ok || s == "" {
			return fmt.Sprint(v) // unexpected shape: refuse rather than guess
		}
		if _, err := guard.ResolveInside(root, s); err != nil {
			return s
		}
	}
	return ""
}

func (c *Claude) write(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	_, err = c.stdin.Write(append(b, '\n'))
	return err
}

func (c *Claude) control(req map[string]any) error {
	id := "pah-" + strconv.FormatInt(c.nextID.Add(1), 10)
	return c.write(map[string]any{"type": "control_request", "request_id": id, "request": req})
}

func (c *Claude) Send(text string) error {
	return c.write(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": text}})
}
func (c *Claude) Interrupt() error { return c.control(map[string]any{"subtype": "interrupt"}) }
func (c *Claude) SetMode(m string) error {
	if m != "manual" && m != "plan" {
		return ErrModeNotAllowed
	}
	return c.control(map[string]any{"subtype": "set_permission_mode", "mode": m})
}
func (c *Claude) SetModel(m string) error {
	return c.control(map[string]any{"subtype": "set_model", "model": m})
}
func (c *Claude) Events() <-chan Event { return c.events }
func (c *Claude) Close() error {
	_ = c.stdin.Close()
	c.cancel()
	return nil
}

var _ Session = (*Claude)(nil)
