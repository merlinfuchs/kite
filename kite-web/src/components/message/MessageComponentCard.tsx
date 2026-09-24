import { ReactNode } from "react";
import { NodeId } from "@/lib/message/document";
import { useNodeActions } from "@/lib/message/state";
import { nodeScope } from "@/lib/message/validationStore";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageNodeActions from "./MessageNodeActions";

/** The card, header and node actions that every leaf component editor shares. */
export default function MessageComponentCard({
  id,
  label,
  numbered = true,
  defaultOpen,
  children,
}: {
  id: NodeId;
  label: string;
  /** Appends the position among its siblings, off for single-slot nodes like accessories. */
  numbered?: boolean;
  defaultOpen?: boolean;
  children: ReactNode;
}) {
  const actions = useNodeActions(id);

  return (
    <Card className="p-3">
      <MessageCollapsibleSection
        title={numbered ? `${label} ${actions.index + 1}` : label}
        size="md"
        validation={nodeScope(id)}
        className="space-y-3"
        defaultOpen={defaultOpen}
        actions={<MessageNodeActions actions={actions} />}
      >
        {children}
      </MessageCollapsibleSection>
    </Card>
  );
}
