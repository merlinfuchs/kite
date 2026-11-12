import { javascript, javascriptLanguage } from "@codemirror/lang-javascript";
import { githubDark, githubLight } from "@uiw/codemirror-theme-github";
import ReactCodeMirror, { EditorView } from "@uiw/react-codemirror";
import { CircleAlertIcon, Maximize2Icon } from "lucide-react";
import { useTheme } from "next-themes";
import { useContext, useMemo, useRef, useState } from "react";
import { CodeNames, codeCompletions } from "@/lib/flow/codeCompletions";
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

const basicSetup = { foldGutter: false, indentOnInput: true };

// JavaScript code, with highlighting and line numbers, which errors refer to.
// The expand button opens the same code in a wide dialog, which edits it
// directly.
export default function FlowCodeInput({
  title,
  description,
  value,
  onChange,
  placeholder,
  error,
  hint,
  names,
}: {
  title: string;
  description?: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  error?: string;
  // Shown after the description in the dialog.
  hint?: string;
  // Inputs and stored variables of the block, which the editor completes.
  names: CodeNames;
}) {
  const { resolvedTheme } = useTheme();
  const readOnly = useContext(FlowReadOnlyContext);
  const [expanded, setExpanded] = useState(false);
  const theme = resolvedTheme === "dark" ? githubDark : githubLight;

  // Read when completing, so the extensions aren't rebuilt on every change.
  const namesRef = useRef(names);
  namesRef.current = names;
  const [panelExtensions, dialogExtensions] = useMemo(() => {
    const completions = codeCompletions(() => namesRef.current).map(
      (autocomplete) => javascriptLanguage.data.of({ autocomplete })
    );
    // The description of a completion goes next to the list, which only
    // fits in the dialog when it's narrow. The dialog's transform makes it
    // cut off what's outside it.
    const narrowInfo = EditorView.theme({
      ".cm-tooltip.cm-completionInfo": {
        maxWidth: "16rem !important",
        overflowWrap: "anywhere",
      },
    });
    return [
      // Long lines wrap so the editor fits the settings panel.
      [javascript(), ...completions, narrowInfo, EditorView.lineWrapping],
      [javascript(), ...completions, narrowInfo],
    ];
  }, []);

  const editorProps = {
    className:
      "rounded border text-sm min-w-0 [&_.cm-editor]:rounded [&_.cm-gutters]:rounded-l",
    value,
    placeholder,
    basicSetup,
    theme,
    onChange,
    readOnly,
  };

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
        {...editorProps}
        minHeight="120px"
        maxHeight="400px"
        extensions={panelExtensions}
      />
      {errorText}

      <Dialog open={expanded} onOpenChange={setExpanded}>
        <DialogContent className="sm:max-w-[min(1000px,calc(100vw-2rem))]">
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            {(description || hint) && (
              <DialogDescription>
                {[description, hint].filter(Boolean).join(" ")}
              </DialogDescription>
            )}
          </DialogHeader>
          <ReactCodeMirror
            {...editorProps}
            height="min(600px, 65vh)"
            extensions={dialogExtensions}
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
