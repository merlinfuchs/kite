package jobs

import (
	"context"
	"log/slog"

	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type ErrorHandler struct{}

func (*ErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	slog.
		With(logattr.Error(err)).
		With("job_id", job.ID).
		With("job_kind", job.Kind).
		With("job_args", string(job.EncodedArgs)).
		Error("Job failed")
	return nil
}

func (*ErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any) *river.ErrorHandlerResult {
	slog.
		With("panic_value", panicVal).
		With("job_id", job.ID).
		With("job_kind", job.Kind).
		With("job_args", string(job.EncodedArgs)).
		Error("Job panicked")
	return nil
}
