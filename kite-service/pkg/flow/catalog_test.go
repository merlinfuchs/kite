package flow

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type catalogSchema struct {
	Properties map[string]*catalogSchema `json:"properties"`
	Items      *catalogSchema            `json:"items"`
	Enum       []any                     `json:"enum"`
}

func loadCatalog(t *testing.T) map[string]*catalogSchema {
	var catalog struct {
		Nodes map[string]struct {
			DataSchema *catalogSchema `json:"data_schema"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal(CatalogJSON, &catalog))

	res := make(map[string]*catalogSchema, len(catalog.Nodes))
	for nodeType, node := range catalog.Nodes {
		res[nodeType] = node.DataSchema
	}
	return res
}

func TestCatalogHasEveryNodeType(t *testing.T) {
	catalog := loadCatalog(t)
	nodeTypes := flowNodeTypeConstants(t)

	for _, nodeType := range nodeTypes {
		assert.Contains(t, catalog, nodeType)
	}
	for nodeType := range catalog {
		assert.Contains(t, nodeTypes, nodeType)
	}
}

// The editor must not store settings the service doesn't read.
func TestCatalogFieldsExist(t *testing.T) {
	for nodeType, schema := range loadCatalog(t) {
		if schema != nil {
			checkCatalogFields(t, nodeType, schema, reflect.TypeOf(FlowNodeData{}))
		}
	}
}

func checkCatalogFields(t *testing.T, path string, schema *catalogSchema, goType reflect.Type) {
	for goType.Kind() == reflect.Pointer || goType.Kind() == reflect.Slice {
		goType = goType.Elem()
	}
	if goType.Kind() != reflect.Struct {
		return
	}

	if schema.Items != nil {
		checkCatalogFields(t, path+"[]", schema.Items, goType)
	}

	fields := make(map[string]reflect.Type)
	for i := range goType.NumField() {
		field := goType.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		fields[name] = field.Type
	}

	for name, prop := range schema.Properties {
		fieldType, ok := fields[name]
		if assert.Truef(t, ok, "%s.%s has no field in %s", path, name, goType) {
			checkCatalogFields(t, path+"."+name, prop, fieldType)
		}
	}
}

func flowNodeTypeConstants(t *testing.T) []string {
	file, err := parser.ParseFile(token.NewFileSet(), "data.go", nil, 0)
	require.NoError(t, err)

	var res []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value := spec.(*ast.ValueSpec)
			if ident, ok := value.Type.(*ast.Ident); !ok || ident.Name != "FlowNodeType" {
				continue
			}
			for _, v := range value.Values {
				s, err := strconv.Unquote(v.(*ast.BasicLit).Value)
				require.NoError(t, err)
				res = append(res, s)
			}
		}
	}
	return res
}
