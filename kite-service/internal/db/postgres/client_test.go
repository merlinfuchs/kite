package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestConstraintErrors(t *testing.T) {
	wrap := func(code string) error { return fmt.Errorf("query: %w", &pgconn.PgError{Code: code}) }

	assert.ErrorIs(t, uniqueViolation(wrap("23505")), store.ErrAlreadyExists)
	assert.ErrorIs(t, foreignKeyViolation(wrap("23503")), store.ErrNotFound)
	// Other errors stay as they are.
	other := errors.New("connection reset")
	assert.Equal(t, other, foreignKeyViolation(other))
	assert.Nil(t, foreignKeyViolation(nil))
}
