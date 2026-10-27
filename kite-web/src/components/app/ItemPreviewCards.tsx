import { componentLabel } from "@/lib/flow/templates";
import { FlowContextType, FlowPreviewItemsContext } from "@/lib/flow/context";
import { FlowData } from "@/lib/types/flow.gen";
import { ChevronDownIcon, MailIcon } from "lucide-react";
import { ReactNode, useMemo, useState } from "react";
import { Card } from "../ui/card";
import { Checkbox } from "../ui/checkbox";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "../ui/collapsible";
import { cn } from "@/lib/utils";
import { parsePreviewMessage } from "@/lib/message/schema";
import dynamic from "next/dynamic";
import { isNoteNode } from "@/lib/flow/notes";

// The previews pull in the flow editor and the message renderer. They only
// show when an item is opened, so they're loaded then.
const TemplateFlowPreview = dynamic(() => import("./TemplateFlowPreview"), {
  ssr: false,
});
const MessagePreviewBox = dynamic(
  () => import("../message/MessagePreviewBox"),
  { ssr: false }
);

// ItemPreviewCards show the commands, event listeners and message templates
// of a template or an import, and open to show their blocks.

export type ItemCheckbox = {
  checked: boolean;
  disabled?: boolean;
  onCheckedChange?: (checked: boolean) => void;
};

function ItemCard({
  icon: Icon,
  name,
  description,
  meta,
  checkbox,
  defaultOpen,
  children,
}: {
  icon: React.ComponentType<{ className?: string }>;
  name: string;
  description: string;
  meta: string;
  checkbox?: ItemCheckbox;
  defaultOpen?: boolean;
  children: React.ReactNode;
}) {
  const [open, setOpen] = useState(!!defaultOpen);

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <Card className="overflow-hidden">
        <div className="flex items-center hover:bg-muted/40 transition-colors">
          {/* Outside the trigger, so checking it doesn't open the item. */}
          {checkbox && (
            <Checkbox
              className="ml-4 flex-none"
              checked={checkbox.checked}
              disabled={checkbox.disabled}
              onCheckedChange={(checked) =>
                checkbox.onCheckedChange?.(checked === true)
              }
              aria-label={`Include ${name}`}
            />
          )}
          <CollapsibleTrigger className="min-w-0 flex-1 flex items-center gap-3 p-4 text-left">
            <Icon className="h-5 w-5 text-muted-foreground flex-none" />
            <div className="min-w-0 flex-1">
              <div className="font-medium truncate">{name}</div>
              {description && (
                <div className="text-sm text-muted-foreground">
                  {description}
                </div>
              )}
            </div>
            <span className="text-xs text-muted-foreground flex-none hidden sm:block">
              {meta}
            </span>
            <ChevronDownIcon
              className={cn(
                "h-4 w-4 text-muted-foreground flex-none transition-transform",
                open && "rotate-180"
              )}
            />
          </CollapsibleTrigger>
        </div>
        <CollapsibleContent className="px-4 pb-4">
          {children}
        </CollapsibleContent>
      </Card>
    </Collapsible>
  );
}

export function FlowItemCard({
  icon,
  name,
  description,
  context,
  flow,
  checkbox,
  defaultOpen,
}: {
  icon: React.ComponentType<{ className?: string }>;
  name: string;
  description: string;
  context: FlowContextType;
  flow: FlowData;
  checkbox?: ItemCheckbox;
  defaultOpen?: boolean;
}) {
  const { nodes, edges } = flow;

  return (
    <ItemCard
      icon={icon}
      name={name}
      description={description}
      meta={`${nodes.filter((n) => !isNoteNode(n)).length} blocks`}
      checkbox={checkbox}
      defaultOpen={defaultOpen}
    >
      <TemplateFlowPreview
        name={name}
        nodes={nodes}
        edges={edges}
        context={context}
      />
    </ItemCard>
  );
}

export function MessageItemCard({
  name,
  description,
  data,
  flowSources,
  checkbox,
  defaultOpen,
}: {
  name: string;
  description: string | null;
  data: unknown;
  flowSources: Record<string, FlowData>;
  checkbox?: ItemCheckbox;
  defaultOpen?: boolean;
}) {
  const message = useMemo(() => parsePreviewMessage(data), [data]);
  const flows = useMemo(() => Object.entries(flowSources), [flowSources]);

  return (
    <ItemCard
      icon={MailIcon}
      name={name}
      description={description ?? "A message template."}
      meta={`${flows.length} ${flows.length === 1 ? "button" : "buttons"}`}
      checkbox={checkbox}
      defaultOpen={defaultOpen}
    >
      <div className="space-y-4">
        {message && <MessagePreviewBox msg={message} />}
        {flows.map(([id, flow]) => (
          <div key={id}>
            <div className="text-sm font-medium mb-2">
              When {componentLabel(message, id) ?? "a button"} is clicked
            </div>
            <TemplateFlowPreview
              name={`${name}: ${
                componentLabel(message, id) ?? "a button"
              } is clicked`}
              nodes={flow.nodes}
              edges={flow.edges}
              context="component_button"
            />
          </div>
        ))}
      </div>
    </ItemCard>
  );
}

// PreviewItemsProvider gives the previews inside it the stored variables and
// message templates their blocks use by key, like the ones of a template.
export function PreviewItemsProvider({
  variables,
  messages,
  children,
}: {
  variables?: { key: string; name: string; scoped: boolean }[];
  messages?: { key: string; name: string; data: unknown }[];
  children: ReactNode;
}) {
  const items = useMemo(
    () => ({
      variables: variables?.map(({ key, name, scoped }) => ({
        id: key,
        name,
        scoped,
      })),
      messages: messages?.map(({ key, name, data }) => ({
        id: key,
        name,
        data,
      })),
    }),
    [variables, messages]
  );

  return (
    <FlowPreviewItemsContext.Provider value={items}>
      {children}
    </FlowPreviewItemsContext.Provider>
  );
}
