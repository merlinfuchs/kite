import { FlatFile } from "@/util/filetree";
import { initializeMonaco } from "@/util/monaco";
import Editor, { useMonaco } from "@monaco-editor/react";
import { useEffect, useRef, useState } from "react";

interface Props {
  files: FlatFile[];
  openFilePath: string | null;
  onChange: () => void;
  onSave: () => void;
  onDeploy: () => void;
}

export default function CodeEditor({
  files,
  openFilePath,
  onChange,
  onSave,
  onDeploy,
}: Props) {
  const editorRef = useRef<any>(null);
  const monaco = useMonaco();

  const monacoIntitialized = useRef(false);
  useEffect(() => {
    if (!monaco || monacoIntitialized.current) return;
    monacoIntitialized.current = true;
    initializeMonaco(monaco);
  }, [monaco]);

  useEffect(() => {
    if (!monaco) return;

    const existingUris = [];
    for (const file of files) {
      const uri = monaco.Uri.parse(`ts:filename/${file.path}`);
      existingUris.push(uri.toString());

      if (monaco.editor.getModel(uri)) {
        monaco.editor.getModel(uri).setValue(file.content);
        continue;
      } else {
        monaco.editor.createModel(
          file.content,
          "typescript",
          monaco.Uri.parse(`ts:filename/${file.path}`)
        );
      }
    }

    for (const model of monaco.editor.getModels()) {
      if (!existingUris.includes(model.uri.toString())) {
        model.dispose();
      }
    }
  }, [monaco, files]);

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

  useEffect(() => {
    async function onKeyDown(e: KeyboardEvent) {
      if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();

        const editor = editorRef.current;
        if (editor) {
          editor.getAction("editor.action.formatDocument").run();
        }

        // TODO: debounce
        setTimeout(() => {
          onSave();
        }, 100);
      }

      if (e.key === "p" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();

        const editor = editorRef.current;
        if (editor) {
          editor.getAction("editor.action.formatDocument").run();
        }

        // TODO: debounce
        setTimeout(() => {
          onSave();
          onDeploy();
        }, 100);
      }
    }

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onSave]);

  function handleEditorDidMount(editor: any, monaco: any) {
    editorRef.current = editor;
    showOpenFile(monaco, editor);
  }

  function handleEditorValueChange(value: string | undefined, e: any) {
    // isFlush is true when the change comes from outside (e.g. open file changed)
    // if (e.isFlush) return;
    const openFile = files.find((f) => f.path === openFilePath);
    if (!openFile) return;

    // We intentionally don't trigger a state update here, because we don't want to re-render the editor
    // Monaco will handle the state of the editor, this is just for saving the changes
    openFile.content = value || "";

    onChange();
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
      onChange={handleEditorValueChange}
    />
  );
}
