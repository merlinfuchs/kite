import { ReactNode, useState } from "react";
import { useRouter } from "next/router";
import { useQueryClient } from "@tanstack/react-query";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import {
  useCommandsImportMutation,
  useEventListenersImportMutation,
  useMessagesImportMutation,
  useShareCodeResolveMutation,
} from "@/lib/api/mutations";
import { apiRequest } from "@/lib/api/client";
import { useAppId } from "@/lib/hooks/params";
import { useAppSecrets, useMessages, useVariables } from "@/lib/hooks/api";
import { getReferencedSecretNames } from "@/lib/flow/secrets";
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
import { ShareCodeInput, ShareCodePanel } from "./ShareCode";
import { BracesIcon, KeyRoundIcon } from "lucide-react";

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
  children,
}: {
  kind: Kind;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-lg">
        <ImportForm kind={kind} onImported={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so the variable and message queries only run while it's open.
function ImportForm({
  kind,
  onImported,
}: {
  kind: Kind;
  onImported: () => void;
}) {
  const [useJson, setUseJson] = useState(false);
  const [json, setJson] = useState("");
  const [code, setCode] = useState("");
  const [importing, setImporting] = useState(false);

  const router = useRouter();
  const queryClient = useQueryClient();
  const appId = useAppId();
  const variables = useVariables();
  const messages = useMessages();
  const secrets = useAppSecrets();

  const commandsImportMutation = useCommandsImportMutation(appId);
  const eventListenersImportMutation = useEventListenersImportMutation(appId);
  const messagesImportMutation = useMessagesImportMutation(appId);
  const shareCodeResolveMutation = useShareCodeResolveMutation();

  const { label, entryNodeType, href } = kinds[kind];

  const loading =
    importing ||
    shareCodeResolveMutation.isPending ||
    !variables ||
    !messages ||
    !secrets;

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

  async function importShareData(parsed: ShareData | null | undefined) {
    const flow = parsed?.flow_source;
    const mainMessage = parsed?.message;

    if (kind === "message") {
      if (!isSharedMessage(mainMessage)) {
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

    const variableIds = new Set(variables?.flatMap((v) => (v ? [v.id] : [])));
    const existingIds = new Set(messages?.flatMap((m) => (m ? [m.id] : [])));

    const bundled = (
      Array.isArray(parsed?.messages) ? parsed.messages : []
    ).filter(isSharedMessage);
    const shared = kind === "message" ? [mainMessage!, ...bundled] : bundled;

    const idMap = new Map<string, string>();
    setImporting(true);
    try {
      const usedFlows = [
        ...(flow ? [flow] : []),
        ...shared.flatMap((m) => Object.values(m.flow_sources ?? {})),
      ];
      const secretNames = new Set(secrets?.map((s) => s?.name));
      const missingSecrets = Array.from(
        new Set(usedFlows.flatMap((f) => getReferencedSecretNames(f)))
      ).filter((name) => !secretNames.has(name));

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
                    source: parsed?.source ?? "discord",
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

  function onImport() {
    if (useJson) {
      let parsed;
      try {
        parsed = JSON.parse(json);
      } catch {}
      importShareData(parsed);
      return;
    }

    if (!code) {
      toast.error("Enter a share code");
      return;
    }

    shareCodeResolveMutation.mutate(code, {
      onSuccess: (res) => {
        if (!res.success) {
          toast.error(
            res.error.code === "unknown_share_code"
              ? "Unknown share code"
              : `Failed to resolve share code: ${res.error.message} (${res.error.code})`
          );
          return;
        }
        if (res.data.type !== kind) {
          toast.error(`This share code is not for a ${label}`);
          return;
        }
        importShareData(res.data.data);
      },
    });
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>Import {label}</DialogTitle>
        <DialogDescription>
          {useJson
            ? "Paste the JSON that was exported from another app."
            : "Enter the share code that was exported from another app."}
        </DialogDescription>
      </DialogHeader>

      {useJson ? (
        <Textarea
          value={json}
          onChange={(e) => setJson(e.target.value)}
          placeholder={
            kind === "message" ? '{"message": ...}' : '{"flow_source": ...}'
          }
          minRows={8}
          maxRows={8}
          className="resize-none break-all font-mono md:text-xs"
          autoFocus
        />
      ) : (
        <ShareCodePanel>
          <ShareCodeInput
            value={code}
            onChange={setCode}
            onSubmit={() => !loading && onImport()}
          />
        </ShareCodePanel>
      )}

      <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
        <Button variant="ghost" onClick={() => setUseJson(!useJson)}>
          {useJson ? (
            <KeyRoundIcon className="mr-2 h-4 w-4" />
          ) : (
            <BracesIcon className="mr-2 h-4 w-4" />
          )}
          {useJson ? "Use share code" : "Use JSON"}
        </Button>
        <div className="flex flex-col-reverse gap-2 sm:flex-row">
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <LoadingButton onClick={onImport} loading={loading}>
            Import
          </LoadingButton>
        </div>
      </DialogFooter>
    </>
  );
}
