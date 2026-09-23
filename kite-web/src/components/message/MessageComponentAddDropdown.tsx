import { ChevronDownIcon } from "lucide-react";
import { DocumentStore, NewNode, NodeId } from "@/lib/message/document";
import { useDocumentStoreApi } from "@/lib/message/state";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";

const componentTypes: {
  label: string;
  node: NewNode;
  rootOnly?: boolean;
  /** Fills the new row with a select menu instead of leaving it for buttons. */
  selectMenu?: boolean;
}[] = [
  { label: "Button Row", node: { type: "actionRow" } },
  { label: "Select Menu", node: { type: "actionRow" }, selectMenu: true },
  { label: "Section", node: { type: "section" } },
  { label: "Text Display", node: { type: "textDisplay", content: "" } },
  { label: "Media Gallery", node: { type: "mediaGallery" } },
  { label: "File", node: { type: "file", file: { url: "" } } },
  {
    label: "Separator",
    node: { type: "separator", divider: true, spacing: 1 },
  },
  { label: "Container", node: { type: "container" }, rootOnly: true },
];

/** Puts a select menu with one option into an empty row. */
export function insertSelectMenu(
  insert: DocumentStore["insert"],
  rowId: NodeId
) {
  const menuId = insert(rowId, "components", "end", { type: "selectMenu" });
  insert(menuId, "options", "end", { type: "selectOption", label: "" });
}

export default function MessageComponentAddDropdown({
  parentId,
  context,
  disabled,
  size,
}: {
  parentId: NodeId;
  context: "root" | "container";
  disabled?: boolean;
  size?: "sm";
}) {
  const { insert } = useDocumentStoreApi().getState();

  const add = (node: NewNode, selectMenu?: boolean) => {
    const id = insert(parentId, "components", "end", node);

    if (selectMenu) {
      insertSelectMenu(insert, id);
    }

    // Sections need at least one text and an accessory, so start with both.
    if (node.type === "section") {
      insert(id, "components", "end", { type: "textDisplay", content: "" });
      insert(id, "accessory", "end", {
        type: "thumbnail",
        media: { url: "" },
      });
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild disabled={disabled}>
        <Button size={size} className="space-x-2">
          <div>Add Component</div>
          <ChevronDownIcon className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        {componentTypes
          .filter((c) => !c.rootOnly || context === "root")
          .map((c) => (
            <DropdownMenuItem
              key={c.label}
              onClick={() => add(c.node, c.selectMenu)}
            >
              {c.label}
            </DropdownMenuItem>
          ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
