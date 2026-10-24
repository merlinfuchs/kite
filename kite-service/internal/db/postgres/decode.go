package postgres

import "log/slog"

// decodeRows decodes the rows of a query that loads the rows of every app at
// once. Rows that can't be decoded are logged and skipped, so one of them,
// like a flow from an older version that's no longer read, can't stop the
// others from loading. They're loaded again once they're updated or the
// service restarts.
func decodeRows[R, M any](rows []R, decode func(R) (M, error), attrs func(R) []any) []M {
	res := make([]M, 0, len(rows))
	for _, row := range rows {
		m, err := decode(row)
		if err != nil {
			slog.Error(
				"Skipped row that can't be decoded",
				append(attrs(row), slog.String("error", err.Error()))...,
			)
			continue
		}
		res = append(res, m)
	}
	return res
}
