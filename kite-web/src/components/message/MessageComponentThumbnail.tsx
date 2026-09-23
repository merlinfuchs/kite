import { NodeId } from "@/lib/message/document";
import MessageComponentCard from "./MessageComponentCard";
import MessageComponentMediaFields from "./MessageComponentMediaFields";

/** Only ever a section accessory, which holds a single node. */
export default function MessageComponentThumbnail({ id }: { id: NodeId }) {
  return (
    <MessageComponentCard id={id} label="Thumbnail" numbered={false}>
      <MessageComponentMediaFields id={id} />
    </MessageComponentCard>
  );
}
