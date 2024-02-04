package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/urfave/cli/v2"
)

func buildCMD() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build the Kite plugin in the current directory",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
			},
		},
		Action: func(c *cli.Context) error {
			basePath := c.Args().Get(0)
			if basePath == "" {
				basePath = "."
			}

			cfg, err := config.LoadworkspaceConfig(basePath)
			if err != nil {
				return err
			}

			return runBuild(basePath, c.Bool("debug"), cfg.Module)
		},
	}
}

func runBuild(basePath string, debug bool, cfg *config.ModuleConfig) error {
	switch cfg.Type {
	case "go":
		return runBuildGo(basePath, debug, cfg)
	case "rust":
		return runBuildRust(basePath, debug, cfg)
	case "js":
		return runBuildJS(basePath, debug, cfg)
	}

	return nil
}

func runBuildGo(basePath string, debug bool, cfg *config.ModuleConfig) error {
	basePath = filepath.Clean(basePath)
	outPath := filepath.Join(basePath, cfg.Build.Out)

	buildArgs := []string{"build", "-o", outPath, "-target=wasi", "-scheduler=none"}

	// Either optimize for size or for debugging (see https://tinygo.org/docs/guides/optimizing-binaries/)
	if !debug {
		buildArgs = append(buildArgs, "-no-debug", "-opt=z")
	} else {
		buildArgs = append(buildArgs, "-opt=1")
	}

	buildArgs = append(buildArgs, basePath)
	buildCMD := exec.Command("tinygo", buildArgs...)

	buildCMD.Stdout = os.Stdout
	buildCMD.Stderr = os.Stderr

	if err := buildCMD.Run(); err != nil {
		return err
	}

	if err := preInitializeWASM(outPath); err != nil {
		return err
	}

	if !debug {
		if err := optimizeWASM(outPath); err != nil {
			return err
		}
	}

	return nil
}

func runBuildRust(basePath string, debug bool, cfg *config.ModuleConfig) error {
	basePath = filepath.Clean(basePath)
	manifestPath := filepath.Join(basePath, "Cargo.toml")

	cmdArgs := []string{"build", "--target", "wasm32-unknown-unknown", "--manifest-path", manifestPath, "--out-dir", basePath}

	if !debug {
		cmdArgs = append(cmdArgs, "--release")
	}

	cmd := exec.Command("cargo", cmdArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if !debug {
		outPath := filepath.Join(basePath, cfg.Build.Out)

		if err := optimizeWASM(outPath); err != nil {
			return err
		}
	}

	return nil
}

func runBuildJS(basePath string, debug bool, cfg *config.ModuleConfig) error {
	basePath = filepath.Clean(basePath)

	bundle, err := bundleJS(basePath, debug, cfg)
	if err != nil {
		return err
	}
	fmt.Println("Bundle size: ", len(bundle))

	outPath := filepath.Join(basePath, cfg.Build.Out)

	cmdArgs := []string{"-", outPath}

	if !debug {
		cmdArgs = append(cmdArgs, "--optimize")
	}

	cmd := exec.Command("kitejs-compiler", cmdArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// pass bundle as stdin
	cmd.Stdin = bytes.NewReader(bundle)

	if err := cmd.Run(); err != nil {
		return err
	}

	// For JS plugins we always optimize the WASM because the users logic is in the JS
	if err := optimizeWASM(outPath); err != nil {
		return err
	}

	return nil
}

func bundleJS(basePath string, debug bool, cfg *config.ModuleConfig) ([]byte, error) {
	entryPoint, err := discoverJSEntrypoint(basePath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to discover entrypoint: %v", err)
	}

	absoluteBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute base path: %v", err)
	}

	result := api.Build(api.BuildOptions{
		AbsWorkingDir:     absoluteBasePath,
		EntryPoints:       []string{entryPoint},
		Bundle:            true,
		MinifyIdentifiers: !debug,
		MinifySyntax:      !debug,
		MinifyWhitespace:  !debug,
		TreeShaking:       api.TreeShakingTrue,
		Format:            api.FormatESModule,
		Write:             false,
	})
	if len(result.Errors) != 0 {
		return nil, fmt.Errorf("failed to bundle: %v", result.Errors)
	}

	return result.OutputFiles[0].Contents, nil
}

type PartialPackageJSON struct {
	Main string `json:"main"`
}

func discoverJSEntrypoint(basePath string, cfg *config.ModuleConfig) (string, error) {
	packageJSONPath := filepath.Join(basePath, "package.json")
	packageJSONRaw, err := os.ReadFile(packageJSONPath)
	if err == nil {
		var packageJSON PartialPackageJSON
		if err := json.Unmarshal(packageJSONRaw, &packageJSON); err != nil {
			return "", err
		}

		if packageJSON.Main != "" {
			return packageJSON.Main, nil
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}

	if cfg.Build.In != "" {
		return cfg.Build.In, nil
	}

	indexTSPath := filepath.Join(basePath, "index.ts")
	if _, err := os.Stat(indexTSPath); err == nil {
		return "index.ts", nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	indexJSPath := filepath.Join(basePath, "index.js")
	if _, err := os.Stat(indexJSPath); err == nil {
		return "index.js", nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	return "", fmt.Errorf("no entrypoint found")
}

func preInitializeWASM(path string) error {
	if !execCMDExists("wizer") {
		return fmt.Errorf("Wizer is not installed. Please install it using `cargo install wizer --all-features`")
	}

	tempPath := path + ".temp"
	initArgs := []string{path, "-o", tempPath, "--allow-wasi", "--wasm-bulk-memory=true", "--init-func=_start"}

	initCMD := exec.Command("wizer", initArgs...)

	initCMD.Stdout = os.Stdout
	initCMD.Stderr = os.Stderr

	if err := initCMD.Run(); err != nil {
		return err
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func optimizeWASM(path string) error {
	if execCMDExists("wasm-opt") {
		optArgs := []string{"--enable-bulk-memory", "-O3", "-o", path, path}

		optCMD := exec.Command("wasm-opt", optArgs...)

		optCMD.Stdout = os.Stdout
		optCMD.Stderr = os.Stderr

		if err := optCMD.Run(); err != nil {
			return err
		}
	}

	return nil
}

func execCMDExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
