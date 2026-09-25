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
import { useShareCodeCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
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
      <DialogContent className="max-w-2xl">
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
          className="min-h-[78px] max-h-[218px]"
          onFocus={(e) => e.target.select()}
        />
        <DialogFooter className="sm:justify-between">
          <Button variant="link" onClick={() => setShowJson(false)}>
            Use a share code instead
          </Button>
          <Button onClick={() => copy(json)}>Copy to clipboard</Button>
        </DialogFooter>
      </>
    );
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>
          Generate a share code and import it into another app.
        </DialogDescription>
      </DialogHeader>
      <div className="flex flex-col items-center gap-3 py-6">
        {code ? (
          <>
            <span className="text-3xl font-mono tracking-widest">{code}</span>
            <Button variant="outline" onClick={() => copy(code)}>
              Copy code
            </Button>
          </>
        ) : (
          <LoadingButton
            onClick={generateCode}
            loading={createMutation.isPending}
          >
            Generate share code
          </LoadingButton>
        )}
        <Button variant="link" onClick={() => setShowJson(true)}>
          Use JSON instead
        </Button>
      </div>
    </>
  );
}
