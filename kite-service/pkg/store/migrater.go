package store

import "github.com/golang-migrate/migrate/v4"

// Common migrater interface for stores.
type Migrater interface {
	Up() error
	Down() error
	To(version uint) error

	Force(version int) error
	Version() (uint, bool, error)
	List() ([]string, error)
	Close() error
	SetLogger(logger migrate.Logger)
}
