package variable

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryValueStore struct {
	// Methods the tests don't use panic.
	store.VariableValueStore

	values []model.VariableValue
	limit  int
	offset int
}

func (s *memoryValueStore) SearchVariableValues(ctx context.Context, appID string, variableID string, search string, limit int, offset int) ([]*model.VariableValue, int, error) {
	s.limit = limit
	s.offset = offset
	return nil, 0, nil
}

func (s *memoryValueStore) UpdateVariableValue(ctx context.Context, appID string, operation model.VariableValueOperation, value model.VariableValue) (*model.VariableValue, error) {
	s.values = append(s.values, value)
	return &value, nil
}

func request(t *testing.T, variable *model.Variable, method string, target string, body string, h handler.HandlerFunc) (int, map[string]any) {
	t.Helper()

	srv := handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		c.Variable = variable
		return h(c)
	})

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func errorCode(res map[string]any) string {
	return res["error"].(map[string]any)["code"].(string)
}

func TestVariableValueListPaging(t *testing.T) {
	s := &memoryValueStore{}
	h := NewVariableHandler(nil, s)
	list := handler.Typed(h.HandleVariableValueList)
	variable := &model.Variable{ID: "var", Scoped: true}

	code, res := request(t, variable, http.MethodGet, "/values", "", list)
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, 25, s.limit)
	assert.Equal(t, 0, s.offset)

	code, res = request(t, variable, http.MethodGet, "/values?limit=100&offset=50", "", list)
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, 100, s.limit)
	assert.Equal(t, 50, s.offset)

	for _, limit := range []string{"0", "101", "-1", "abc", "1.5"} {
		code, res = request(t, variable, http.MethodGet, "/values?limit="+limit, "", list)
		assert.Equal(t, http.StatusBadRequest, code, limit)
		assert.Equal(t, "invalid_limit", errorCode(res), limit)
	}

	for _, offset := range []string{"-1", "abc"} {
		code, res = request(t, variable, http.MethodGet, "/values?offset="+offset, "", list)
		assert.Equal(t, http.StatusBadRequest, code, offset)
		assert.Equal(t, "invalid_offset", errorCode(res), offset)
	}
}

func TestVariableValueSetScope(t *testing.T) {
	s := &memoryValueStore{}
	h := NewVariableHandler(nil, s)
	set := handler.TypedWithBody(h.HandleVariableValueSet)
	scoped := &model.Variable{ID: "var", Scoped: true}
	unscoped := &model.Variable{ID: "var"}

	// A scoped variable needs a scope, whitespace doesn't count.
	for _, scope := range []string{"", "   "} {
		code, res := request(t, scoped, http.MethodPut, "/values", `{"scope":"`+scope+`","type":"string","value":"a"}`, set)
		assert.Equal(t, http.StatusBadRequest, code, scope)
		assert.Equal(t, "validation_failed", errorCode(res), scope)
	}

	code, res := request(t, unscoped, http.MethodPut, "/values", `{"scope":"123","type":"string","value":"a"}`, set)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "validation_failed", errorCode(res))
	assert.Empty(t, s.values)

	code, res = request(t, scoped, http.MethodPut, "/values", `{"scope":" 123 ","type":"string","value":"a"}`, set)
	require.Equal(t, http.StatusOK, code, res)
	require.Len(t, s.values, 1)
	assert.Equal(t, "123", s.values[0].Scope.String)

	// Whitespace on an unscoped variable is no scope.
	code, res = request(t, unscoped, http.MethodPut, "/values", `{"scope":" ","type":"string","value":"a"}`, set)
	require.Equal(t, http.StatusOK, code, res)
	require.Len(t, s.values, 2)
	assert.False(t, s.values[1].Scope.Valid)
}
