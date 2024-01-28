import * as esbuild from "esbuild-wasm";
import path from "path";
import { FlatFile } from "./filetree";
import toml from "toml";

function customResolver(tree: Record<string, string>): esbuild.Plugin {
  const map = new Map(Object.entries(tree));

  return {
    name: "example",

    setup: (build: esbuild.PluginBuild) => {
      build.onResolve({ filter: /.*/ }, (args: esbuild.OnResolveArgs) => {
        if (args.kind === "entry-point") {
          return { path: "/" + args.path };
        }

        if (args.kind === "import-statement") {
          const dir = path.dirname(args.importer);

          const filePath = path.join(dir, args.path);

          return { path: filePath };
        }

        throw Error("not resolvable");
      });

      build.onLoad({ filter: /.*/ }, (args: esbuild.OnLoadArgs) => {
        if (!map.has(args.path)) {
          throw Error("not loadable");
        }
        const ext = path.extname(args.path);

        const contents = map.get(args.path)!;

        const loader =
          ext === ".ts"
            ? "ts"
            : ext === ".tsx"
            ? "tsx"
            : ext === ".js"
            ? "js"
            : ext === ".jsx"
            ? "jsx"
            : "default";

        return { contents, loader };
      });
    },
  };
}

let initialized = false;

async function initialize() {
  if (initialized) return;
  initialized = true;

  await esbuild.initialize({
    wasmURL: "/app/esbuild.wasm",
  });
}

export async function compileWorkspace(files: FlatFile[], entry: string) {
  await initialize();

  const plugin = customResolver(
    Object.fromEntries(
      files.map((f) => [
        f.path.startsWith("/") ? f.path : "/" + f.path,
        f.content,
      ])
    )
  );

  let result2 = await esbuild.build({
    entryPoints: [entry],
    bundle: true,
    minify: true,
    write: false,
    treeShaking: true,
    format: "esm",
    plugins: [plugin],
  });

  return result2.outputFiles[0].text;
}

export function readManifestFromWorkspace(files: FlatFile[]) {
  const manifestFile = files.find((f) => f.path === "manifest.toml");
  if (!manifestFile) return null;

  try {
    return toml.parse(manifestFile.content);
  } catch (e) {
    console.log(e);
    return null;
  }
}
