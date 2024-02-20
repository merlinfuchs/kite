package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gopkg.in/guregu/null.v4"
)

func timeToTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}

func nullStringToText(s null.String) pgtype.Text {
	return pgtype.Text{
		String: s.String,
		Valid:  s.Valid,
	}
}

func stringToText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

func textToNullString(t pgtype.Text) null.String {
	return null.NewString(t.String, t.Valid)
}

func timestampToNullTime(t pgtype.Timestamp) null.Time {
	return null.NewTime(t.Time, t.Valid)
}
