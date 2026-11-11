package blog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS posts (
	id         uuid PRIMARY KEY,
	title      text NOT NULL,
	body       text NOT NULL,
	author     text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
)`

// PostgresStore は PostgreSQL に記事を保存する Store。
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore は接続を確立し、テーブルがなければ作成する。
func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() { s.pool.Close() }

func (s *PostgresStore) List(ctx context.Context) ([]*Post, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, title, body, author, created_at FROM posts ORDER BY created_at DESC, id`)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	return pgx.CollectRows(rows, scanPost)
}

func (s *PostgresStore) Get(ctx context.Context, id string) (*Post, error) {
	// uuid として解釈できない ID は「存在しない」として扱う
	if _, err := uuid.Parse(id); err != nil {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, title, body, author, created_at FROM posts WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	p, err := pgx.CollectExactlyOneRow(rows, scanPost)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (s *PostgresStore) Create(ctx context.Context, input NewPost) (*Post, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`INSERT INTO posts (id, title, body, author) VALUES ($1, $2, $3, $4)
		 RETURNING id::text, title, body, author, created_at`,
		uuid.NewString(), input.Title, input.Body, input.Author)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	return pgx.CollectExactlyOneRow(rows, scanPost)
}

func scanPost(row pgx.CollectableRow) (*Post, error) {
	var p Post
	if err := row.Scan(&p.ID, &p.Title, &p.Body, &p.Author, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}
