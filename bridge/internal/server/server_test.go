package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/DeepuExplainsAi/pocketAiHarness/bridge/internal/backend"
)

func TestListenRefusesNonLoopback(t *testing.T) {
	for _, addr := range []string{"0.0.0.0:0", "192.168.1.5:0", ":0", "example.com:80"} {
		if l, err := Listen(addr); err == nil {
			l.Close()
			t.Errorf("%s was accepted", addr)
		}
	}
	l, err := Listen("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	l.Close()
}

func dial(t *testing.T, url string, h http.Header) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)
	return websocket.Dial(ctx, "ws"+strings.TrimPrefix(url, "http")+"/ws", &websocket.DialOptions{HTTPHeader: h})
}

func TestRejectsMissingWrongTokenAndBrowserOrigin(t *testing.T) {
	s := New(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	cases := map[string]http.Header{
		"no token":    {},
		"wrong token": {"Authorization": {"Bearer nope"}},
		"origin set":  {"Authorization": {"Bearer " + s.Token()}, "Origin": {"https://evil.example"}},
	}
	for name, h := range cases {
		c, resp, err := dial(t, ts.URL, h)
		if err == nil {
			c.CloseNow()
			t.Errorf("%s: connection accepted", name)
			continue
		}
		if resp == nil || resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s: want 403, got %v", name, resp)
		}
	}
}

const fakeCLI = `#!/bin/sh
echo '{"type":"system","subtype":"init"}'
echo '{"type":"control_request","request_id":"q1","request":{"subtype":"can_use_tool","tool_name":"Bash","input":{"command":"ls"},"description":"list files"}}'
while read -r line; do
  case "$line" in *control_response*) echo "$line" > "$FAKE_OUT"; echo '{"type":"result"}'; exit 0;; esac
done
`

func read(t *testing.T, c *websocket.Conn) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, b, err := c.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func write(t *testing.T, c *websocket.Conn, v any) {
	t.Helper()
	b, _ := json.Marshal(v)
	if err := c.Write(context.Background(), websocket.MessageText, b); err != nil {
		t.Fatal(err)
	}
}

func TestPermissionRoundTripThroughTheApp(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "claude.sh")
	out := filepath.Join(dir, "answer.json")
	os.WriteFile(script, []byte(fakeCLI), 0o755)
	t.Setenv("FAKE_OUT", out)

	s := New(func(ctx context.Context, provider string, o backend.Options) (backend.Session, error) {
		return backend.StartClaude(ctx, script, o)
	})
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	c, _, err := dial(t, ts.URL, http.Header{"Authorization": {"Bearer " + s.Token()}})
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	write(t, c, map[string]any{"op": "start", "chat": "c1", "provider": "claude", "cwd": dir})

	var requestID string
	for requestID == "" {
		m := read(t, c)
		if m["type"] == "permission" {
			if m["tool"] != "Bash" || m["chat"] != "c1" {
				t.Fatalf("permission = %v", m)
			}
			requestID = m["request"].(string)
		}
	}
	write(t, c, map[string]any{"op": "decide", "request": requestID, "allow": true})

	for {
		if m := read(t, c); m["type"] == "closed" {
			break
		}
	}
	got, _ := os.ReadFile(out)
	if !strings.Contains(string(got), `"behavior":"allow"`) {
		t.Fatalf("CLI received %s", got)
	}

	write(t, c, map[string]any{"op": "decide", "request": requestID, "allow": true})
	if m := read(t, c); m["type"] != "error" {
		t.Fatalf("second answer was accepted: %v", m)
	}
}
