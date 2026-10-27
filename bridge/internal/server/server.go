// Package server exposes chat sessions to the app over an authenticated local WebSocket.
package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/coder/websocket"

	"github.com/DeepuExplainsAi/pocketAiHarness/bridge/internal/backend"
	"github.com/DeepuExplainsAi/pocketAiHarness/bridge/internal/guard"
	"github.com/DeepuExplainsAi/pocketAiHarness/bridge/internal/perm"
)

// Starter creates a session for a provider. Unknown providers return an error.
type Starter func(ctx context.Context, provider string, o backend.Options) (backend.Session, error)

type Server struct {
	token string
	start Starter
	perms *perm.Broker
}

func New(start Starter) *Server {
	var raw [24]byte
	_, _ = rand.Read(raw[:])
	return &Server{token: hex.EncodeToString(raw[:]), start: start, perms: perm.New()}
}

func (s *Server) Token() string { return s.token }

// Listen binds to loopback only. Any non-loopback address is refused.
func Listen(addr string) (net.Listener, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		return nil, errors.New("bridge only listens on a loopback address")
	}
	return net.Listen("tcp", addr)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.ws)
	return mux
}

type inbound struct {
	Op       string `json:"op"`
	Chat     string `json:"chat"`
	Provider string `json:"provider"`
	Cwd      string `json:"cwd"`
	Model    string `json:"model"`
	Mode     string `json:"mode"`
	Resume   string `json:"resume"`
	Text     string `json:"text"`
	Request  string `json:"request"`
	Allow    bool   `json:"allow"`
	Reason   string `json:"reason"`
}

func (s *Server) authorised(r *http.Request) bool {
	if r.Header.Get("Origin") != "" { // browsers always send Origin; the native app does not
		return false
	}
	got := r.Header.Get("Authorization")
	want := "Bearer " + s.token
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	if !s.authorised(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var wmu sync.Mutex
	send := func(v any) {
		b, _ := json.Marshal(v)
		wmu.Lock()
		defer wmu.Unlock()
		_ = c.Write(ctx, websocket.MessageText, b)
	}

	sessions := map[string]backend.Session{}
	defer func() {
		for _, sess := range sessions {
			_ = sess.Close()
		}
	}()

	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		var m inbound
		if json.Unmarshal(data, &m) != nil {
			send(map[string]any{"type": "error", "message": "invalid message"})
			continue
		}
		s.dispatch(ctx, m, sessions, send)
	}
}

func (s *Server) dispatch(ctx context.Context, m inbound, sessions map[string]backend.Session, send func(any)) {
	fail := func(err error) {
		send(map[string]any{"type": "error", "chat": m.Chat, "message": guard.Redact(err.Error())})
	}

	switch m.Op {
	case "start":
		if _, dup := sessions[m.Chat]; dup || m.Chat == "" {
			fail(errors.New("chat id missing or already in use"))
			return
		}
		chat := m.Chat
		ask := func(ctx context.Context, r backend.ToolRequest) backend.Decision {
			d := s.perms.Ask(ctx, perm.Request{Chat: chat, Tool: r.Tool, Description: r.Description, Input: r.Input},
				func(p perm.Request) {
					send(map[string]any{"type": "permission", "chat": chat, "request": p.ID, "tool": p.Tool, "description": p.Description, "input": p.Input})
				})
			return backend.Decision{Allow: d.Allow, Reason: d.Reason}
		}
		sess, err := s.start(ctx, m.Provider, backend.Options{Cwd: m.Cwd, Model: m.Model, Mode: m.Mode, Resume: m.Resume, Ask: ask})
		if err != nil {
			fail(err)
			return
		}
		sessions[chat] = sess
		send(map[string]any{"type": "started", "chat": chat})
		go func() {
			for e := range sess.Events() {
				send(map[string]any{"type": "event", "chat": chat, "seq": e.Seq, "raw": e.Raw})
			}
			send(map[string]any{"type": "closed", "chat": chat})
		}()
	case "decide":
		if !s.perms.Decide(m.Request, perm.Decision{Allow: m.Allow, Reason: m.Reason}) {
			fail(errors.New("unknown or already answered request"))
		}
	case "send", "interrupt", "mode", "model", "close":
		sess, ok := sessions[m.Chat]
		if !ok {
			fail(errors.New("unknown chat"))
			return
		}
		var err error
		switch m.Op {
		case "send":
			err = sess.Send(m.Text)
		case "interrupt":
			err = sess.Interrupt()
		case "mode":
			err = sess.SetMode(m.Mode)
		case "model":
			err = sess.SetModel(m.Model)
		case "close":
			err = sess.Close()
			delete(sessions, m.Chat)
		}
		if err != nil {
			fail(err)
		}
	default:
		fail(errors.New("unknown operation"))
	}
}
