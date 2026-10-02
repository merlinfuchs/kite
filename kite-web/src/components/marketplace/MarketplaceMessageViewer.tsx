import { useMemo, useState } from "react";
import { Button } from "../ui/button";
import MessagePreview from "../message/MessagePreview";
import { messageSchema } from "@/lib/message/schema";
import { parseMessageData } from "@/lib/message/schemaRestore";
import { MarketplaceListingItem } from "@/lib/types/wire.gen";
import MarketplaceFlowViewer from "./MarketplaceFlowViewer";

interface FlowComponent {
  flowSourceId: string;
  label: string;
  selectMenu: boolean;
}

// Buttons and select menus with a flow, in the order they appear.
function flowComponents(data: unknown): FlowComponent[] {
  const res: FlowComponent[] = [];
  const walk = (components: any[] | undefined) => {
    for (const c of components ?? []) {
      if (c?.flow_source_id) {
        const selectMenu = c.type !== 2;
        res.push({
          flowSourceId: c.flow_source_id,
          label:
            c.label || c.placeholder || (selectMenu ? "Select menu" : "Button"),
          selectMenu,
        });
      }
      for (const option of c?.options ?? []) {
        if (option?.flow_source_id) {
          res.push({
            flowSourceId: option.flow_source_id,
            label: option.label || "Option",
            selectMenu: true,
          });
        }
      }
      walk(c?.components);
      if (c?.accessory) walk([c.accessory]);
    }
  };
  walk((data as any)?.components);
  return res;
}

export default function MarketplaceMessageViewer({
  item,
}: {
  item: MarketplaceListingItem;
}) {
  const [flowSourceId, setFlowSourceId] = useState<string | null>(null);

  const msg = useMemo(() => {
    try {
      // The restore schema upgrades older formats, the preview takes the
      // current one.
      const res = messageSchema.safeParse(parseMessageData(item.message_data));
      return res.success ? res.data : null;
    } catch {
      return null;
    }
  }, [item.message_data]);

  const components = useMemo(
    () =>
      flowComponents(item.message_data).filter(
        (c) => item.message_flow_sources?.[c.flowSourceId]
      ),
    [item.message_data, item.message_flow_sources]
  );

  const selected = components.find((c) => c.flowSourceId === flowSourceId);
  const selectedFlow = flowSourceId
    ? item.message_flow_sources?.[flowSourceId]
    : undefined;

  return (
    <div className="space-y-4">
      <div className="rounded-md border overflow-hidden">
        {msg ? (
          <MessagePreview msg={msg} reducePadding />
        ) : (
          <div className="p-3 text-sm text-muted-foreground">
            This message can&apos;t be previewed.
          </div>
        )}
      </div>

      {components.length > 0 && (
        <div className="space-y-3">
          <div className="text-sm text-muted-foreground">
            Components with a flow, pick one to see what it does.
          </div>
          <div className="flex flex-wrap gap-2">
            {components.map((c) => (
              <Button
                key={c.flowSourceId}
                size="sm"
                variant={
                  flowSourceId === c.flowSourceId ? "default" : "outline"
                }
                onClick={() =>
                  setFlowSourceId(
                    flowSourceId === c.flowSourceId ? null : c.flowSourceId
                  )
                }
              >
                {c.label}
              </Button>
            ))}
          </div>
          {selected && selectedFlow && (
            <MarketplaceFlowViewer
              key={selected.flowSourceId}
              flow={selectedFlow}
              context={
                selected.selectMenu
                  ? "component_select_menu"
                  : "component_button"
              }
            />
          )}
        </div>
      )}
    </div>
  );
}
