import MessagePreviewBox from "@/components/message/MessagePreviewBox";
import { useHookedTheme } from "@/lib/hooks/theme";
import { parsePreviewMessage } from "@/lib/message/documentConvert";
import { json } from "@codemirror/lang-json";
import { githubDark, githubLight } from "@uiw/codemirror-theme-github";
import ReactCodeMirror, { EditorView } from "@uiw/react-codemirror";
import { useMemo } from "react";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";

// FlowMessageViewDialog shows the message a block sends in previews, where
// the message editor can't open: its own message, or its message template
// with the template's name, and its JSON.
export default function FlowMessageViewDialog({
  name,
  message,
}: {
  name?: string;
  message: unknown;
}) {
  const { theme } = useHookedTheme();
  const msg = useMemo(() => parsePreviewMessage(message), [message]);
  if (!msg) return null;

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full">
          View Message
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90dvh] overflow-y-auto">
        <DialogTitle>{name ?? "Message"}</DialogTitle>
        <DialogDescription>The message this block sends.</DialogDescription>
        <Tabs defaultValue="preview">
          <TabsList>
            <TabsTrigger value="preview">Preview</TabsTrigger>
            <TabsTrigger value="json">JSON</TabsTrigger>
          </TabsList>
          <TabsContent value="preview">
            <MessagePreviewBox msg={msg} />
          </TabsContent>
          <TabsContent value="json">
            <ReactCodeMirror
              className="rounded overflow-hidden"
              value={JSON.stringify(msg, null, 2)}
              basicSetup={{ lineNumbers: false, foldGutter: false }}
              extensions={[json(), EditorView.lineWrapping]}
              theme={theme === "light" ? githubLight : githubDark}
              readOnly
            />
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
