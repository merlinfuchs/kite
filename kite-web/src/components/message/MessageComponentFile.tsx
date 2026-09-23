import { FileNode, MessageNode, NodeId } from "@/lib/message/document";
import { useDocument, useDocumentStoreApi, useNode } from "@/lib/message/state";
import { nodeField } from "@/lib/message/validationStore";
import { useAssetQueries } from "@/lib/api/queries";
import { useAppId } from "@/lib/hooks/params";
import { useShallow } from "zustand/react/shallow";
import MessageComponentCard from "./MessageComponentCard";
import MessageInput from "./MessageInput";

export default function MessageComponentFile({ id }: { id: NodeId }) {
  const data = useNode<FileNode>(id);
  const { update } = useDocumentStoreApi().getState();

  // A file component can only point at a file uploaded with the message, which
  // is sent under the name of its asset.
  const assetIds = useDocument(
    useShallow((state) =>
      (state.nodes[state.rootId] as MessageNode).attachments.map(
        (a) => a.asset_id
      )
    )
  );
  const options = useAssetQueries(useAppId(), assetIds).map((asset) => ({
    label: asset.name,
    value: `attachment://${asset.name}`,
  }));

  if (!data) return null;

  return (
    <MessageComponentCard id={id} label="File">
      <div className="flex space-x-3">
        <MessageInput
          type="select"
          label="Attachment"
          value={data.file.url}
          options={options}
          placeholder={
            options.length ? "Select an attachment" : "Add an attachment first"
          }
          onChange={(url) => update<FileNode>(id, { file: { url } })}
          validation={nodeField<FileNode>(id, "file.url")}
        />
        <div className="flex-none">
          <MessageInput
            type="toggle"
            label="Spoiler"
            value={data.spoiler ?? false}
            onChange={(v) => update<FileNode>(id, { spoiler: v || undefined })}
            validation={nodeField<FileNode>(id, "spoiler")}
          />
        </div>
      </div>
    </MessageComponentCard>
  );
}
