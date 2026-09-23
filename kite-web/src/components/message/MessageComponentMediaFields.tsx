import {
  MediaGalleryItemNode,
  NodeId,
  ThumbnailNode,
} from "@/lib/message/document";
import { useDocumentStoreApi, useNode } from "@/lib/message/state";
import { nodeField } from "@/lib/message/validationStore";
import MessageInput from "./MessageInput";

/** The URL, description and spoiler inputs shared by thumbnails and gallery items. */
export default function MessageComponentMediaFields({ id }: { id: NodeId }) {
  const data = useNode<ThumbnailNode | MediaGalleryItemNode>(id);
  const { update } = useDocumentStoreApi().getState();

  if (!data) return null;

  return (
    <>
      <div className="flex space-x-3">
        <MessageInput
          type="url"
          label="Image URL"
          value={data.media.url}
          onChange={(url) => update<ThumbnailNode>(id, { media: { url } })}
          validation={nodeField<ThumbnailNode>(id, "media.url")}
          imageUpload
        />
        <div className="flex-none">
          <MessageInput
            type="toggle"
            label="Spoiler"
            value={data.spoiler ?? false}
            onChange={(v) =>
              update<ThumbnailNode>(id, { spoiler: v || undefined })
            }
            validation={nodeField<ThumbnailNode>(id, "spoiler")}
          />
        </div>
      </div>
      <MessageInput
        type="text"
        label="Description"
        maxLength={1024}
        value={data.description ?? ""}
        onChange={(v) =>
          update<ThumbnailNode>(id, { description: v || undefined })
        }
        validation={nodeField<ThumbnailNode>(id, "description")}
        placeholders
      />
    </>
  );
}
