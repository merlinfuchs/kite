import MessageAttachmentSection from "./MessageAttachmentSection";
import MessageEmbedSection from "./MessageEmbedSection";
import MessageBody from "./MessageBody";
import MessageControls from "./MessageControls";
import MessageValidator from "./MessageValidator";
import MessageComponentsSection from "./MessageComponentsSection";
import MessageComponentsV2Toggle from "./MessageComponentsV2Toggle";
import { useComponentsV2Enabled } from "@/lib/message/state";

export default function MessageEditor({
  disableFlowEditor,
  disableAttachments,
}: {
  disableFlowEditor?: boolean;
  disableAttachments?: boolean;
}) {
  // Components v2 messages can't have content or embeds, the components replace both.
  const componentsV2 = useComponentsV2Enabled();

  return (
    <div className="space-y-8">
      <MessageControls />
      <MessageComponentsV2Toggle />
      {!componentsV2 && <MessageBody />}

      {!disableAttachments && <MessageAttachmentSection />}
      {!componentsV2 && <MessageEmbedSection />}
      <MessageComponentsSection disableFlowEditor={disableFlowEditor} />

      <MessageValidator />
    </div>
  );
}
