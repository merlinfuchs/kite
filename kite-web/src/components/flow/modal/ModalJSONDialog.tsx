import { json, jsonParseLinter } from "@codemirror/lang-json";
import { linter, lintGutter } from "@codemirror/lint";
import { githubDark, githubLight } from "@uiw/codemirror-theme-github";
import ReactCodeMirror from "@uiw/react-codemirror";
import { ReactNode, useState } from "react";
import { toast } from "sonner";
import { importModalData } from "@/lib/flow/modal";
import { useHookedTheme } from "@/lib/hooks/theme";
import { ModalData } from "@/lib/types/flow.gen";
import { Button } from "../../ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../../ui/dialog";

/** View and edit the raw JSON of a modal, like MessageJSONDialog. */
export default function ModalJSONDialog({
  children,
  modal,
  onChange,
}: {
  children: ReactNode;
  modal: ModalData;
  onChange: (modal: ModalData) => void;
}) {
  const { theme } = useHookedTheme();

  const [open, setOpen] = useState(false);
  const [raw, setRaw] = useState("{}");

  const onOpenChange = (open: boolean) => {
    if (open) setRaw(JSON.stringify(modal, null, 2));
    setOpen(open);
  };

  const save = () => {
    try {
      const { modal, ignored } = importModalData(JSON.parse(raw));

      // The settings are checked like any other edit, and errors show in the
      // editor.
      onChange(modal);
      if (ignored.length) {
        toast.warning(
          `Left out what Kite's modals don't support: ${ignored.join(", ")}`
        );
      }
      setOpen(false);
    } catch (e) {
      toast.error(`Failed to parse modal data: ${e}`);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="lg:min-w-[750px]">
        <DialogHeader>
          <DialogTitle>Edit Modal JSON</DialogTitle>
          <DialogDescription>
            View and edit the raw JSON of the modal. This is for advanced users
            only.
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
          onChange={setRaw}
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
