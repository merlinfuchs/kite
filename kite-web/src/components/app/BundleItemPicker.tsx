import { useCommands, useEventListeners, useMessages } from "@/lib/hooks/api";
import { eventListenerLabel } from "@/lib/flow/eventTypes";
import { BundleExportRequest } from "@/lib/types/wire.gen";
import { Checkbox } from "../ui/checkbox";

type ItemSection = {
  title: string;
  field: keyof BundleExportRequest;
  items: { id: string; label: string; description: string }[];
};

// useBundleItemSections are the commands, event listeners and message
// templates of the app that can be shared, or undefined while they load.
export function useBundleItemSections(): ItemSection[] | undefined {
  const commands = useCommands();
  const eventListeners = useEventListeners();
  const messages = useMessages();
  if (!commands || !eventListeners || !messages) return undefined;

  return (
    [
      {
        title: "Commands",
        field: "command_ids",
        items: commands.flatMap((c) =>
          c
            ? [{ id: c.id, label: `/${c.name}`, description: c.description }]
            : []
        ),
      },
      {
        title: "Event listeners",
        field: "event_listener_ids",
        items: eventListeners.flatMap((l) =>
          l
            ? [
                {
                  id: l.id,
                  label: eventListenerLabel(l.source, l.type),
                  description: l.description,
                },
              ]
            : []
        ),
      },
      {
        title: "Message templates",
        field: "message_ids",
        items: messages.flatMap((m) =>
          m
            ? [{ id: m.id, label: m.name, description: m.description ?? "" }]
            : []
        ),
      },
    ] satisfies ItemSection[]
  ).filter((s) => s.items.length);
}

// selectedItems are the selected IDs by the field of the request.
export function selectedItems(
  sections: ItemSection[],
  selected: Set<string>
): BundleExportRequest {
  const req: BundleExportRequest = {
    command_ids: [],
    event_listener_ids: [],
    message_ids: [],
  };
  for (const section of sections) {
    req[section.field] = section.items.flatMap((i) =>
      selected.has(i.id) ? [i.id] : []
    );
  }
  return req;
}

export function BundleItemPicker({
  sections,
  selected,
  onChange,
}: {
  sections: ItemSection[];
  selected: Set<string>;
  onChange: (selected: Set<string>) => void;
}) {
  function toggle(id: string, checked: boolean) {
    const next = new Set(selected);
    if (checked) next.add(id);
    else next.delete(id);
    onChange(next);
  }

  if (!sections.length) {
    return (
      <p className="text-sm text-muted-foreground">
        Your app has no commands, event listeners or message templates yet.
      </p>
    );
  }

  return (
    <div className="space-y-5 min-w-0">
      {sections.map((section) => (
        <div key={section.title}>
          <div className="text-sm font-medium mb-2">{section.title}</div>
          <div className="space-y-2.5">
            {section.items.map((item) => (
              <label
                key={item.id}
                className="flex items-center gap-2.5 text-sm cursor-pointer"
              >
                <Checkbox
                  checked={selected.has(item.id)}
                  onCheckedChange={(checked) => toggle(item.id, !!checked)}
                />
                <span className="truncate">{item.label}</span>
                <span className="text-xs text-muted-foreground truncate ml-auto max-w-[50%]">
                  {item.description}
                </span>
              </label>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
