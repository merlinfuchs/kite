import { ChevronDownIcon } from "lucide-react";
import { MessageNode, NewNode, NodeId } from "@/lib/message/document";
import { useDocument, useDocumentStoreApi } from "@/lib/message/state";
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
  needsAttachment?: boolean;
  /** A component to put inside the new one, like the select menu of its row. */
  child?: NewNode;
}[] = [
  { label: "Button Row", node: { type: "actionRow" } },
  {
    label: "Select Menu",
    node: { type: "actionRow" },
    child: { type: "selectMenu" },
  },
  { label: "Section", node: { type: "section" } },
  { label: "Text Display", node: { type: "textDisplay", content: "" } },
  { label: "Media Gallery", node: { type: "mediaGallery" } },
  {
    label: "File",
    node: { type: "file", file: { url: "" } },
    needsAttachment: true,
  },
  {
    label: "Separator",
    node: { type: "separator", divider: true, spacing: 1 },
  },
  { label: "Container", node: { type: "container" }, rootOnly: true },
];

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
  // A file component can only show a file attached to the message.
  const hasAttachments = useDocument(
    (state) => (state.nodes[state.rootId] as MessageNode).attachments.length > 0
  );

  const add = (node: NewNode, child?: NewNode) => {
    const id = insert(parentId, "components", "end", node);
    if (child) insert(id, "components", "end", child);
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
          .filter((c) => !c.needsAttachment || hasAttachments)
          .map((c) => (
            <DropdownMenuItem
              key={c.label}
              onClick={() => add(c.node, c.child)}
            >
              {c.label}
            </DropdownMenuItem>
          ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
