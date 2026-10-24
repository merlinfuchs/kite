package postgres

import (
	"fmt"
	"log/slog"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRowsSkipsUnreadableRows(t *testing.T) {
	rows := []string{"1", "two", "3"}
	res := decodeRows(rows, strconv.Atoi, func(row string) []any {
		return []any{slog.String("row", row)}
	})
	assert.Equal(t, []int{1, 3}, res)
}

func TestDecodeRowsKeepsOrder(t *testing.T) {
	rows := []int{3, 1, 2}
	res := decodeRows(rows, func(row int) (string, error) {
		return fmt.Sprint(row), nil
	}, func(int) []any { return nil })
	assert.Equal(t, []string{"3", "1", "2"}, res)
}
