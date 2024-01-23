import {
  useCompileJsMutation,
  useDeploymentCreateMutation,
} from "@/api/mutations";
import { compileDeployment } from "@/util/compile";
import { FlatFile } from "@/util/filetree";
import Editor, { useMonaco } from "@monaco-editor/react";
import { useEffect, useMemo, useRef, useState } from "react";

interface Props {
  files: FlatFile[];
  openFilePath: string | null;
}

export default function CodeEditor({ files, openFilePath }: Props) {
  const editorRef = useRef<any>(null);
  const monaco = useMonaco();

  const monacoIntitialized = useRef(false);
  useEffect(() => {
    if (!monaco || monacoIntitialized.current) return;
    monacoIntitialized.current = true;

    monaco.languages.typescript.typescriptDefaults.setCompilerOptions({
      target: monaco.languages.typescript.ScriptTarget.ES2016,
      allowNonTsExtensions: true,
      allowJs: true,
      moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
      module: monaco.languages.typescript.ModuleKind.CommonJS,
      noEmit: false,
      allowImportingTsExtensions: true,
      lib: [],
    });

    monaco.languages.typescript.typescriptDefaults.setDiagnosticsOptions({
      noSemanticValidation: false,
      noSyntaxValidation: false,
    });

    const libSource = `interface Event {type: string; data: any}; interface Call {type: string; data: any;}; declare class Kite {static call(call: Call); static handle(event: Event);};`;
    const libUri = "ts:filename/global.d.ts";

    monaco.languages.typescript.typescriptDefaults.setEagerModelSync(true);
    monaco.languages.typescript.typescriptDefaults.addExtraLib(
      libSource,
      libUri
    );
  }, [monaco]);

  useEffect(() => {
    if (!monaco) return;

    for (const file of files) {
      const uri = monaco.Uri.parse(`ts:filename/${file.path}`);

      if (monaco.editor.getModel(uri)) {
        //monaco.editor.getModel(uri).setValue(file.content);
        continue;
      } else {
        monaco.editor.createModel(
          file.content,
          "typescript",
          monaco.Uri.parse(`ts:filename/${file.path}`)
        );
      }
    }
  }, [monaco, files]);

  const compileMutation = useCompileJsMutation();
  const deployMutation = useDeploymentCreateMutation();

  useEffect(() => {
    console.log("yeet");
    async function onKeyDown(e: KeyboardEvent) {
      if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();

        if (!monaco) return;

        const models = monaco.editor.getModels();
        const files = models.map((model: any) => ({
          path: model.uri.path.slice(9),
          content: model.getValue(),
        }));

        const res = await compileDeployment(files, "index.ts");

        compileMutation.mutate(
          {
            source: res,
          },
          {
            onSuccess: (res) => {
              if (!res.success) {
                console.error(res.error);
                return;
              }

              deployMutation.mutate(
                {
                  key: "default@web",
                  name: "Default Plugin",
                  description: "The default plugin in the web editor",
                  wasm_bytes: res.data.wasm_bytes,
                  guild_id: "615613572164091914",
                  plugin_version_id: null,
                  manifest_default_config: {},
                  config: {},
                  manifest_events: ["DISCORD_MESSAGE_CREATE"],
                  manifest_commands: [],
                },
                {
                  onSuccess: (res) => {
                    console.log(res);
                  },
                  onError: (err) => {
                    console.error(err);
                  },
                }
              );
            },
            onError: (err) => {
              console.error(err);
            },
          }
        );
      }
    }

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [monaco, editorRef.current]);

  function showOpenFile(monaco: any, editor: any) {
    if (openFilePath) {
      const model = monaco.editor.getModel(
        monaco.Uri.parse(`ts:filename/${openFilePath}`)
      );

      if (model) {
        editorRef.current.setModel(model);
      }
    }
  }

  useEffect(() => {
    if (!monaco || !editorRef.current) return;

    showOpenFile(monaco, editorRef.current);
  }, [openFilePath, monaco, editorRef.current]);

  function handleEditorDidMount(editor: any, monaco: any) {
    editorRef.current = editor;
    showOpenFile(monaco, editor);
  }

  if (!openFilePath) {
    return <div></div>;
  }

  return (
    <Editor
      height="100%"
      width="100%"
      theme="vs-dark"
      loading={<div />}
      onMount={handleEditorDidMount}
    />
  );
}
