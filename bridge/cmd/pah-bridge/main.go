// pah-bridge connects the Pocket AI Harness app to agent CLIs on the same device.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/DeepuExplainsAi/pocketAiHarness/bridge/internal/backend"
	"github.com/DeepuExplainsAi/pocketAiHarness/bridge/internal/server"
)

func main() {
	addr := flag.String("listen", "127.0.0.1:0", "loopback address to listen on")
	claudeBin := flag.String("claude", "claude", "path to the claude CLI")
	agyBin := flag.String("agy", "agy", "path to the agy CLI")
	flag.Parse()

	srv := server.New(func(ctx context.Context, provider string, o backend.Options) (backend.Session, error) {
		switch provider {
		case "claude":
			return backend.StartClaude(ctx, *claudeBin, o)
		case "agy":
			return backend.StartAgy(ctx, *agyBin, o)
		}
		return nil, fmt.Errorf("unknown provider %q", provider)
	})

	ln, err := server.Listen(*addr)
	if err != nil {
		log.Fatal(err)
	}
	// The app reads this single line to learn the port and the per-run token.
	fmt.Printf("{\"listen\":%q,\"token\":%q}\n", ln.Addr().String(), srv.Token())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	hs := &http.Server{Handler: srv.Handler()}
	go func() { <-ctx.Done(); _ = hs.Shutdown(context.Background()) }()
	if err := hs.Serve(ln); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
