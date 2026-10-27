import { ReactNode, useState } from "react";
import { useRouter } from "next/router";
import { useQueryClient } from "@tanstack/react-query";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import LoadingButton from "../common/LoadingButton";
import {
  useCommandsImportMutation,
  useEventListenersImportMutation,
  useMessagesImportMutation,
} from "@/lib/api/mutations";
import { apiRequest } from "@/lib/api/client";
import { useAppId } from "@/lib/hooks/params";
import { useAppSecrets, useMessages, useVariables } from "@/lib/hooks/api";
import { missingSecretNames } from "@/lib/flow/secrets";
import {
  SharedMessage,
  flowSourcesReference,
  isSharedMessage,
  remapFlowReferences,
  remapFlowSources,
} from "@/lib/flow/messageTemplates";
import { FlowData } from "@/lib/types/flow.gen";
import {
  MessageDeleteResponse,
  MessageUpdateResponse,
} from "@/lib/types/wire.gen";
import { toast } from "sonner";
import { ShareCodeImportForm } from "./ShareCodeImportForm";
import {
  EventListenerSource,
  eventListenerLabel,
  eventSourceName,
  eventSourceOfFlow,
} from "@/lib/flow/eventTypes";
import { eventFlowContext } from "@/lib/flow/bundleExport";
import { track } from "@/lib/analytics";
import { CircleAlertIcon, SlashSquareIcon } from "lucide-react";
import {
  FlowItemCard,
  MessageItemCard,
  PreviewItemsProvider,
} from "./ItemPreviewCards";

const kinds = {
  command: {
    label: "command",
    entryNodeType: "entry_command",
    href: (appId: string, id: string) => `/apps/${appId}/commands/${id}`,
  },
  event_listener: {
    label: "event listener",
    entryNodeType: "entry_event",
    href: (appId: string, id: string) => `/apps/${appId}/events/${id}`,
  },
  message: {
    label: "message template",
    entryNodeType: null,
    href: (appId: string, id: string) => `/apps/${appId}/messages/${id}`,
  },
};

type Kind = keyof typeof kinds;

type ShareData = {
  flow_source?: FlowData;
  source?: string;
  message?: SharedMessage;
  // Message templates the flow (or message) uses, exported along with it.
  messages?: SharedMessage[];
};

export default function FlowImportDialog({
  kind,
  eventSource,
  children,
}: {
  kind: Kind;
  // Only event listeners of this source can be imported, as each source has
  // its own page.
  eventSource?: EventListenerSource;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      {/* Wider while the preview of what's imported shows. */}
      <DialogContent className="max-w-lg has-[[data-preview]]:max-w-3xl max-h-[90vh] overflow-y-auto">
        <ImportForm
          kind={kind}
          eventSource={eventSource}
          onImported={() => setOpen(false)}
        />
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so the variable and message queries only run while it's open.
function ImportForm({
  kind,
  eventSource,
  onImported,
}: {
  kind: Kind;
  eventSource?: EventListenerSource;
  onImported: () => void;
}) {
  const [importing, setImporting] = useState(false);
  // The data is previewed before it's imported.
  const [preview, setPreview] = useState<{
    data: ShareData;
    via: "json" | "share_code";
  } | null>(null);

  const router = useRouter();
  const queryClient = useQueryClient();
  const appId = useAppId();
  const variables = useVariables();
  const messages = useMessages();
  const secrets = useAppSecrets();

  const commandsImportMutation = useCommandsImportMutation(appId);
  const eventListenersImportMutation = useEventListenersImportMutation(appId);
  const messagesImportMutation = useMessagesImportMutation(appId);

  const { entryNodeType, href } = kinds[kind];
  const label =
    kind === "event_listener" && eventSource
      ? eventSourceName(eventSource).singular
      : kinds[kind].label;

  const loading = importing || !variables || !messages || !secrets;

  // Imports the templates and adds the id each one got in this app to idMap,
  // keyed by its id in the original app. Templates that already exist in this
  // app (exported from the same app) are reused unless forced, so importing
  // twice doesn't pile up copies. Returns the number of cleared references.
  async function importMessages(
    shared: SharedMessage[],
    idMap: Map<string, string>,
    variableIds: Set<string>,
    existingIds: Set<string>,
    forceIds: Set<string>
  ) {
    const toCreate = shared.filter(
      (m) => forceIds.has(m.id) || !existingIds.has(m.id)
    );
    if (toCreate.length === 0) return 0;

    // The new ids aren't known yet, so references between the imported
    // templates are kept for now and pointed at the copies afterwards.
    const keepIds = new Set([...existingIds, ...toCreate.map((m) => m.id)]);

    let removed = 0;
    const requests = toCreate.map((m) => {
      const res = remapFlowSources(m.flow_sources, {
        variableIds,
        messageIds: keepIds,
      });
      removed += res.removed;
      return {
        name: m.name.slice(0, 100),
        description: m.description?.slice(0, 255) || null,
        data: m.data,
        flow_sources: res.flowSources,
      };
    });

    const res = await messagesImportMutation.mutateAsync({
      messages: requests,
    });
    if (!res.success) {
      throw new Error(
        `Failed to import message templates: ${res.error.message} (${res.error.code})`
      );
    }

    toCreate.forEach((m, i) => {
      const created = res.data[i];
      if (created) idMap.set(m.id, created.id);
    });

    const originalIds = new Set(idMap.keys());
    for (const [i, m] of toCreate.entries()) {
      const created = res.data[i];
      const req = requests[i];
      if (!created || !flowSourcesReference(req.flow_sources, originalIds)) {
        continue;
      }

      const { flowSources } = remapFlowSources(req.flow_sources, {
        variableIds,
        messageIds: existingIds,
        messageIdMap: idMap,
      });
      // The update mutation hook is bound to a single message id.
      const updateRes = await apiRequest<MessageUpdateResponse>(
        `/v1/apps/${appId}/messages/${created.id}`,
        {
          method: "PATCH",
          body: JSON.stringify({ ...req, flow_sources: flowSources }),
          headers: { "Content-Type": "application/json" },
        }
      );
      if (!updateRes.success) {
        throw new Error(
          `Failed to link message template ${m.name}: ${updateRes.error.message} (${updateRes.error.code})`
        );
      }
    }

    queryClient.invalidateQueries({ queryKey: ["apps", appId, "messages"] });
    return removed;
  }

  // Deletes the templates of a failed import, so no orphaned copies are left.
  async function deleteMessages(ids: string[]) {
    if (ids.length === 0) return;
    await Promise.allSettled(
      ids.map((id) =>
        apiRequest<MessageDeleteResponse>(`/v1/apps/${appId}/messages/${id}`, {
          method: "DELETE",
        })
      )
    );
    queryClient.invalidateQueries({ queryKey: ["apps", appId, "messages"] });
  }

  // previewShareData shows the data before it's imported, if it's for this
  // page.
  function previewShareData(
    parsed: ShareData | null | undefined,
    via: "json" | "share_code"
  ) {
    const flow = parsed?.flow_source;

    if (kind === "message") {
      if (!isSharedMessage(parsed?.message)) {
        toast.error(`Invalid share code, make sure it's for a ${label}`);
        return;
      }
    } else if (
      !Array.isArray(flow?.nodes) ||
      !Array.isArray(flow?.edges) ||
      !flow.nodes.some((n) => n.type === entryNodeType)
    ) {
      toast.error(`Invalid share code, make sure it's for a ${label}`);
      return;
    }

    // An event listener belongs on the page of its source, e.g. a webhook
    // can't be imported on the page of Discord events.
    if (kind === "event_listener" && eventSource) {
      const flowSource = eventSourceOfFlow(flow);
      if (flowSource !== eventSource) {
        const other = eventSourceName(flowSource);
        toast.error(
          `This is a ${other.singular}, not a ${label}. Import it on the ${other.plural} page.`
        );
        return;
      }
    }

    setPreview({ data: parsed!, via });
  }

  // sharedMessages are the message templates of the data, the imported one
  // first if it's a message template.
  function sharedMessages(parsed: ShareData) {
    const bundled = (
      Array.isArray(parsed.messages) ? parsed.messages : []
    ).filter(isSharedMessage);
    // Each template is created once, even if the JSON lists it twice.
    const seenIds = new Set<string>();
    return (
      kind === "message" ? [parsed.message!, ...bundled] : bundled
    ).filter((m) => !seenIds.has(m.id) && seenIds.add(m.id));
  }

  // usedFlows are all flows of the data, to find the secrets they use.
  function usedFlows(parsed: ShareData, shared: SharedMessage[]) {
    return [
      ...(parsed.flow_source ? [parsed.flow_source] : []),
      ...shared.flatMap((m) => Object.values(m.flow_sources ?? {})),
    ];
  }

  async function importShareData(
    parsed: ShareData,
    via: "json" | "share_code"
  ) {
    const flow = parsed.flow_source;

    const variableIds = new Set(variables?.flatMap((v) => (v ? [v.id] : [])));
    const existingIds = new Set(messages?.flatMap((m) => (m ? [m.id] : [])));

    const shared = sharedMessages(parsed);

    const idMap = new Map<string, string>();
    setImporting(true);
    try {
      const missingSecrets = missingSecretNames(
        usedFlows(parsed, shared),
        secrets?.map((s) => s?.name) ?? []
      );

      let removed = await importMessages(
        shared,
        idMap,
        variableIds,
        existingIds,
        new Set(kind === "message" ? [shared[0].id] : [])
      );

      let importedId: string | undefined;
      if (kind === "message") {
        importedId = idMap.get(shared[0].id);
      } else {
        const remapped = remapFlowReferences(flow!, {
          variableIds,
          messageIds: existingIds,
          messageIdMap: idMap,
        });
        removed += remapped.removed;

        const res =
          kind === "command"
            ? await commandsImportMutation.mutateAsync({
                commands: [{ flow_source: remapped.flow, enabled: true }],
              })
            : await eventListenersImportMutation.mutateAsync({
                event_listeners: [
                  {
                    source: eventSourceOfFlow(flow),
                    flow_source: remapped.flow,
                    enabled: true,
                  },
                ],
              });
        if (!res.success) {
          throw new Error(
            `Failed to import ${label}: ${res.error.message} (${res.error.code})`
          );
        }
        importedId = res.data[0]?.id;
      }
      track(`${kind}_created`, {
        source: "import",
        via,
      });

      const extraTemplates = idMap.size - (kind === "message" ? 1 : 0);
      toast.success(
        extraTemplates > 0
          ? `Imported ${label} with ${extraTemplates} message template(s)!`
          : `Imported ${label}!`
      );
      if (removed > 0) {
        toast.warning(
          `${removed} block(s) referenced variables or message templates from another app, reselect them in the editor.`
        );
      }
      if (missingSecrets.length > 0) {
        toast.warning(
          `This ${label} uses secrets your app doesn't have: ${missingSecrets.join(
            ", "
          )}. Create them under Secrets.`
        );
      }
      onImported();

      if (importedId) {
        const id = importedId;
        setTimeout(() => router.push(href(appId, id)), 500);
      }
    } catch (e) {
      await deleteMessages(Array.from(idMap.values()));
      toast.error(e instanceof Error ? e.message : `Failed to import ${label}`);
    } finally {
      setImporting(false);
    }
  }

  if (preview) {
    const shared = sharedMessages(preview.data);
    const existingIds = new Set(messages?.flatMap((m) => (m ? [m.id] : [])));

    // The blocks use the shared message templates, which the app may not
    // have yet. Variables from another app aren't imported.
    return (
      <PreviewItemsProvider
        messages={shared.map((m) => ({ key: m.id, name: m.name }))}
      >
        <ImportPreview
          kind={kind}
          label={label}
          data={preview.data}
          // Templates that already exist in this app are reused.
          extraMessages={shared
            .slice(kind === "message" ? 1 : 0)
            .filter((m) => !existingIds.has(m.id))}
          missingSecrets={missingSecretNames(
            usedFlows(preview.data, shared),
            secrets?.map((s) => s?.name) ?? []
          )}
          loading={loading}
          onBack={() => setPreview(null)}
          onImport={() => importShareData(preview.data, preview.via)}
        />
      </PreviewItemsProvider>
    );
  }

  return (
    <ShareCodeImportForm
      title={`Import ${label}`}
      jsonPlaceholder={
        kind === "message" ? '{"message": ...}' : '{"flow_source": ...}'
      }
      submitLabel="Continue"
      loading={loading}
      onData={(data, type) => {
        if (type === null) {
          previewShareData(data as ShareData, "json");
        } else if (type === "bundle") {
          toast.error(
            "This share code has several items, import it under Templates"
          );
        } else if (type !== kind) {
          toast.error(`This share code is not for a ${label}`);
        } else {
          previewShareData(data as ShareData, "share_code");
        }
      }}
    />
  );
}

// ImportPreview shows what's imported, so it can be checked before it's added
// to the app.
function ImportPreview({
  kind,
  label,
  data,
  extraMessages,
  missingSecrets,
  loading,
  onBack,
  onImport,
}: {
  kind: Kind;
  label: string;
  data: ShareData;
  extraMessages: SharedMessage[];
  missingSecrets: string[];
  loading: boolean;
  onBack: () => void;
  onImport: () => void;
}) {
  const flow = data.flow_source;
  const entry = flow?.nodes.find(
    (n) => n.type === kinds[kind].entryNodeType
  )?.data;
  const source = eventSourceOfFlow(flow);

  return (
    <>
      <DialogHeader>
        <DialogTitle>Import {label}</DialogTitle>
        <DialogDescription>
          A copy is added to your app, so you can change it like anything you
          built yourself.
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-5" data-preview>
        {kind === "message" ? (
          <MessageItemCard
            name={data.message!.name}
            description={data.message!.description ?? null}
            data={data.message!.data}
            flowSources={data.message!.flow_sources ?? {}}
            defaultOpen
          />
        ) : (
          <FlowItemCard
            icon={
              kind === "command"
                ? SlashSquareIcon
                : eventSourceName(source).icon
            }
            name={
              kind === "command"
                ? `/${String(entry?.name ?? "")}`
                : eventListenerLabel(source, String(entry?.event_type ?? ""))
            }
            description={String(entry?.description ?? "")}
            context={kind === "command" ? "command" : eventFlowContext(source)}
            flow={flow!}
            defaultOpen
          />
        )}

        {!!extraMessages.length && (
          <div>
            <div className="text-sm font-medium mb-2.5">
              Also adds the message templates it uses
            </div>
            <div className="space-y-3">
              {extraMessages.map((message) => (
                <MessageItemCard
                  key={message.id}
                  name={message.name}
                  description={message.description ?? null}
                  data={message.data}
                  flowSources={message.flow_sources ?? {}}
                />
              ))}
            </div>
          </div>
        )}

        {!!missingSecrets.length && (
          <div className="flex gap-2 text-sm text-amber-600 dark:text-amber-500">
            <CircleAlertIcon className="h-4 w-4 flex-none mt-0.5" />
            <span>
              This {label} uses secrets your app doesn&apos;t have:{" "}
              {missingSecrets.join(", ")}. Create them under Secrets after
              importing.
            </span>
          </div>
        )}
      </div>

      {/* Stays in view while the items scroll. */}
      <DialogFooter className="sticky -bottom-6 -mx-6 -mb-6 px-6 py-4 border-t bg-background flex-row justify-between gap-2 sm:justify-between sm:space-x-0">
        <Button variant="ghost" onClick={onBack}>
          Back
        </Button>
        <LoadingButton onClick={onImport} loading={loading}>
          Import
        </LoadingButton>
      </DialogFooter>
    </>
  );
}
