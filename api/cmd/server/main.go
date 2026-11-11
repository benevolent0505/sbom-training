// server はブログの GraphQL API サーバー。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/benevolent0505/sbom-training/api/graph"
	"github.com/benevolent0505/sbom-training/api/internal/blog"
)

// version はビルド時に -ldflags "-X main.version=..." で埋め込む。
var version = "dev"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// ECS はタスク停止時に SIGTERM を送るので、受け取ったら処理中のリクエストを待って終了する
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, closeStore, err := newStore(ctx)
	if err != nil {
		return err
	}
	defer closeStore()

	mux := http.NewServeMux()
	mux.Handle("/query", graph.NewHandler(store))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	if os.Getenv("ENABLE_PLAYGROUND") == "true" {
		mux.Handle("GET /{$}", playground.Handler("Blog API", "/query"))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	slog.Info("listening", "addr", srv.Addr, "version", version)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func newStore(ctx context.Context) (blog.Store, func(), error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Warn("DATABASE_URL is not set; using in-memory store")
		return blog.NewMemoryStore(), func() {}, nil
	}
	store, err := blog.NewPostgresStore(ctx, databaseURL)
	if err != nil {
		return nil, nil, err
	}
	return store, store.Close, nil
}
