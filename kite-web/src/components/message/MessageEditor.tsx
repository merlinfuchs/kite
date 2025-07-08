import MessageAttachmentSection from "./MessageAttachmentSection";
import MessageEmbedSection from "./MessageEmbedSection";
import MessageBody from "./MessageBody";
import MessageControls from "./MessageControls";
import MessageValidator from "./MessageValidator";
import MessageComponentsSection from "./MessageComponentsSection";

export default function MessageEditor({
  disableComponents,
}: {
  disableComponents?: boolean;
}) {
  return (
    <div className="space-y-8">
      <MessageControls />
      <MessageBody />

      <MessageAttachmentSection />
      <MessageEmbedSection />
      {!disableComponents && <MessageComponentsSection />}

      <MessageValidator />
    </div>
  );
}
