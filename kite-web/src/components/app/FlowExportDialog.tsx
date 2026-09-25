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
import { apiRequest } from "@/lib/api/client";
import { toast } from "sonner";

async function createShareCode(data: string): Promise<string> {
  const res = await apiRequest<{ code: string }>("/v1/share", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ data }),
  });
  if (!res.success) {
    throw new Error(res.error.message);
  }
  return res.data.code;
}

export default function FlowExportDialog({
  title,
  shareData,
  children,
}: {
  title: string;
  shareData: Record<string, unknown>;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [showJson, setShowJson] = useState(false);
  const [code, setCode] = useState<string | null>(null);
  const [generating, setGenerating] = useState(false);

  const shareCode = open ? JSON.stringify(shareData) : "";

  function copy(text: string) {
    navigator.clipboard
      .writeText(text)
      .then(() => toast.success("Copied to clipboard!"))
      .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
  }

  async function generateCode() {
    setGenerating(true);
    try {
      const generatedCode = await createShareCode(shareCode);
      setCode(generatedCode);
    } catch {
      toast.error("Failed to generate a share code, use the raw JSON instead");
    } finally {
      setGenerating(false);
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        if (!open) {
          setCode(null);
          setShowJson(false);
        }
      }}
    >
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>
            {showJson
              ? "Copy the raw JSON below and import it into another app."
              : "Share the code below, or copy the raw JSON instead."}
          </DialogDescription>
        </DialogHeader>

        {!showJson && (
          <div className="flex flex-col items-center gap-3 py-6">
            {!code ? (
              <LoadingButton onClick={generateCode} loading={generating}>
                Generate Share Code
              </LoadingButton>
            ) : (
              <>
                <span className="text-3xl font-mono tracking-widest">
                  {code}
                </span>
                <Button variant="outline" onClick={() => copy(code)}>
                  Copy Code
                </Button>
              </>
            )}
            <button
              type="button"
              onClick={() => setShowJson(true)}
              className="text-sm text-muted-foreground underline underline-offset-2"
            >
              Use raw JSON instead
            </button>
          </div>
        )}

        {showJson && (
          <>
            <Textarea
              readOnly
              value={shareCode}
              className="min-h-[78px] max-h-[218px]"
              onFocus={(e) => e.target.select()}
            />
            <button
              type="button"
              onClick={() => setShowJson(false)}
              className="text-sm text-muted-foreground underline underline-offset-2 self-start"
            >
              Use share code instead
            </button>
            <DialogFooter>
              <Button onClick={() => copy(shareCode)}>
                Copy to clipboard
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}