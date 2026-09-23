import { FileNode, MessageNode, NodeId } from "@/lib/message/document";
import {
  useDocument,
  useDocumentStoreApi,
  useNode,
  useNodeActions,
  useNodeIndex,
} from "@/lib/message/state";
import { nodeField, nodeScope } from "@/lib/message/validationStore";
import { useAssetQueries } from "@/lib/api/queries";
import { useAppId } from "@/lib/hooks/params";
import { useShallow } from "zustand/react/shallow";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageInput from "./MessageInput";
import MessageNodeActions from "./MessageNodeActions";

export default function MessageComponentFile({ id }: { id: NodeId }) {
  const data = useNode<FileNode>(id);
  const { index } = useNodeIndex(id);
  const actions = useNodeActions(id);
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
  const assets = useAssetQueries(useAppId(), assetIds);
  const options = assets
    .map((q) => (q.data?.success ? q.data.data.name : undefined))
    .filter((name): name is string => !!name)
    .map((name) => ({ label: name, value: `attachment://${name}` }));

  if (!data) return null;

  return (
    <Card className="p-3">
      <MessageCollapsibleSection
        title={`File ${index + 1}`}
        size="md"
        validation={nodeScope(id)}
        actions={<MessageNodeActions actions={actions} />}
      >
        <div className="flex space-x-3">
          <MessageInput
            type="select"
            label="Attachment"
            value={data.file.url}
            options={options}
            placeholder={
              options.length
                ? "Select an attachment"
                : "Add an attachment first"
            }
            onChange={(url) => update<FileNode>(id, { file: { url } })}
            validation={nodeField<FileNode>(id, "file.url")}
          />
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Spoiler"
              value={data.spoiler ?? false}
              onChange={(v) =>
                update<FileNode>(id, { spoiler: v || undefined })
              }
              validation={nodeField<FileNode>(id, "spoiler")}
            />
          </div>
        </div>
      </MessageCollapsibleSection>
    </Card>
  );
}
