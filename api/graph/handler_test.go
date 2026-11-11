package graph_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benevolent0505/sbom-training/api/graph"
	"github.com/benevolent0505/sbom-training/api/internal/blog"
)

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func do(t *testing.T, h http.Handler, query string, variables map[string]any) gqlResponse {
	t.Helper()
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var resp gqlResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return resp
}

func TestCreateAndQueryPosts(t *testing.T) {
	h := graph.NewHandler(blog.NewMemoryStore())

	created := do(t, h,
		`mutation($input: NewPost!) { createPost(input: $input) { id title author createdAt } }`,
		map[string]any{"input": map[string]any{"title": "Hello SBOM", "body": "本文", "author": "alice"}})
	if len(created.Errors) > 0 {
		t.Fatalf("createPost errors: %+v", created.Errors)
	}
	var createData struct {
		CreatePost struct{ ID, Title, Author, CreatedAt string }
	}
	if err := json.Unmarshal(created.Data, &createData); err != nil {
		t.Fatal(err)
	}
	id := createData.CreatePost.ID
	if id == "" || createData.CreatePost.Title != "Hello SBOM" || createData.CreatePost.CreatedAt == "" {
		t.Fatalf("unexpected createPost payload: %s", created.Data)
	}

	listed := do(t, h, `{ posts { id title } }`, nil)
	var listData struct {
		Posts []struct{ ID, Title string }
	}
	if err := json.Unmarshal(listed.Data, &listData); err != nil {
		t.Fatal(err)
	}
	if len(listData.Posts) != 1 || listData.Posts[0].ID != id {
		t.Errorf("posts = %s, want the created post", listed.Data)
	}

	single := do(t, h, `query($id: ID!) { post(id: $id) { body } }`, map[string]any{"id": id})
	if string(single.Data) != `{"post":{"body":"本文"}}` {
		t.Errorf("post = %s", single.Data)
	}

	missing := do(t, h, `{ post(id: "missing") { id } }`, nil)
	if string(missing.Data) != `{"post":null}` || len(missing.Errors) != 0 {
		t.Errorf("missing post = %s, errors = %+v", missing.Data, missing.Errors)
	}
}

func TestCreatePostRejectsInvalidInput(t *testing.T) {
	h := graph.NewHandler(blog.NewMemoryStore())
	resp := do(t, h,
		`mutation { createPost(input: {title: " ", body: "b", author: "a"}) { id } }`, nil)
	if len(resp.Errors) != 1 {
		t.Fatalf("errors = %+v, want exactly one", resp.Errors)
	}
	if got := resp.Errors[0].Message; got != "invalid input\ntitle is required" {
		t.Errorf("message = %q", got)
	}
}

type failingStore struct{ blog.Store }

func (failingStore) List(context.Context) ([]*blog.Post, error) {
	return nil, errors.New("pq: password authentication failed for user admin")
}

func TestInternalErrorsAreMasked(t *testing.T) {
	h := graph.NewHandler(failingStore{})
	resp := do(t, h, `{ posts { id } }`, nil)
	if len(resp.Errors) != 1 || resp.Errors[0].Message != "internal server error" {
		t.Errorf("errors = %+v, want masked internal error", resp.Errors)
	}
}
