import { ReactNode, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { ShareCodeDisplay, ShareCodePanel } from "./ShareCode";
import { useShareCodeCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { BracesIcon, CopyIcon, KeyRoundIcon } from "lucide-react";
import { toast } from "sonner";

type ExportProps = {
  title: string;
  type: "command" | "event_listener";
  shareData: Record<string, unknown>;
};

export default function FlowExportDialog({
  children,
  ...props
}: ExportProps & { children: ReactNode }) {
  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-lg">
        <ExportForm {...props} />
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so its state resets every time it's opened.
function ExportForm({ title, type, shareData }: ExportProps) {
  const [showJson, setShowJson] = useState(false);
  const [code, setCode] = useState<string | null>(null);

  const appId = useAppId();
  const createMutation = useShareCodeCreateMutation(appId);

  function copy(text: string) {
    navigator.clipboard
      .writeText(text)
      .then(() => toast.success("Copied to clipboard!"))
      .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
  }

  function generateCode() {
    createMutation.mutate(
      { type, data: shareData },
      {
        onSuccess: (res) => {
          if (res.success) {
            setCode(res.data.code);
          } else {
            toast.error(
              `Failed to generate share code: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  if (showJson) {
    const json = JSON.stringify(shareData);

    return (
      <>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>
            Copy the JSON below and import it into another app.
          </DialogDescription>
        </DialogHeader>
        <Textarea
          readOnly
          value={json}
          className="h-36 resize-none break-all font-mono text-xs"
          onFocus={(e) => e.target.select()}
        />
        <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
          <Button variant="ghost" onClick={() => setShowJson(false)}>
            <KeyRoundIcon className="mr-2 h-4 w-4" />
            Use share code
          </Button>
          <Button onClick={() => copy(json)}>
            <CopyIcon className="mr-2 h-4 w-4" />
            Copy JSON
          </Button>
        </DialogFooter>
      </>
    );
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>
          Anyone with the code can import this into their app. Codes expire
          after 90 days without use.
        </DialogDescription>
      </DialogHeader>
      <ShareCodePanel>
        <ShareCodeDisplay code={code} />
      </ShareCodePanel>
      <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
        <Button variant="ghost" onClick={() => setShowJson(true)}>
          <BracesIcon className="mr-2 h-4 w-4" />
          Use JSON
        </Button>
        {code ? (
          <Button onClick={() => copy(code)}>
            <CopyIcon className="mr-2 h-4 w-4" />
            Copy code
          </Button>
        ) : (
          <LoadingButton
            onClick={generateCode}
            loading={createMutation.isPending}
          >
            Generate code
          </LoadingButton>
        )}
      </DialogFooter>
    </>
  );
}
