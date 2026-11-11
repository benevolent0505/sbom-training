package graph

import "github.com/benevolent0505/sbom-training/api/internal/blog"

//go:generate go tool gqlgen generate

// Resolver はリゾルバが共有する依存を保持する。
type Resolver struct {
	Store blog.Store
}
