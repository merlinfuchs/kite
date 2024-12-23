package eval

import (
	"context"

	"gopkg.in/guregu/null.v4"
)

type VariableProvider interface {
	VariableValue(ctx context.Context, id string, scope null.String) (string, error)
}
