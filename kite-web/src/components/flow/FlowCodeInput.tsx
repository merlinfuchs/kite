import { javascript } from "@codemirror/lang-javascript";
import { githubDark, githubLight } from "@uiw/codemirror-theme-github";
import ReactCodeMirror, { EditorView } from "@uiw/react-codemirror";
import { CircleAlertIcon, Maximize2Icon } from "lucide-react";
import { useTheme } from "next-themes";
import { useContext, useState } from "react";
import { FlowReadOnlyContext } from "@/lib/flow/context";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";

// JavaScript code, with highlighting and line numbers, which errors refer to.
// Long lines wrap so the editor fits the settings panel. The expand button
// opens the same code in a wide dialog, which edits it directly.
export default function FlowCodeInput({
  title,
  description,
  value,
  onChange,
  placeholder,
  error,
  hint,
}: {
  title: string;
  description?: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  error?: string;
  // Shown above the code in the dialog.
  hint?: string;
}) {
  const { resolvedTheme } = useTheme();
  const readOnly = useContext(FlowReadOnlyContext);
  const [expanded, setExpanded] = useState(false);
  const theme = resolvedTheme === "dark" ? githubDark : githubLight;

  const errorText = error && (
    <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
      <CircleAlertIcon className="h-5 w-5 flex-none" />
      <div>{error}</div>
    </div>
  );

  return (
    <div>
      <div className="flex items-start justify-between gap-3 mb-2">
        <div>
          <div className="font-medium text-foreground">{title}</div>
          {description && (
            <div className="text-muted-foreground text-sm mt-1">
              {description}
            </div>
          )}
        </div>
        <Button
          variant="outline"
          size="icon"
          className="h-8 w-8 flex-none"
          title="Open in a larger editor"
          onClick={() => setExpanded(true)}
        >
          <Maximize2Icon className="h-4 w-4" />
        </Button>
      </div>
      <ReactCodeMirror
        className="rounded overflow-hidden border text-sm min-w-0"
        minHeight="120px"
        maxHeight="400px"
        value={value}
        placeholder={placeholder}
        basicSetup={{ foldGutter: false, indentOnInput: true }}
        extensions={[javascript(), EditorView.lineWrapping]}
        theme={theme}
        onChange={onChange}
        readOnly={readOnly}
      />
      {errorText}

      <Dialog open={expanded} onOpenChange={setExpanded}>
        <DialogContent className="sm:max-w-[min(1000px,calc(100vw-2rem))]">
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            {hint && <DialogDescription>{hint}</DialogDescription>}
          </DialogHeader>
          <ReactCodeMirror
            className="rounded overflow-hidden border text-sm min-w-0"
            height="min(600px, 65vh)"
            value={value}
            placeholder={placeholder}
            basicSetup={{ foldGutter: false, indentOnInput: true }}
            extensions={[javascript()]}
            theme={theme}
            onChange={onChange}
            readOnly={readOnly}
            autoFocus={!readOnly}
          />
          {errorText}
          <DialogFooter>
            <Button onClick={() => setExpanded(false)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
