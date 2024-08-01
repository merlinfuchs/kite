package template

import (
	"bytes"
	"fmt"
	"io"
	"maps"

	"github.com/botlabs-gg/yagpdb/v2/lib/template"
)

const DefaultMaxOps = 1000
const DefaultMaxOutput = 4000

const DelimLeft = "{{"
const DelimRight = "}}"

type TemplateContext struct {
	name  string
	data  map[string]interface{}
	funcs map[string]interface{}

	MaxOps    int
	MaxOutput int64
}

func NewContext(name string, maxOps int, providers ...ContextProvider) *TemplateContext {
	data := make(map[string]interface{}, len(standardDataMap))
	maps.Copy(data, standardDataMap)

	funcs := make(map[string]interface{}, len(standardFuncMap))
	maps.Copy(funcs, standardFuncMap)

	for _, provider := range providers {
		provider.ProvideData(data)
		provider.ProvideFuncs(funcs)
	}

	if maxOps == 0 {
		maxOps = DefaultMaxOps
	}

	return &TemplateContext{
		name:  name,
		data:  data,
		funcs: funcs,

		MaxOps:    maxOps,
		MaxOutput: DefaultMaxOutput,
	}
}

func (c *TemplateContext) ParseAndExecute(text string) (string, error) {
	if text == "" || !ContainsVariable(text) {
		return text, nil
	}

	tmpl, err := c.Parse(text)
	if err != nil {
		return "", err
	}

	return c.Execute(tmpl)
}

func (c *TemplateContext) Parse(text string) (*template.Template, error) {
	return template.New(c.name).
		Delims(DelimLeft, DelimRight).
		Funcs(c.funcs).
		Parse(text)
}

func (c *TemplateContext) Execute(tmpl *template.Template) (string, error) {
	tmpl = tmpl.MaxOps(c.MaxOps)

	var buf bytes.Buffer
	w := LimitWriter(&buf, c.MaxOutput)

	err := tmpl.Execute(w, c.data)
	if err != nil {
		if err == io.ErrShortWrite {
			err = fmt.Errorf("output exceeded %d characters", c.MaxOutput)
		}
		return "", err
	}

	res := buf.String()

	return res, nil
}

func (c *TemplateContext) AddProvider(provider ContextProvider) {
	provider.ProvideData(c.data)
	provider.ProvideFuncs(c.funcs)
}

func (c *TemplateContext) SetData(key string, value interface{}) {
	c.data[key] = value
}
