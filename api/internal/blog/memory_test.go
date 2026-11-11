package blog

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	tick := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { tick = tick.Add(time.Minute); return tick }

	first, err := s.Create(ctx, NewPost{Title: " first ", Body: "body", Author: "alice"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if first.Title != "first" {
		t.Errorf("Title = %q, want trimmed %q", first.Title, "first")
	}
	second, err := s.Create(ctx, NewPost{Title: "second", Body: "body", Author: "bob"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	posts, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(posts) != 2 || posts[0].ID != second.ID || posts[1].ID != first.ID {
		t.Errorf("List should return newest first, got %+v", posts)
	}

	got, err := s.Get(ctx, first.ID)
	if err != nil || got == nil || got.ID != first.ID {
		t.Errorf("Get(%q) = %+v, %v", first.ID, got, err)
	}
	if missing, err := s.Get(ctx, "nope"); err != nil || missing != nil {
		t.Errorf("Get(missing) = %+v, %v; want nil, nil", missing, err)
	}
}

func TestNewPostValidate(t *testing.T) {
	tests := map[string]NewPost{
		"empty title":  {Title: "  ", Body: "b", Author: "a"},
		"empty body":   {Title: "t", Body: "", Author: "a"},
		"empty author": {Title: "t", Body: "b", Author: "\t"},
	}
	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			if err := in.Validate(); !errors.Is(err, ErrInvalidInput) {
				t.Errorf("Validate() = %v, want ErrInvalidInput", err)
			}
		})
	}
}
