package graph

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/benevolent0505/sbom-training/api/internal/blog"
)

// NewHandler は GraphQL エンドポイントの http.Handler を返す。
func NewHandler(store blog.Store) http.Handler {
	srv := handler.New(NewExecutableSchema(Config{Resolvers: &Resolver{Store: store}}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	return srv
}

// maskInternal は入力エラー以外の詳細をログにだけ残し、クライアントには汎用メッセージを返す。
func maskInternal(ctx context.Context, field string, err error) error {
	if errors.Is(err, blog.ErrInvalidInput) {
		return err
	}
	slog.ErrorContext(ctx, "resolver failed", "field", field, "error", err)
	return errors.New("internal server error")
}
