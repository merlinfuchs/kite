import MessageAttachmentSection from "./MessageAttachmentSection";
import MessageEmbedSection from "./MessageEmbedSection";
import MessageBody from "./MessageBody";
import MessageControls from "./MessageControls";
import MessageValidator from "./MessageValidator";

export default function MessageEditor() {
  return (
    <div className="space-y-8">
      <MessageControls />
      <MessageBody />

      <MessageAttachmentSection />
      <MessageEmbedSection />

      <MessageValidator />
    </div>
  );
}
