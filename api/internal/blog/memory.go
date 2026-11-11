package blog

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore はプロセス内メモリに記事を保持する Store。
// DATABASE_URL が未設定のローカル実行とテストで使う。
type MemoryStore struct {
	mu    sync.RWMutex
	posts []*Post
	now   func() time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{now: time.Now}
}

func (s *MemoryStore) List(context.Context) ([]*Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := slices.Clone(s.posts)
	slices.Reverse(out)
	return out, nil
}

func (s *MemoryStore) Get(_ context.Context, id string) (*Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.posts {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) Create(_ context.Context, input NewPost) (*Post, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	p := &Post{
		ID:        uuid.NewString(),
		Title:     input.Title,
		Body:      input.Body,
		Author:    input.Author,
		CreatedAt: s.now().UTC(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.posts = append(s.posts, p)
	return p, nil
}
