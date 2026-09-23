import {
  useChildIds,
  useComponentsV2Enabled,
  useDocumentStoreApi,
  useRootId,
} from "@/lib/message/state";
import { slotLimit } from "@/lib/message/document";
import { slotScope } from "@/lib/message/validationStore";
import CollapsibleSection from "./MessageCollapsibleSection";
import { Button } from "../ui/button";
import MessageComponentEntry from "./MessageComponentEntry";
import MessageComponentAddDropdown from "./MessageComponentAddDropdown";

export default function MessageComponentsSection({
  disableFlowEditor,
}: {
  disableFlowEditor?: boolean;
}) {
  const rootId = useRootId();
  const componentIds = useChildIds(rootId, "components");
  const componentsV2 = useComponentsV2Enabled();
  const { insert, removeChildren } = useDocumentStoreApi().getState();

  const limit = slotLimit("message", "components", componentsV2);

  return (
    <CollapsibleSection
      title="Components"
      validation={slotScope(rootId, "components")}
      className="space-y-4"
    >
      {componentIds.map((id) => (
        <MessageComponentEntry
          key={id}
          id={id}
          disableFlowEditor={disableFlowEditor}
        />
      ))}
      <div className="flex space-x-3">
        {componentsV2 ? (
          <MessageComponentAddDropdown
            parentId={rootId}
            context="root"
            disabled={componentIds.length >= limit}
          />
        ) : (
          <Button
            onClick={() =>
              insert(rootId, "components", "end", { type: "actionRow" })
            }
            disabled={componentIds.length >= limit}
          >
            Add Button Row
          </Button>
        )}
        <Button
          onClick={() => removeChildren(rootId, "components")}
          variant="outline"
        >
          Clear Components
        </Button>
      </div>
    </CollapsibleSection>
  );
}
