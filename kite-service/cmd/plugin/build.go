package plugin

import (
	"os"
	"os/exec"
	"path/filepath"

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
			basePath := c.String("path")

			cfg, err := config.LoadPluginConfig(basePath)
			if err != nil {
				return err
			}

			return runBuild(basePath, c.Bool("debug"), cfg)
		},
	}
}

func runBuild(basePath string, debug bool, cfg *config.PluginConfig) error {
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

func runBuildGo(basePath string, debug bool, cfg *config.PluginConfig) error {
	basePath = filepath.Clean(basePath)
	outPath := filepath.Join(basePath, cfg.Build.Out)

	buildArgs := []string{"build", "-o", outPath, "-target=wasi", "-scheduler=none"}

	// Either optimize for size or for debugging (see https://tinygo.org/docs/guides/optimizing-binaries/)
	if !debug {
		buildArgs = append(buildArgs, "-no-debug")
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

	if !debug {
		if err := optimizeWASM(outPath); err != nil {
			return err
		}
	}

	return nil
}

func runBuildRust(basePath string, debug bool, cfg *config.PluginConfig) error {
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

func optimizeWASM(path string) error {
	if execCMDExists("wasm-opt") {
		optArgs := []string{"-Oz", "-o", path, path}

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

func runBuildJS(basePath string, debug bool, cfg *config.PluginConfig) error {
	// https://github.com/bytecodealliance/javy
	// https://docs.rs/quickjs-wasm-rs/latest/quickjs_wasm_rs/
	// https://github.com/seddonm1/quickjs
	// https://github.com/extism/js-pdk
	// https://github.com/tetratelabs/wazero/issues/1580

	return nil
}
