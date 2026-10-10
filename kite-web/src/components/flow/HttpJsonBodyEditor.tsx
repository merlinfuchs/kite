import { useCallback, useMemo, useRef } from "react";
import ReactCodeMirror, {
  EditorView,
  keymap,
  ReactCodeMirrorRef,
} from "@uiw/react-codemirror";
import { json } from "@codemirror/lang-json";
import { Diagnostic, linter, lintGutter } from "@codemirror/lint";
import { githubDark, githubLight } from "@uiw/codemirror-theme-github";
import { useHookedTheme } from "@/lib/hooks/theme";
import {
  formatJsonTemplate,
  validateJsonTemplate,
} from "@/lib/flow/jsonTemplate";
import FlowPlaceholderExplorer from "./FlowPlaceholderExplorer";

const jsonTemplateLinter = linter((view) => {
  const error = validateJsonTemplate(view.state.doc.toString());
  if (!error) return [];

  const diagnostic: Diagnostic = {
    from: error.from,
    to: error.to,
    severity: "error",
    message: error.message,
  };
  return [diagnostic];
});

function isEmptyDocument(text: string) {
  const trimmed = text.trim();
  return trimmed === "" || trimmed === "{}" || trimmed === "[]";
}

// Pasting a whole JSON document into an empty editor formats it, so bodies
// copied from API docs or minified requests are readable right away.
const formatOnPaste = EditorView.domEventHandlers({
  paste(event, view) {
    const text = event.clipboardData?.getData("text/plain");
    if (!text || !isEmptyDocument(view.state.doc.toString())) return false;

    const formatted = formatJsonTemplate(text.trim());
    if (formatted === null) return false;

    event.preventDefault();
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: formatted },
      selection: { anchor: formatted.length },
    });
    return true;
  },
});

// Shift+Alt+F tidies up the JSON, like in VS Code.
const formatKeymap = keymap.of([
  {
    key: "Shift-Alt-f",
    preventDefault: true,
    run(view) {
      const formatted = formatJsonTemplate(view.state.doc.toString());
      if (formatted === null) return false;

      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: formatted },
      });
      return true;
    },
  },
]);

export default function HttpJsonBodyEditor({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  const { theme } = useHookedTheme();
  const editorRef = useRef<ReactCodeMirrorRef>(null);

  const extensions = useMemo(
    () => [
      json(),
      lintGutter(),
      jsonTemplateLinter,
      formatOnPaste,
      formatKeymap,
      EditorView.lineWrapping,
    ],
    []
  );

  const insertPlaceholder = useCallback(
    (placeholder: string) => {
      const view = editorRef.current?.view;
      const text = `{{${placeholder}}}`;
      if (!view) {
        onChange(value + text);
        return;
      }

      view.dispatch(view.state.replaceSelection(text));
      view.focus();
    },
    [value, onChange]
  );

  const error = validateJsonTemplate(value);

  return (
    <div className="space-y-2">
      <div className="relative">
        <ReactCodeMirror
          ref={editorRef}
          className="rounded-md overflow-hidden border border-input text-sm"
          value={value}
          height="auto"
          minHeight="140px"
          maxHeight="400px"
          width="100%"
          placeholder={'{\n  "content": "Hello {{user.name}}"\n}'}
          basicSetup={{
            lineNumbers: true,
            foldGutter: false,
            indentOnInput: true,
            bracketMatching: true,
            closeBrackets: true,
          }}
          extensions={extensions}
          theme={theme === "light" ? githubLight : githubDark}
          onChange={(v) => onChange(v)}
        />
        <FlowPlaceholderExplorer onSelect={insertPlaceholder} />
      </div>
      {error && <div className="text-xs text-destructive">{error.message}</div>}
    </div>
  );
}
