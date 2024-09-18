import { useCurrentMessage } from "@/lib/message/state";
import CollapsibleSection from "./MessageCollapsibleSection";
import { useShallow } from "zustand/react/shallow";
import { Button } from "../ui/button";
import MessageAttachment from "./MessageAttachment";
import { ChangeEvent, useCallback, useRef } from "react";
import { useAssetCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";

export default function MessageAttachmentSection() {
  const attachments = useCurrentMessage(
    useShallow((state) => state.attachments.map((e) => e.asset_id))
  );
  const addAttachment = useCurrentMessage((state) => state.addAttachment);
  const clearAttachments = useCurrentMessage((state) => state.clearAttachments);

  const createMutation = useAssetCreateMutation(useAppId());
  const inputRef = useRef<HTMLInputElement>(null);

  const onFileUpload = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return;

      const toastId = toast.loading("Uploading attachment...");

      createMutation.mutateAsync(file, {
        onSuccess: (res) => {
          if (res.success) {
            addAttachment({
              asset_id: res.data.id,
            });
          } else {
            toast.error(
              `Failed to upload asset: ${res.error.message} (${res.error.code})`
            );
          }
        },
        onSettled: () => {
          toast.dismiss(toastId);
          e.target.value = "";
        },
      });
    },
    [inputRef, createMutation, addAttachment]
  );

  return (
    <CollapsibleSection
      title="Attachments"
      defaultOpen={false}
      className="space-y-4"
    >
      <div className="flex flex-wrap gap-4">
        {attachments.map((id, i) => (
          <MessageAttachment key={id} attachmentIndex={i} assetId={id} />
        ))}
      </div>
      <div className="space-x-3">
        <input
          type="file"
          className="hidden"
          ref={inputRef}
          onChange={onFileUpload}
        />

        <Button
          onClick={() => inputRef.current?.click()}
          disabled={!!inputRef.current?.value}
        >
          Add Attachment
        </Button>
        <Button onClick={clearAttachments} variant="destructive">
          Clear Attachments
        </Button>
      </div>
    </CollapsibleSection>
  );
}
