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
  useFunctionsImportMutation,
  useMessagesImportMutation,
  useShareCodeResolveMutation,
} from "@/lib/api/mutations";
import { apiRequest } from "@/lib/api/client";
import { useAppId } from "@/lib/hooks/params";
import {
  useAppSecrets,
  useFunctions,
  useMessages,
  useVariables,
} from "@/lib/hooks/api";
import { getReferencedSecretNames } from "@/lib/flow/secrets";
import {
  SharedMessage,
  isSharedMessage,
  remapFlowReferences,
  remapFlowSources,
} from "@/lib/flow/messageTemplates";
import {
  SharedFunction,
  getFunctionName,
  getMissingFunctions,
  isSharedFunction,
} from "@/lib/flow/functions";
import { FlowData } from "@/lib/types/flow.gen";
import {
  FunctionDeleteResponse,
  FunctionUpdateResponse,
  MessageDeleteResponse,
  MessageUpdateResponse,
} from "@/lib/types/wire.gen";
import { toast } from "sonner";
import { ShareCodeInput, ShareCodePanel } from "./ShareCode";
import { BracesIcon, KeyRoundIcon, SquareFunctionIcon } from "lucide-react";

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
  function: {
    label: "function",
    entryNodeType: null,
    href: (appId: string, id: string) => `/apps/${appId}/functions/${id}`,
  },
};

type Kind = keyof typeof kinds;

type ShareData = {
  flow_source?: FlowData;
  source?: string;
  message?: SharedMessage;
  function?: SharedFunction;
  // Message templates and functions the flow (or message or function) uses,
  // exported along with it.
  messages?: SharedMessage[];
  functions?: SharedFunction[];
};

// A share whose flows call functions the app doesn't have, waiting for the
// user to decide whether to import those functions too.
type FunctionPrompt = {
  parsed: ShareData;
  functions: SharedFunction[];
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

function dedupeById<T extends { id: string }>(items: T[]) {
  const seen = new Set<string>();
  return items.filter((item) => !seen.has(item.id) && seen.add(item.id));
}

function sameFlows(a: unknown, b: unknown) {
  return JSON.stringify(a) === JSON.stringify(b);
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
  const [functionPrompt, setFunctionPrompt] = useState<FunctionPrompt | null>(
    null
  );

  const router = useRouter();
  const queryClient = useQueryClient();
  const appId = useAppId();
  const variables = useVariables();
  const messages = useMessages();
  const functions = useFunctions();
  const secrets = useAppSecrets();

  const commandsImportMutation = useCommandsImportMutation(appId);
  const eventListenersImportMutation = useEventListenersImportMutation(appId);
  const messagesImportMutation = useMessagesImportMutation(appId);
  const functionsImportMutation = useFunctionsImportMutation(appId);
  const shareCodeResolveMutation = useShareCodeResolveMutation();

  const { label, entryNodeType, href } = kinds[kind];

  const loading =
    importing ||
    shareCodeResolveMutation.isPending ||
    !variables ||
    !messages ||
    !functions ||
    !secrets;

  // Deletes what a failed import created, so no orphaned copies are left.
  async function deleteCreated(messageIds: string[], functionIds: string[]) {
    await Promise.allSettled([
      ...messageIds.map((id) =>
        apiRequest<MessageDeleteResponse>(`/v1/apps/${appId}/messages/${id}`, {
          method: "DELETE",
        })
      ),
      ...functionIds.map((id) =>
        apiRequest<FunctionDeleteResponse>(
          `/v1/apps/${appId}/functions/${id}`,
          { method: "DELETE" }
        )
      ),
    ]);
    queryClient.invalidateQueries({ queryKey: ["apps", appId, "messages"] });
    queryClient.invalidateQueries({ queryKey: ["apps", appId, "functions"] });
  }

  // Checks the share and asks whether to import the functions it calls that
  // the app doesn't have, before importing it.
  function importShareData(parsed: ShareData | null | undefined) {
    const flow = parsed?.flow_source;

    if (kind === "message") {
      if (!isSharedMessage(parsed?.message)) {
        toast.error(`Invalid share code, make sure it's for a ${label}`);
        return;
      }
    } else if (kind === "function") {
      if (!isSharedFunction(parsed?.function)) {
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

    const bundledMessages = (
      Array.isArray(parsed?.messages) ? parsed.messages : []
    ).filter(isSharedMessage);
    const mainFunction = kind === "function" ? parsed!.function! : undefined;
    const bundledFunctions = dedupeById(
      (Array.isArray(parsed?.functions) ? parsed.functions : []).filter(
        isSharedFunction
      )
    ).filter((f) => f.id !== mainFunction?.id);

    const existingFunctionIds = new Set(
      functions?.flatMap((f) => (f ? [f.id] : []))
    );
    // The function being imported is created anyway, so it isn't missing.
    if (mainFunction) existingFunctionIds.add(mainFunction.id);

    const missing = getMissingFunctions(
      [
        flow,
        mainFunction?.flow_source,
        ...(kind === "message"
          ? Object.values(parsed!.message!.flow_sources ?? {})
          : []),
        ...bundledMessages.flatMap((m) => Object.values(m.flow_sources ?? {})),
      ],
      bundledFunctions,
      existingFunctionIds
    );

    if (missing.length > 0) {
      setFunctionPrompt({ parsed: parsed!, functions: missing });
      return;
    }

    runImport(parsed!, []);
  }

  // Imports the share, along with its message templates and the functions in
  // extraFunctions. Calls of functions that are neither in the app nor
  // imported are cleared, so the user picks a function again.
  async function runImport(
    parsed: ShareData,
    extraFunctions: SharedFunction[]
  ) {
    setFunctionPrompt(null);

    const flow = parsed.flow_source;
    const mainMessage = kind === "message" ? parsed.message! : undefined;
    const mainFunction = kind === "function" ? parsed.function! : undefined;

    const variableIds = new Set(variables?.flatMap((v) => (v ? [v.id] : [])));
    const existingMessageIds = new Set(
      messages?.flatMap((m) => (m ? [m.id] : []))
    );
    const existingFunctionIds = new Set(
      functions?.flatMap((f) => (f ? [f.id] : []))
    );

    const bundledMessages = (
      Array.isArray(parsed.messages) ? parsed.messages : []
    ).filter(isSharedMessage);
    // Each template and function is created once, even if the JSON lists it
    // twice. The template or function being imported is always created, the
    // others are reused if they already exist in this app (exported from the
    // same app), so importing twice doesn't pile up copies.
    const messagesToCreate = dedupeById([
      ...(mainMessage ? [mainMessage] : []),
      ...bundledMessages,
    ]).filter((m) => m.id === mainMessage?.id || !existingMessageIds.has(m.id));
    const functionsToCreate = dedupeById([
      ...(mainFunction ? [mainFunction] : []),
      ...extraFunctions,
    ]);

    // The new ids aren't known yet, so references between the imported
    // templates and functions are kept for now and pointed at the copies
    // afterwards. Everything else the app doesn't have is cleared.
    const keepMapping = {
      variableIds,
      messageIds: new Set([
        ...Array.from(existingMessageIds),
        ...messagesToCreate.map((m) => m.id),
      ]),
      functionIds: new Set([
        ...Array.from(existingFunctionIds),
        ...functionsToCreate.map((f) => f.id),
      ]),
    };

    const messageIdMap = new Map<string, string>();
    const functionIdMap = new Map<string, string>();
    const finalMapping = {
      variableIds,
      messageIds: existingMessageIds,
      messageIdMap,
      functionIds: existingFunctionIds,
      functionIdMap,
    };

    setImporting(true);
    try {
      const usedFlows = [
        ...(flow ? [flow] : []),
        ...messagesToCreate.flatMap((m) => Object.values(m.flow_sources ?? {})),
        ...functionsToCreate.map((f) => f.flow_source),
      ];
      const secretNames = new Set(secrets?.map((s) => s?.name));
      const missingSecrets = Array.from(
        new Set(usedFlows.flatMap((f) => getReferencedSecretNames(f)))
      ).filter((name) => !secretNames.has(name));

      let removed = 0;

      const messageRequests = messagesToCreate.map((m) => {
        const res = remapFlowSources(m.flow_sources, keepMapping);
        removed += res.removed;
        return {
          name: m.name.slice(0, 100),
          description: m.description?.slice(0, 255) || null,
          data: m.data,
          flow_sources: res.flowSources,
        };
      });
      if (messageRequests.length > 0) {
        const res = await messagesImportMutation.mutateAsync({
          messages: messageRequests,
        });
        if (!res.success) {
          throw new Error(
            `Failed to import message templates: ${res.error.message} (${res.error.code})`
          );
        }
        messagesToCreate.forEach((m, i) => {
          const created = res.data[i];
          if (created) messageIdMap.set(m.id, created.id);
        });
      }

      const functionRequests = functionsToCreate.map((f) => {
        const res = remapFlowReferences(f.flow_source, keepMapping);
        removed += res.removed;
        return { flow_source: res.flow };
      });
      if (functionRequests.length > 0) {
        const res = await functionsImportMutation.mutateAsync({
          functions: functionRequests,
        });
        if (!res.success) {
          throw new Error(
            `Failed to import functions: ${res.error.message} (${res.error.code})`
          );
        }
        functionsToCreate.forEach((f, i) => {
          const created = res.data[i];
          if (created) functionIdMap.set(f.id, created.id);
        });
      }

      // Now that every copy exists, point the copies at each other.
      for (const [i, m] of messagesToCreate.entries()) {
        const id = messageIdMap.get(m.id);
        const req = messageRequests[i];
        const { flowSources } = remapFlowSources(
          req.flow_sources,
          finalMapping
        );
        if (!id || sameFlows(flowSources, req.flow_sources)) continue;

        // The update mutation hook is bound to a single message id.
        const updateRes = await apiRequest<MessageUpdateResponse>(
          `/v1/apps/${appId}/messages/${id}`,
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
      for (const [i, f] of functionsToCreate.entries()) {
        const id = functionIdMap.get(f.id);
        const req = functionRequests[i];
        const { flow: remapped } = remapFlowReferences(
          req.flow_source,
          finalMapping
        );
        if (!id || sameFlows(remapped, req.flow_source)) continue;

        const updateRes = await apiRequest<FunctionUpdateResponse>(
          `/v1/apps/${appId}/functions/${id}`,
          {
            method: "PATCH",
            body: JSON.stringify({ flow_source: remapped }),
            headers: { "Content-Type": "application/json" },
          }
        );
        if (!updateRes.success) {
          throw new Error(
            `Failed to link function ${getFunctionName(f.flow_source)}: ${
              updateRes.error.message
            } (${updateRes.error.code})`
          );
        }
      }
      queryClient.invalidateQueries({ queryKey: ["apps", appId, "messages"] });
      queryClient.invalidateQueries({ queryKey: ["apps", appId, "functions"] });

      let importedId: string | undefined;
      if (mainMessage) {
        importedId = messageIdMap.get(mainMessage.id);
      } else if (mainFunction) {
        importedId = functionIdMap.get(mainFunction.id);
      } else {
        const remapped = remapFlowReferences(flow!, finalMapping);
        removed += remapped.removed;

        const res =
          kind === "command"
            ? await commandsImportMutation.mutateAsync({
                commands: [{ flow_source: remapped.flow, enabled: true }],
              })
            : await eventListenersImportMutation.mutateAsync({
                event_listeners: [
                  {
                    source: parsed.source ?? "discord",
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

      const extraTemplates = messageIdMap.size - (mainMessage ? 1 : 0);
      const extraFunctionCount = functionIdMap.size - (mainFunction ? 1 : 0);
      const extras = [
        extraTemplates > 0 ? `${extraTemplates} message template(s)` : null,
        extraFunctionCount > 0 ? `${extraFunctionCount} function(s)` : null,
      ].filter(Boolean);
      toast.success(
        extras.length > 0
          ? `Imported ${label} with ${extras.join(" and ")}!`
          : `Imported ${label}!`
      );
      if (removed > 0) {
        toast.warning(
          `${removed} block(s) referenced variables, message templates or functions your app doesn't have, reselect them in the editor.`
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
      await deleteCreated(
        Array.from(messageIdMap.values()),
        Array.from(functionIdMap.values())
      );
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

  if (functionPrompt) {
    const count = functionPrompt.functions.length;
    return (
      <>
        <DialogHeader>
          <DialogTitle>Import functions too?</DialogTitle>
          <DialogDescription>
            This {label} uses{" "}
            {count === 1 ? "a function" : `${count} functions`} your app
            doesn&apos;t have. Import {count === 1 ? "it" : "them"} as well,
            already connected to the {label}?
          </DialogDescription>
        </DialogHeader>
        <ul className="space-y-2">
          {functionPrompt.functions.map((f) => (
            <li
              key={f.id}
              className="flex items-center space-x-2 rounded-md bg-muted px-3 py-2 text-sm"
            >
              <SquareFunctionIcon className="h-4 w-4 flex-none text-muted-foreground" />
              <span className="truncate">{getFunctionName(f.flow_source)}</span>
            </li>
          ))}
        </ul>
        <p className="text-sm text-muted-foreground">
          Without them, the {label} is still imported, but its Call function
          blocks have no function selected until you pick one.
        </p>
        <DialogFooter className="gap-2 sm:space-x-0">
          <Button
            variant="outline"
            disabled={importing}
            onClick={() => runImport(functionPrompt.parsed, [])}
          >
            Import without
          </Button>
          <LoadingButton
            loading={importing}
            onClick={() =>
              runImport(functionPrompt.parsed, functionPrompt.functions)
            }
          >
            Import with functions
          </LoadingButton>
        </DialogFooter>
      </>
    );
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
            kind === "message"
              ? '{"message": ...}'
              : kind === "function"
              ? '{"function": ...}'
              : '{"flow_source": ...}'
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
