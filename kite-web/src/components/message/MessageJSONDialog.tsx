import { ReactNode, useCallback, useEffect, useMemo, useState } from "react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import ReactCodeMirror from "@uiw/react-codemirror";
import { json, jsonParseLinter } from "@codemirror/lang-json";
import { githubDark, githubLight } from "@uiw/codemirror-theme-github";
import { linter, lintGutter } from "@codemirror/lint";
import { Button } from "../ui/button";
import { useCurrentMessage } from "@/lib/message/state";
import { parseMessageData } from "@/lib/message/schemaRestore";
import { toast } from "sonner";
import { useHookedTheme } from "@/lib/hooks/theme";

export default function MessageJSONDialog({
  children,
}: {
  children: ReactNode;
}) {
  const { theme } = useHookedTheme();

  const [open, setOpen] = useState(false);
  const [raw, setRaw] = useState("{}");

  const msg = useCurrentMessage((s) => s);

  useEffect(() => {
    setRaw(JSON.stringify(msg, null, 2));
  }, [msg]);

  const save = useCallback(() => {
    try {
      const data = parseMessageData(JSON.parse(raw));

      msg.replace(data);
      setOpen(false);
    } catch (e) {
      toast.error(`Failed to parse message data: ${e}`);
    }
  }, [msg, raw]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="lg:min-w-[750px]">
        <DialogHeader>
          <DialogTitle>Edit Message JSON</DialogTitle>
          <DialogDescription>
            View and edit the raw JSON of the message. This is for advanced
            users only.
          </DialogDescription>
        </DialogHeader>
        <ReactCodeMirror
          className="flex-1 rounded overflow-hidden max-h-[400px]"
          height="100%"
          width="100%"
          value={raw}
          basicSetup={{
            lineNumbers: false,
            foldGutter: false,
            indentOnInput: true,
          }}
          extensions={[lintGutter(), json(), linter(jsonParseLinter())]}
          theme={theme === "light" ? githubLight : githubDark}
          onChange={(v) => setRaw(v)}
        />
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button onClick={save}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
