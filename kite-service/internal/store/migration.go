package store

import "github.com/golang-migrate/migrate/v4"

type MigrationStore interface {
	Up() error
	Down() error
	To(version uint) error

	Force(version int) error
	Version() (uint, bool, error)
	List() ([]string, error)
	Close() error
	SetLogger(logger migrate.Logger)
}
