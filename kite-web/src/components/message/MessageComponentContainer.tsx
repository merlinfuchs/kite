import { useMemo } from "react";
import { ContainerNode, NodeId, slotLimit } from "@/lib/message/document";
import {
  useChildIds,
  useDocumentStoreApi,
  useNode,
  useNodeActions,
} from "@/lib/message/state";
import { nodeField, nodeScope, slotScope } from "@/lib/message/validationStore";
import { colorIntToHex } from "@/tools/common/utils/color";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageComponentAddDropdown from "./MessageComponentAddDropdown";
import MessageComponentEntry from "./MessageComponentEntry";
import MessageInput from "./MessageInput";
import MessageNodeActions from "./MessageNodeActions";

export default function MessageComponentContainer({
  id,
  disableFlowEditor,
}: {
  id: NodeId;
  disableFlowEditor?: boolean;
}) {
  const data = useNode<ContainerNode>(id);
  const childIds = useChildIds(id, "components");
  const actions = useNodeActions(id);
  const { index } = actions;
  const { update, removeChildren } = useDocumentStoreApi().getState();

  const colorHex = useMemo(
    () =>
      data?.accent_color !== undefined
        ? colorIntToHex(data.accent_color)
        : "#1f2225",
    [data?.accent_color]
  );

  if (!data) return null;

  return (
    <Card
      className="px-4 py-3 border-l-4 rounded-l-sm"
      style={{ borderLeftColor: colorHex }}
    >
      <MessageCollapsibleSection
        title={`Container ${index + 1}`}
        size="lg"
        validation={nodeScope(id)}
        className="space-y-4"
        actions={<MessageNodeActions actions={actions} size="lg" />}
      >
        <div className="flex space-x-3">
          <MessageInput
            type="color"
            label="Accent Color"
            value={data.accent_color}
            onChange={(accent_color) =>
              update<ContainerNode>(id, { accent_color })
            }
            validation={nodeField<ContainerNode>(id, "accent_color")}
          />
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Spoiler"
              value={data.spoiler ?? false}
              onChange={(v) =>
                update<ContainerNode>(id, { spoiler: v || undefined })
              }
              validation={nodeField<ContainerNode>(id, "spoiler")}
            />
          </div>
        </div>

        <MessageCollapsibleSection
          title="Components"
          size="md"
          validation={slotScope(id, "components")}
          className="space-y-3"
        >
          {childIds.map((childId) => (
            <MessageComponentEntry
              key={childId}
              id={childId}
              disableFlowEditor={disableFlowEditor}
            />
          ))}
          <div className="flex space-x-3">
            <MessageComponentAddDropdown
              parentId={id}
              context="container"
              size="sm"
              disabled={childIds.length >= slotLimit("container", "components")}
            />
            <Button
              size="sm"
              variant="destructive"
              onClick={() => removeChildren(id, "components")}
            >
              Clear Components
            </Button>
          </div>
        </MessageCollapsibleSection>
      </MessageCollapsibleSection>
    </Card>
  );
}
