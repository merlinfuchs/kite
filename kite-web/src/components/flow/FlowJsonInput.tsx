import { json, jsonParseLinter } from "@codemirror/lang-json";
import { linter, lintGutter } from "@codemirror/lint";
import { githubDark, githubLight } from "@uiw/codemirror-theme-github";
import ReactCodeMirror, { ReactCodeMirrorRef } from "@uiw/react-codemirror";
import { CircleAlertIcon } from "lucide-react";
import { useRef, useState } from "react";
import { useHookedTheme } from "@/lib/hooks/theme";
import FlowPlaceholderExplorer from "./FlowPlaceholderExplorer";

// A JSON object or list whose string values can contain placeholders. The
// value is only updated while the text is valid.
export default function FlowJsonInput({
  value,
  onChange,
}: {
  value: unknown;
  onChange: (value: Record<string, unknown> | unknown[]) => void;
}) {
  const { theme } = useHookedTheme();
  const ref = useRef<ReactCodeMirrorRef>(null);

  const [raw, setRaw] = useState(() => JSON.stringify(value ?? {}, null, 2));
  const [invalid, setInvalid] = useState(false);

  function update(text: string) {
    setRaw(text);
    try {
      const parsed = JSON.parse(text);
      if (typeof parsed !== "object" || parsed === null) {
        throw new Error("not an object or list");
      }
      onChange(parsed);
      setInvalid(false);
    } catch {
      setInvalid(true);
    }
  }

  function insertPlaceholder(placeholder: string) {
    const view = ref.current?.view;
    if (!view) return;

    const { from, to } = view.state.selection.main;
    // Placeholders only work in strings, so outside of one a string is added.
    const quotes = view.state.doc.sliceString(0, from).match(/(?<!\\)"/g);
    const inString = (quotes?.length ?? 0) % 2 === 1;
    const insert = inString ? `{{${placeholder}}}` : `"{{${placeholder}}}"`;

    view.dispatch({
      changes: { from, to, insert },
      selection: { anchor: from + insert.length },
    });
    view.focus();
  }

  return (
    <div>
      <div className="relative">
        <ReactCodeMirror
          ref={ref}
          className="rounded overflow-hidden"
          maxHeight="400px"
          value={raw}
          basicSetup={{
            lineNumbers: false,
            foldGutter: false,
            indentOnInput: true,
          }}
          extensions={[lintGutter(), json(), linter(jsonParseLinter())]}
          theme={theme === "light" ? githubLight : githubDark}
          onChange={update}
        />
        <FlowPlaceholderExplorer onSelect={insertPlaceholder} />
      </div>
      {invalid && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>Invalid JSON. Changes are saved once it&apos;s fixed.</div>
        </div>
      )}
    </div>
  );
}
