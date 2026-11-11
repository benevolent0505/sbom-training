// Package blog はブログ記事のドメイン型と永続化のインターフェースを定義する。
package blog

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrInvalidInput は入力値が不正な場合に返される。
var ErrInvalidInput = errors.New("invalid input")

// Post はブログ記事。
type Post struct {
	ID        string
	Title     string
	Body      string
	Author    string
	CreatedAt time.Time
}

// NewPost は記事作成時の入力。
type NewPost struct {
	Title  string
	Body   string
	Author string
}

// Validate は前後の空白を取り除いたうえで必須項目を検証する。
func (n *NewPost) Validate() error {
	n.Title = strings.TrimSpace(n.Title)
	n.Body = strings.TrimSpace(n.Body)
	n.Author = strings.TrimSpace(n.Author)
	switch {
	case n.Title == "":
		return errors.Join(ErrInvalidInput, errors.New("title is required"))
	case n.Body == "":
		return errors.Join(ErrInvalidInput, errors.New("body is required"))
	case n.Author == "":
		return errors.Join(ErrInvalidInput, errors.New("author is required"))
	}
	return nil
}

// Store は記事の永続化先。
type Store interface {
	// List は作成日時の新しい順に記事を返す。
	List(ctx context.Context) ([]*Post, error)
	// Get は ID に一致する記事を返す。存在しない場合は (nil, nil)。
	Get(ctx context.Context, id string) (*Post, error)
	// Create は記事を保存し、採番された ID を含む記事を返す。
	Create(ctx context.Context, input NewPost) (*Post, error)
}
