package storage

import "fmt"

var (
	ErrAliasExists = fmt.Errorf("alias already exists")
	ErrNotFound    = fmt.Errorf("url not found")
)
