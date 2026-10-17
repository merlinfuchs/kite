import {
  useChildIds,
  useComponentsV2Enabled,
  useDocumentStoreApi,
  useRootId,
} from "@/lib/message/state";
import { ChevronDownIcon } from "lucide-react";
import { slotLimit } from "@/lib/message/document";
import { selectMenuTypes } from "@/lib/message/selectMenu";
import { slotScope } from "@/lib/message/validationStore";
import CollapsibleSection from "./MessageCollapsibleSection";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
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
      <div className="flex flex-wrap gap-3">
        {componentsV2 ? (
          <MessageComponentAddDropdown
            parentId={rootId}
            context="root"
            disabled={componentIds.length >= limit}
          />
        ) : (
          <>
            <Button
              onClick={() =>
                insert(rootId, "components", "end", { type: "actionRow" })
              }
              disabled={componentIds.length >= limit}
            >
              Add Button Row
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger
                asChild
                disabled={componentIds.length >= limit}
              >
                <Button className="space-x-2">
                  <div>Add Select Menu</div>
                  <ChevronDownIcon className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start">
                {selectMenuTypes.map((t) => (
                  <DropdownMenuItem
                    key={t.type}
                    onClick={() => {
                      const rowId = insert(rootId, "components", "end", {
                        type: "actionRow",
                      });
                      insert(rowId, "components", "end", {
                        type: "selectMenu",
                        select_type: t.type === 3 ? undefined : t.type,
                      });
                    }}
                  >
                    {t.label}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </>
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
