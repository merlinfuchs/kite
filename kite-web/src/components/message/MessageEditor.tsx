import MessageAttachmentSection from "./MessageAttachmentSection";
import MessageEmbedSection from "./MessageEmbedSection";
import MessageBody from "./MessageBody";
import MessageControls from "./MessageControls";
import MessageValidator from "./MessageValidator";
import MessageComponentsSection from "./MessageComponentsSection";

export default function MessageEditor({
  disableFlowEditor,
}: {
  disableFlowEditor?: boolean;
}) {
  return (
    <div className="space-y-8">
      <MessageControls />
      <MessageBody />

      <MessageAttachmentSection />
      <MessageEmbedSection />
      <MessageComponentsSection disableFlowEditor={disableFlowEditor} />

      <MessageValidator />
    </div>
  );
}
