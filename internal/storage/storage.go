package storage

// URLStorage defines the interface for URL storage operations.
// This allows swapping implementations (sqlite, postgres, redis, etc.)
type URLStorage interface {
	SaveURL(url string, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
}
