package storage

import "fmt"

var (
	ErrAliasExists = fmt.Errorf("alias already exists")
	ErrNotFound    = fmt.Errorf("url not found")
)

// URLStorage defines the interface for URL storage operations.
// This allows swapping implementations (sqlite, postgres, redis, etc.)
type URLStorage interface {
	SaveURL(url string, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
}
