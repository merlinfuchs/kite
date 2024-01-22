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

    const libSource = `interface Call {name: string}; declare class Kite {static makeCall(call: Call)};`;
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

  useEffect(() => {
    async function onKeyDown(e: KeyboardEvent) {
      if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();

        console.log("save");

        if (!monaco) return;

        const models = monaco.editor.getModels();
        const files = models.map((model: any) => ({
          path: model.uri.path.slice(9),
          content: model.getValue(),
        }));

        console.log(files);

        const res = await compileDeployment(files, "index.ts");
        console.log(res);
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
