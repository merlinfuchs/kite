import { useEventListenerWebhookSecretRegenerateMutation } from "@/lib/api/mutations";
import { getWebhookUrl } from "@/lib/flow/webhook";
import { useEventListener } from "@/lib/hooks/api";
import { useAppId, useEventId } from "@/lib/hooks/params";
import { CopyIcon, RefreshCwIcon } from "lucide-react";
import { useCallback } from "react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import ConfirmDialog from "./ConfirmDialog";

// Shows the webhook URL of the event listener that is being edited. The URL
// isn't part of the flow, so exports and share codes don't contain it.
export default function WebhookUrlInput() {
  const listener = useEventListener();
  const regenerateMutation = useEventListenerWebhookSecretRegenerateMutation(
    useAppId(),
    useEventId()
  );

  const url = listener ? getWebhookUrl(listener) : null;

  const copy = useCallback(() => {
    if (!url) return;

    navigator.clipboard
      .writeText(url)
      .then(() => toast.success("Copied to clipboard!"))
      .catch(() => toast.error("Failed to copy to clipboard"));
  }, [url]);

  const regenerate = useCallback(() => {
    regenerateMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success(
            "Webhook URL regenerated! It may take a few seconds until the new URL works."
          );
        } else {
          toast.error(
            `Failed to regenerate webhook URL: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }, [regenerateMutation]);

  return (
    <div>
      <div className="font-medium text-foreground">Webhook URL</div>
      <div className="text-muted-foreground text-sm mt-1">
        Send a POST request to this URL to run the flow. Anyone with the URL can
        run it, so keep it secret.
      </div>
      <div className="mt-2 flex items-center space-x-2">
        <Input
          type="text"
          readOnly
          value={url ?? ""}
          onFocus={(e) => e.target.select()}
        />
        <Button
          size="icon"
          variant="outline"
          className="flex-none"
          onClick={copy}
          disabled={!url}
          aria-label="Copy webhook URL"
        >
          <CopyIcon className="h-4 w-4" />
        </Button>
        <ConfirmDialog
          title="Are you sure that you want to regenerate the webhook URL?"
          description="The current URL stops working, and everything that sends requests to it has to be updated."
          onConfirm={regenerate}
        >
          <Button
            size="icon"
            variant="outline"
            className="flex-none"
            disabled={!url || regenerateMutation.isPending}
            aria-label="Regenerate webhook URL"
          >
            <RefreshCwIcon className="h-4 w-4" />
          </Button>
        </ConfirmDialog>
      </div>
    </div>
  );
}
