import CollapsibleSection from "./MessageCollapsibleSection";

export default function MessageAttachmentsSection() {
  return (
    <CollapsibleSection title="Attachments" defaultOpen={false}>
      <div className="space-y-4"></div>
    </CollapsibleSection>
  );
}
