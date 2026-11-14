import { ReactNode, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import {
  ArrowLeftIcon,
  BookmarkIcon,
  CheckIcon,
  CircleAlertIcon,
  CircleXIcon,
  ExternalLinkIcon,
  SlashSquareIcon,
  UploadIcon,
} from "lucide-react";
import { toast } from "sonner";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { FlowItemCard } from "./ItemPreviewCards";
import { usePlanLimit } from "./PlanLimit";
import {
  useCommandsImportMutation,
  useEventListenersImportMutation,
} from "@/lib/api/mutations";
import { useCommands, useEventListeners } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { track } from "@/lib/analytics";
import { eventListenerLabel, eventSourceName } from "@/lib/flow/eventTypes";
import { ConvertedItem, convertBotGhostItems } from "@/lib/botghost/convert";
import { parseBotGhostExport } from "@/lib/botghost/parse";
import { convertInventorItems } from "@/lib/inventor/convert";
import { parseInventorExport } from "@/lib/inventor/parse";
import { helperBookmarklet, helperConsoleSource } from "@/lib/importHelper";

const helperName = "Move to Kite";

// What differs between the bot makers that can be imported from: where the
// helper runs, what it copies and how that's read.
interface ImportSource {
  // Used in the analytics events.
  id: "botghost" | "inventor";
  name: string;
  dashboardUrl: string;
  // What to do in the dashboard, up to clicking the bookmark.
  copyStep: string;
  // What the helper finds, like "commands and custom events".
  what: string;
}

// readExport reads what the helper copied, or throws a message for the user.
// The helper is the same for both bot makers, so either page takes either's
// text. It says which of them it was, for the analytics.
function readExport(text: string): {
  id: ImportSource["id"];
  items: ConvertedItem[];
} {
  if (text.includes('"kite_inventor_export"')) {
    return {
      id: "inventor",
      items: convertInventorItems(parseInventorExport(text)),
    };
  }
  const parsed = parseBotGhostExport(text);
  return {
    id: "botghost",
    items: convertBotGhostItems([...parsed.commands, ...parsed.events]),
  };
}

const botGhostSource: ImportSource = {
  id: "botghost",
  name: "BotGhost",
  dashboardUrl: "https://dashboard.botghost.com",
  copyStep:
    "Open the bot you want to move in the BotGhost dashboard and go to its Commands page. Click the Move to Kite bookmark, then Copy for Kite.",
  what: "commands and custom events",
};

const inventorSource: ImportSource = {
  id: "inventor",
  name: "inventor.bot",
  dashboardUrl: "https://dash.inventor.bot",
  copyStep:
    "Open the bot you want to move in the inventor.bot dashboard, on any of its pages. Click the Move to Kite bookmark, then Copy for Kite.",
  what: "commands",
};

type Imported = {
  item: ConvertedItem;
  // The ID it got in Kite, or why it couldn't be added.
  id?: string;
  enabled?: boolean;
  error?: string;
};

function itemLabel(item: ConvertedItem) {
  return item.kind === "command"
    ? `/${item.name}`
    : `${eventListenerLabel("discord", item.eventType)} (${item.sourceName})`;
}

export default function BotGhostImport() {
  return <BuilderImport source={botGhostSource} />;
}

export function InventorImport() {
  return <BuilderImport source={inventorSource} />;
}

function BuilderImport({ source }: { source: ImportSource }) {
  const [items, setItems] = useState<ConvertedItem[] | null>(null);
  const [imported, setImported] = useState<Imported[] | null>(null);

  if (imported) {
    return (
      <ImportDone
        imported={imported}
        onRestart={() => {
          setImported(null);
          setItems(null);
        }}
      />
    );
  }
  if (items) {
    return (
      <ImportReview
        source={source}
        items={items}
        onBack={() => setItems(null)}
        onImported={setImported}
      />
    );
  }
  return <ImportSteps source={source} onItems={setItems} />;
}

function Step({
  index,
  title,
  children,
}: {
  index: number;
  title: string;
  children: ReactNode;
}) {
  return (
    <Card className="p-5 flex gap-4">
      <div className="flex-none h-7 w-7 rounded-full border-2 border-primary text-primary flex items-center justify-center text-sm font-semibold">
        {index}
      </div>
      <div className="min-w-0 flex-1 space-y-3">
        <div className="font-medium pt-0.5">{title}</div>
        {children}
      </div>
    </Card>
  );
}

function ImportSteps({
  source,
  onItems,
}: {
  source: ImportSource;
  onItems: (items: ConvertedItem[]) => void;
}) {
  const [text, setText] = useState("");
  const bookmark = useRef<HTMLAnchorElement>(null);
  const file = useRef<HTMLInputElement>(null);

  // React doesn't render javascript: addresses, so the bookmark gets its
  // address once it's on the page.
  useEffect(() => {
    bookmark.current?.setAttribute("href", helperBookmarklet);
  }, []);

  function read(value: string) {
    try {
      const { id, items } = readExport(value);
      if (items.length === 0) {
        toast.error(
          `No ${source.name} ${source.what} were found in this text.`
        );
        return;
      }
      track(`${id}_import_read`, {
        items: items.length,
        unsupported: items.filter((i) => i.status === "unsupported").length,
      });
      onItems(items);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Couldn't read this text.");
    }
  }

  return (
    <div className="space-y-4 max-w-3xl">
      <Step index={1} title="Add the helper to your bookmarks">
        <p className="text-sm text-muted-foreground">
          BotGhost and inventor.bot have no export, so a small helper copies
          your commands out of their dashboards. Drag this button to the
          bookmarks bar of your browser. If you don&apos;t see the bar, press
          Ctrl+Shift+B, or Cmd+Shift+B on a Mac.
        </p>
        <div className="flex flex-wrap items-center gap-3">
          <a
            ref={bookmark}
            draggable
            onClick={(e) => {
              e.preventDefault();
              toast.info("Drag this button to your bookmarks bar.");
            }}
            className="inline-flex items-center gap-2 rounded-md border border-dashed border-primary bg-primary/10 px-3 py-2 text-sm font-medium text-primary cursor-grab"
          >
            <BookmarkIcon className="h-4 w-4" />
            {helperName}
          </a>
          <Button
            variant="ghost"
            size="sm"
            onClick={() =>
              navigator.clipboard
                .writeText(helperConsoleSource)
                .then(() =>
                  toast.success(
                    `Helper copied. Paste it into the console of the ${source.name} dashboard.`
                  )
                )
            }
          >
            Can&apos;t drag it? Copy the helper instead
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          The helper only reads the commands and events of the bot you have
          open, and only when you click it. It sends nothing anywhere: you paste
          what it copies here yourself.
        </p>
      </Step>

      <Step index={2} title={`Copy your commands in ${source.name}`}>
        <p className="text-sm text-muted-foreground">{source.copyStep}</p>
        <Button variant="outline" size="sm" asChild>
          <a href={source.dashboardUrl} target="_blank" rel="noreferrer">
            Open {source.name}
            <ExternalLinkIcon className="h-3.5 w-3.5 ml-2" />
          </a>
        </Button>
      </Step>

      <Step index={3} title="Paste them here">
        <Textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Paste what the helper copied"
          minRows={8}
          maxRows={8}
          className="resize-none break-all font-mono md:text-xs"
        />
        <input
          ref={file}
          type="file"
          accept=".json,application/json,text/plain"
          className="hidden"
          onChange={async (e) => {
            const picked = e.target.files?.[0];
            e.target.value = "";
            if (picked) read(await picked.text());
          }}
        />
        <div className="flex flex-wrap gap-2">
          <Button disabled={!text.trim()} onClick={() => read(text)}>
            Continue
          </Button>
          <Button variant="outline" onClick={() => file.current?.click()}>
            <UploadIcon className="h-4 w-4 mr-2" />
            Upload the saved file
          </Button>
        </div>
      </Step>
    </div>
  );
}

const statusBadges = {
  ready: { label: "Ready", className: "bg-green-500/15 text-green-600" },
  review: {
    label: "Check after importing",
    className: "bg-amber-500/15 text-amber-600",
  },
  unsupported: {
    label: "Can't be imported",
    className: "bg-muted text-muted-foreground",
  },
};

function StatusBadge({ status }: { status: ConvertedItem["status"] }) {
  const badge = statusBadges[status];
  return (
    <Badge
      variant="outline"
      className={`border-0 flex-none whitespace-nowrap ${badge.className}`}
    >
      {badge.label}
    </Badge>
  );
}

function Notes({ notes }: { notes: string[] }) {
  if (notes.length === 0) return null;
  return (
    <ul className="mt-2 space-y-1.5 text-sm text-amber-600 dark:text-amber-500">
      {notes.map((note) => (
        <li key={note} className="flex gap-2">
          <CircleAlertIcon className="h-4 w-4 flex-none mt-0.5" />
          <span>{note}</span>
        </li>
      ))}
    </ul>
  );
}

function ImportReview({
  source,
  items,
  onBack,
  onImported,
}: {
  source: ImportSource;
  items: ConvertedItem[];
  onBack: () => void;
  onImported: (imported: Imported[]) => void;
}) {
  const appId = useAppId();
  const commands = useCommands();
  const listeners = useEventListeners();
  const commandLimit = usePlanLimit(
    "max_commands",
    commands?.filter((c) => c?.enabled).length
  );
  const listenerLimit = usePlanLimit(
    "max_event_listeners",
    listeners?.filter((l) => l?.enabled && l.source === "discord").length
  );
  const commandsImport = useCommandsImportMutation(appId);
  const listenersImport = useEventListenersImportMutation(appId);

  const importable = useMemo(() => items.filter((i) => i.flow), [items]);
  const unsupported = items.filter((i) => !i.flow);
  const [selected, setSelected] = useState(() => new Set(importable));
  const [progress, setProgress] = useState<number | null>(null);

  const picked = importable.filter((i) => selected.has(i));
  // What the plan still has room for runs right away, the rest is added
  // turned off.
  const room = {
    command: commandLimit
      ? Math.max(0, commandLimit.limit - commandLimit.count)
      : Infinity,
    event: listenerLimit
      ? Math.max(0, listenerLimit.limit - listenerLimit.count)
      : Infinity,
  };
  const off =
    Math.max(
      0,
      picked.filter((i) => i.kind === "command").length - room.command
    ) +
    Math.max(0, picked.filter((i) => i.kind === "event").length - room.event);

  async function run() {
    const left = { ...room };
    const imported: Imported[] = [];
    // One at a time, so one that fails doesn't stop the others.
    for (const [i, item] of picked.entries()) {
      setProgress(i);
      const enabled = left[item.kind] > 0;
      try {
        const res =
          item.kind === "command"
            ? await commandsImport.mutateAsync({
                commands: [{ flow_source: item.flow!, enabled }],
              })
            : await listenersImport.mutateAsync({
                event_listeners: [
                  { source: "discord", flow_source: item.flow!, enabled },
                ],
              });
        if (!res.success) {
          imported.push({ item, error: res.error.message });
          continue;
        }
        if (enabled) left[item.kind]--;
        imported.push({ item, id: res.data[0]?.id, enabled });
        track(
          `${item.kind === "command" ? "command" : "event_listener"}_created`,
          {
            source: "import",
            via: source.id,
          }
        );
      } catch (e) {
        imported.push({
          item,
          error: e instanceof Error ? e.message : "Something went wrong",
        });
      }
    }
    setProgress(null);
    onImported(imported);
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm">
        <Button variant="ghost" size="sm" onClick={onBack} className="-ml-2">
          <ArrowLeftIcon className="h-4 w-4 mr-2" />
          Back
        </Button>
        <span className="text-muted-foreground">
          Found {items.length} in your {source.name} bot.{" "}
          {importable.filter((i) => i.status === "ready").length} are ready,{" "}
          {importable.filter((i) => i.status === "review").length} need a check
          after importing
          {unsupported.length > 0 &&
            `, and ${unsupported.length} can't be imported`}
          .
        </span>
      </div>

      <div className="space-y-3">
        {importable.map((item, i) => (
          <div key={i}>
            <FlowItemCard
              icon={
                item.kind === "command"
                  ? SlashSquareIcon
                  : eventSourceName("discord").icon
              }
              name={itemLabel(item)}
              description={item.description}
              context={item.kind === "command" ? "command" : "event_discord"}
              flow={item.flow!}
              badge={<StatusBadge status={item.status} />}
              checkbox={{
                checked: selected.has(item),
                disabled: progress !== null,
                onCheckedChange: (checked) =>
                  setSelected((old) => {
                    const next = new Set(old);
                    if (checked) next.add(item);
                    else next.delete(item);
                    return next;
                  }),
              }}
            />
            <Notes notes={item.notes} />
          </div>
        ))}
      </div>

      {unsupported.length > 0 && (
        <div>
          <div className="text-sm font-medium mb-2.5">
            Can&apos;t be imported
          </div>
          <div className="space-y-2">
            {unsupported.map((item, i) => (
              <Card key={i} className="p-4 flex gap-3 items-start">
                <CircleXIcon className="h-5 w-5 text-muted-foreground flex-none" />
                <div className="min-w-0 text-sm">
                  <div className="font-medium">
                    {item.kind === "command"
                      ? `/${item.name}`
                      : item.sourceName}
                  </div>
                  <div className="text-muted-foreground">
                    {item.notes.join(" ")}
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}

      <div className="sticky bottom-0 -mx-4 px-4 py-4 border-t bg-background flex flex-wrap items-center justify-between gap-3">
        <div className="text-sm text-muted-foreground max-w-xl">
          {off > 0
            ? `Your plan has room for ${
                picked.length - off
              } more to run. The other ${off} are added turned off.`
            : `Copies are added to your app. Nothing changes in ${source.name}.`}
        </div>
        <LoadingButton
          loading={progress !== null}
          disabled={picked.length === 0}
          onClick={run}
        >
          {progress !== null
            ? `Importing ${progress + 1} of ${picked.length}`
            : `Import ${picked.length} selected`}
        </LoadingButton>
      </div>
    </div>
  );
}

function ImportDone({
  imported,
  onRestart,
}: {
  imported: Imported[];
  onRestart: () => void;
}) {
  const appId = useAppId();
  const added = imported.filter((i) => i.id);
  const failed = imported.filter((i) => !i.id);

  return (
    <div className="space-y-6 max-w-3xl">
      <Card className="p-5 flex items-center gap-4">
        <div className="flex-none h-9 w-9 rounded-full bg-primary text-primary-foreground flex items-center justify-center">
          <CheckIcon className="h-5 w-5" />
        </div>
        <div className="flex-auto">
          <div className="font-semibold">
            Imported {added.length} of {imported.length}
          </div>
          <div className="text-sm text-muted-foreground">
            Commands show up in Discord within a minute. Open the ones marked
            below first, they have a note on the canvas with what to check.
          </div>
        </div>
      </Card>

      <div className="space-y-2">
        {added.map(({ item, id, enabled }) => (
          <Card key={id} className="p-4 flex flex-wrap items-center gap-3">
            <div className="min-w-0 flex-1">
              <div className="font-medium truncate">{itemLabel(item)}</div>
              {!enabled && (
                <div className="text-sm text-muted-foreground">
                  Added turned off, as your plan has no room for it.
                </div>
              )}
            </div>
            <StatusBadge status={item.status} />
            <Button variant="outline" size="sm" asChild>
              <Link
                href={{
                  pathname:
                    item.kind === "command"
                      ? "/apps/[appId]/commands/[cmdId]"
                      : "/apps/[appId]/events/[eventId]",
                  query:
                    item.kind === "command"
                      ? { appId, cmdId: id }
                      : { appId, eventId: id },
                }}
              >
                Open
              </Link>
            </Button>
          </Card>
        ))}
        {failed.map(({ item, error }, i) => (
          <Card key={i} className="p-4 flex gap-3 items-start">
            <CircleXIcon className="h-5 w-5 text-destructive flex-none" />
            <div className="min-w-0 text-sm">
              <div className="font-medium">{itemLabel(item)}</div>
              <div className="text-muted-foreground">{error}</div>
            </div>
          </Card>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        <Button asChild>
          <Link href={{ pathname: "/apps/[appId]/commands", query: { appId } }}>
            Go to commands
          </Link>
        </Button>
        <Button variant="outline" onClick={onRestart}>
          Import more
        </Button>
      </div>
    </div>
  );
}
