package perm

import (
	"context"
	"testing"
	"time"
)

func ask(b *Broker, ctx context.Context) (<-chan Decision, <-chan Request) {
	reqs := make(chan Request, 1)
	out := make(chan Decision, 1)
	go func() { out <- b.Ask(ctx, Request{Tool: "Bash"}, func(r Request) { reqs <- r }) }()
	return out, reqs
}

func TestAllowAndDeny(t *testing.T) {
	b := New()
	out, reqs := ask(b, context.Background())
	r := <-reqs
	if !b.Decide(r.ID, Decision{Allow: true}) {
		t.Fatal("decide returned false")
	}
	if d := <-out; !d.Allow {
		t.Fatal("expected allow")
	}
}

func TestDefaultsToDenyOnCancel(t *testing.T) {
	b := New()
	ctx, cancel := context.WithCancel(context.Background())
	out, reqs := ask(b, ctx)
	<-reqs
	cancel()
	select {
	case d := <-out:
		if d.Allow {
			t.Fatal("cancelled request must not allow")
		}
	case <-time.After(time.Second):
		t.Fatal("Ask did not return")
	}
}

func TestUnknownAndRepeatedIDsAreRejected(t *testing.T) {
	b := New()
	if b.Decide("nope", Decision{Allow: true}) {
		t.Fatal("unknown id accepted")
	}
	out, reqs := ask(b, context.Background())
	r := <-reqs
	b.Decide(r.ID, Decision{})
	<-out
	if b.Decide(r.ID, Decision{Allow: true}) {
		t.Fatal("answered request accepted twice")
	}
}
