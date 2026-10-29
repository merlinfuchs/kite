import { useState } from "react";
import {
  DialogClose,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { ShareCodeInput, ShareCodePanel } from "./ShareCode";
import { useShareCodeResolveMutation } from "@/lib/api/mutations";
import { BracesIcon, KeyRoundIcon } from "lucide-react";
import { toast } from "sonner";

// ShareCodeImportForm asks for a share code or the exported JSON, and passes
// on what it holds. type is the type of the share code, or null for JSON.
export function ShareCodeImportForm({
  title,
  jsonPlaceholder,
  submitLabel,
  loading,
  onData,
}: {
  title: string;
  jsonPlaceholder: string;
  submitLabel: string;
  loading?: boolean;
  onData: (data: unknown, type: string | null) => void;
}) {
  const [useJson, setUseJson] = useState(false);
  const [json, setJson] = useState("");
  const [code, setCode] = useState("");

  const resolveMutation = useShareCodeResolveMutation();
  const busy = loading || resolveMutation.isPending;

  function onSubmit() {
    if (useJson) {
      let parsed;
      try {
        parsed = JSON.parse(json);
      } catch {}
      onData(parsed, null);
      return;
    }

    if (!code) {
      toast.error("Enter a share code");
      return;
    }

    resolveMutation.mutate(code, {
      onSuccess: (res) => {
        if (!res.success) {
          toast.error(
            res.error.code === "unknown_share_code"
              ? "Unknown share code"
              : `Failed to resolve share code: ${res.error.message} (${res.error.code})`
          );
          return;
        }
        onData(res.data.data, res.data.type);
      },
    });
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>
          {useJson
            ? "Paste the JSON that was exported from another app."
            : "Enter the share code that was exported from another app."}
        </DialogDescription>
      </DialogHeader>

      {useJson ? (
        <Textarea
          value={json}
          onChange={(e) => setJson(e.target.value)}
          placeholder={jsonPlaceholder}
          minRows={8}
          maxRows={8}
          className="resize-none break-all font-mono md:text-xs"
          autoFocus
        />
      ) : (
        <ShareCodePanel>
          <ShareCodeInput
            value={code}
            onChange={setCode}
            onSubmit={() => !busy && onSubmit()}
          />
        </ShareCodePanel>
      )}

      <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
        <Button variant="ghost" onClick={() => setUseJson(!useJson)}>
          {useJson ? (
            <KeyRoundIcon className="mr-2 h-4 w-4" />
          ) : (
            <BracesIcon className="mr-2 h-4 w-4" />
          )}
          {useJson ? "Use share code" : "Use JSON"}
        </Button>
        <div className="flex flex-col-reverse gap-2 sm:flex-row">
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <LoadingButton onClick={onSubmit} loading={busy}>
            {submitLabel}
          </LoadingButton>
        </div>
      </DialogFooter>
    </>
  );
}
