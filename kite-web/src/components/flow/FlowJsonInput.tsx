import { json, jsonParseLinter } from "@codemirror/lang-json";
import { linter, lintGutter } from "@codemirror/lint";
import { githubDark, githubLight } from "@uiw/codemirror-theme-github";
import ReactCodeMirror, { ReactCodeMirrorRef } from "@uiw/react-codemirror";
import { CircleAlertIcon } from "lucide-react";
import { useRef, useState } from "react";
import { useHookedTheme } from "@/lib/hooks/theme";
import JsonEditor from "../common/JsonEditor";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";
import FlowPlaceholderExplorer from "./FlowPlaceholderExplorer";

type Mode = "visual" | "json";

// A JSON object or list whose string values can contain placeholders. It can
// be edited as a tree or as text, where placeholders can be inserted. The text
// only updates the value while it's valid.
export default function FlowJsonInput({
  value,
  onChange,
}: {
  value: unknown;
  onChange: (value: Record<string, unknown> | unknown[]) => void;
}) {
  const { theme } = useHookedTheme();
  const ref = useRef<ReactCodeMirrorRef>(null);

  const [mode, setMode] = useState<Mode>("visual");
  const [raw, setRaw] = useState("");
  const [invalid, setInvalid] = useState(false);

  function switchMode(newMode: Mode) {
    if (newMode === "json") {
      setRaw(JSON.stringify(value ?? {}, null, 2));
      setInvalid(false);
    }
    setMode(newMode);
  }

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
      <div className="flex items-center justify-between gap-3 mb-2">
        <Tabs value={mode} onValueChange={(v) => switchMode(v as Mode)}>
          <TabsList className="h-8 py-0">
            <TabsTrigger value="visual" className="text-xs" disabled={invalid}>
              Visual
            </TabsTrigger>
            <TabsTrigger value="json" className="text-xs">
              JSON
            </TabsTrigger>
          </TabsList>
        </Tabs>
        {mode === "visual" && (
          <div className="text-muted-foreground text-xs">
            Switch to JSON to insert placeholders.
          </div>
        )}
      </div>
      {mode === "visual" ? (
        <JsonEditor src={value ?? {}} onChange={onChange} />
      ) : (
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
      )}
      {invalid && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>Invalid JSON. Changes are saved once it&apos;s fixed.</div>
        </div>
      )}
    </div>
  );
}
