package module

import (
	"context"
	"fmt"
	"time"

	"github.com/merlinfuchs/kite/kite-types/manifest"
)

func (p *Module) getAPIVersion(ctx context.Context) (manifest.APIVersion, error) {
	fn := p.m.ExportedFunction("kite_get_api_version")
	if fn == nil {
		return 0, fmt.Errorf("kite_get_api_version not defined")
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	res, err := fn.Call(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to call kite_get_api_version: %w", err)
	}

	if len(res) != 1 {
		return 0, fmt.Errorf("kite_get_api_version returned invalid number of results")
	}

	return manifest.APIVersion(res[0]), nil
}

func (p *Module) getAPIEncoding(ctx context.Context) (manifest.APIEncoding, error) {
	fn := p.m.ExportedFunction("kite_get_api_encoding")
	if fn == nil {
		return 0, fmt.Errorf("kite_get_api_encoding not defined")
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	res, err := fn.Call(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to call kite_get_api_encoding: %w", err)
	}

	if len(res) != 1 {
		return 0, fmt.Errorf("kite_get_api_encoding returned invalid number of results")
	}

	return manifest.APIEncoding(res[0]), nil
}

func (p *Module) GetManifest(ctx context.Context) (*manifest.Manifest, error) {
	fn := p.m.ExportedFunction("kite_describe")
	if fn == nil {
		return nil, fmt.Errorf("kite_describe not defined")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	_, err := fn.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to call kite_describe: %w", err)
	}

	return &p.currentManifest, nil
}
