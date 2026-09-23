import {
  useDocument,
  useDocumentStoreApi,
  useRootId,
} from "@/lib/message/state";
import { MessageNode } from "@/lib/message/document";
import { nodeScope } from "@/lib/message/validationStore";
import CollapsibleSection from "./MessageCollapsibleSection";
import { useShallow } from "zustand/react/shallow";
import { Button } from "../ui/button";
import MessageAttachment from "./MessageAttachment";
import { ChangeEvent, useCallback, useRef } from "react";
import { useAssetCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";

export default function MessageAttachmentSection() {
  const rootId = useRootId();
  const attachments = useDocument(
    useShallow((state) =>
      ((state.nodes[rootId] as MessageNode | undefined)?.attachments ?? []).map(
        (a) => a.asset_id
      )
    )
  );
  const store = useDocumentStoreApi();

  const addAttachment = useCallback(
    (assetId: string) => {
      const root = store.getState().nodes[rootId] as MessageNode;
      store.getState().update<MessageNode>(rootId, {
        attachments: [...root.attachments, { asset_id: assetId }],
      });
    },
    [store, rootId]
  );

  const clearAttachments = useCallback(
    () => store.getState().update<MessageNode>(rootId, { attachments: [] }),
    [store, rootId]
  );

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
            addAttachment(res.data.id);
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
    [createMutation, addAttachment]
  );

  return (
    <CollapsibleSection
      title="Attachments"
      validation={nodeScope<MessageNode>(rootId, ["attachments"])}
      className="space-y-4"
    >
      <div className="flex flex-wrap gap-4">
        {attachments.map((id, i) => (
          <MessageAttachment key={id} attachmentIndex={i} assetId={id} />
        ))}
      </div>
      <div className="space-x-3">
        <Button
          onClick={() => inputRef.current?.click()}
          disabled={!!inputRef.current?.value || attachments.length >= 10}
        >
          Add Attachment
        </Button>
        <Button onClick={clearAttachments} variant="outline">
          Clear Attachments
        </Button>
      </div>

      <input
        type="file"
        className="hidden"
        ref={inputRef}
        onChange={onFileUpload}
      />
    </CollapsibleSection>
  );
}
