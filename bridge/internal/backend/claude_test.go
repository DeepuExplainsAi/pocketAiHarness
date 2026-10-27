package backend

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeClaude emulates the stream-json control protocol: it asks to use a tool,
// then stores the host's control_response so the test can inspect it.
const fakeClaude = `#!/bin/sh
echo '{"type":"system","subtype":"init","session_id":"s1"}'
echo '{"type":"control_request","request_id":"q1","request":{"subtype":"can_use_tool","tool_name":"Write","input":{"file_path":"'"$FAKE_TARGET"'"},"description":"write a file"}}'
while read -r line; do
  case "$line" in *control_response*) echo "$line" > "$FAKE_OUT"; echo '{"type":"result","subtype":"success"}'; exit 0;; esac
done
`

// run starts the fake CLI. target receives the workspace path and returns the file the tool wants to write.
func run(t *testing.T, target func(ws string) string, ask Asker) (answer string, events []Event) {
	t.Helper()
	dir := t.TempDir()
	ws := filepath.Join(dir, "ws")
	if err := os.Mkdir(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "claude.sh")
	if err := os.WriteFile(script, []byte(fakeClaude), 0o755); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "answer.json")
	t.Setenv("FAKE_TARGET", target(ws))
	t.Setenv("FAKE_OUT", out)

	s, err := StartClaude(context.Background(), script, Options{Cwd: ws, Ask: ask})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case e, ok := <-s.Events():
			if !ok {
				b, _ := os.ReadFile(out)
				return string(b), events
			}
			events = append(events, e)
		case <-deadline:
			t.Fatal("timeout waiting for the session to finish")
		}
	}
}

func inside(ws string) string { return filepath.Join(ws, "a.txt") }

func TestAllowedByUser(t *testing.T) {
	asked := 0
	ans, _ := run(t, inside, func(_ context.Context, r ToolRequest) Decision {
		asked++
		if r.Tool != "Write" {
			t.Errorf("tool = %q", r.Tool)
		}
		return Decision{Allow: true}
	})
	if asked != 1 {
		t.Fatalf("asked %d times", asked)
	}
	for _, want := range []string{`"request_id":"q1"`, `"behavior":"allow"`, `"updatedInput"`} {
		if !strings.Contains(ans, want) {
			t.Errorf("answer missing %s: %s", want, ans)
		}
	}
}

func TestDeniedByUser(t *testing.T) {
	ans, _ := run(t, inside, func(context.Context, ToolRequest) Decision {
		return Decision{Reason: "not now"}
	})
	if !strings.Contains(ans, `"behavior":"deny"`) || !strings.Contains(ans, "not now") {
		t.Fatalf("answer = %s", ans)
	}
}

func TestOutsideWorkspaceIsDeniedWithoutAsking(t *testing.T) {
	asked := false
	ans, _ := run(t, func(string) string { return "/etc/passwd" }, func(context.Context, ToolRequest) Decision {
		asked = true
		return Decision{Allow: true}
	})
	if asked {
		t.Fatal("user was asked about a path outside the workspace")
	}
	if !strings.Contains(ans, `"behavior":"deny"`) {
		t.Fatalf("answer = %s", ans)
	}
}

func TestNoApprovalChannelDenies(t *testing.T) {
	ans, _ := run(t, inside, nil)
	if !strings.Contains(ans, `"behavior":"deny"`) {
		t.Fatalf("answer = %s", ans)
	}
}

func TestControlMessagesAreNotForwardedAsChat(t *testing.T) {
	_, evs := run(t, inside, func(context.Context, ToolRequest) Decision { return Decision{Allow: true} })
	for _, e := range evs {
		if strings.Contains(string(e.Raw), "control_request") {
			t.Fatalf("control_request leaked to the app: %s", e.Raw)
		}
	}
	if len(evs) != 2 || evs[0].Seq != 1 || evs[1].Seq != 2 {
		t.Fatalf("events = %+v", evs)
	}
}
