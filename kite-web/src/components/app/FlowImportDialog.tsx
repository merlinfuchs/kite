import { ReactNode, useState } from "react";
import { useRouter } from "next/router";
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
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useMessages, useVariables } from "@/lib/hooks/api";
import { FlowData } from "@/lib/types/flow.gen";
import {
  CommandsImportResponse,
  EventListenersImportResponse,
} from "@/lib/types/wire.gen";
import { APIResponse } from "@/lib/api/response";
import { apiRequest } from "@/lib/api/client";
import { toast } from "sonner";

async function resolveShareCode(code: string): Promise<string> {
  const res = await apiRequest<{ data: string }>(
    `/v1/share/${encodeURIComponent(code)}`
  );
  if (!res.success) {
    throw new Error(res.error.message || "Code not found");
  }
  return res.data.data;
}

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
};

type Kind = keyof typeof kinds;

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
      <DialogContent className="max-w-2xl">
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
  const [useCode, setUseCode] = useState(true);
  const [shareCode, setShareCode] = useState("");
  const [codeInput, setCodeInput] = useState("");
  const [resolving, setResolving] = useState(false);

  const router = useRouter();
  const appId = useAppId();
  const variables = useVariables();
  const messages = useMessages();

  const commandsImportMutation = useCommandsImportMutation(appId);
  const eventListenersImportMutation = useEventListenersImportMutation(appId);

  const { label, entryNodeType, href } = kinds[kind];

  // Same parsing/sanitization path regardless of whether the JSON came from
  // the textarea or was resolved from a share code.
  function importFlowData(raw: string) {
    let parsed: { flow_source?: FlowData; source?: string } | undefined;
    try {
      parsed = JSON.parse(raw);
    } catch {}

    const flow = parsed?.flow_source;
    if (
      !Array.isArray(flow?.nodes) ||
      !Array.isArray(flow?.edges) ||
      !flow.nodes.some((n) => n.type === entryNodeType)
    ) {
      toast.error(`Invalid share code, make sure it's for a ${label}`);
      return;
    }

    const { flow: sanitized, removed } = removeForeignReferences(
      flow,
      new Set(variables?.flatMap((v) => (v ? [v.id] : []))),
      new Set(messages?.flatMap((m) => (m ? [m.id] : [])))
    );

    const onSuccess = (
      res: APIResponse<CommandsImportResponse | EventListenersImportResponse>
    ) => {
      if (!res.success) {
        toast.error(
          `Failed to import ${label}: ${res.error.message} (${res.error.code})`
        );
        return;
      }

      toast.success(`Imported ${label}!`);
      if (removed > 0) {
        toast.warning(
          `${removed} block(s) referenced variables or message templates from another app, reselect them in the editor.`
        );
      }
      onImported();

      const imported = res.data[0];
      if (imported) {
        setTimeout(() => router.push(href(appId, imported.id)), 500);
      }
    };

    if (kind === "command") {
      commandsImportMutation.mutate(
        { commands: [{ flow_source: sanitized, enabled: true }] },
        { onSuccess }
      );
    } else {
      eventListenersImportMutation.mutate(
        {
          event_listeners: [
            {
              source: parsed?.source ?? "discord",
              flow_source: sanitized,
              enabled: true,
            },
          ],
        },
        { onSuccess }
      );
    }
  }

  async function onImport() {
    if (!useCode) {
      importFlowData(shareCode);
      return;
    }

    const normalized = codeInput.trim().toUpperCase();
    if (normalized.length !== 6) {
      toast.error("Enter a 6-character code");
      return;
    }
    setResolving(true);
    try {
      const data = await resolveShareCode(normalized);
      importFlowData(data);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to import");
    } finally {
      setResolving(false);
    }
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>Import {label}</DialogTitle>
        <DialogDescription>
          {useCode
            ? "Enter a share code that was exported from another app."
            : "Paste a share code (JSON) that was exported from another app."}
        </DialogDescription>
      </DialogHeader>

      {useCode ? (
        <div className="flex flex-col items-center gap-3 py-6">
          <input
            value={codeInput}
            onChange={(e) => setCodeInput(e.target.value.toUpperCase())}
            maxLength={6}
            placeholder="ABC123"
            className="w-40 rounded-md border border-input bg-background px-3 py-2 text-center text-2xl font-mono tracking-widest"
          />
          <button
            type="button"
            onClick={() => setUseCode(false)}
            className="text-sm text-muted-foreground underline underline-offset-2"
          >
            Paste JSON instead
          </button>
        </div>
      ) : (
        <>
          <Textarea
            value={shareCode}
            onChange={(e) => setShareCode(e.target.value)}
            className="min-h-[78px] max-h-[218px]"
          />
          <button
            type="button"
            onClick={() => setUseCode(true)}
            className="text-sm text-muted-foreground underline underline-offset-2 self-start"
          >
            Use a code instead
          </button>
        </>
      )}

      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <LoadingButton
          onClick={onImport}
          loading={
            commandsImportMutation.isPending ||
            eventListenersImportMutation.isPending ||
            resolving ||
            !variables ||
            !messages
          }
        >
          Import
        </LoadingButton>
      </DialogFooter>
    </>
  );
}

// Variable and message template IDs belong to the app the flow was exported
// from, so they are cleared when they don't exist in the current app.
function removeForeignReferences(
  flow: FlowData,
  variableIds: Set<string>,
  messageIds: Set<string>
) {
  let removed = 0;

  const nodes = flow.nodes.map((node) => {
    const data = { ...node.data };
    let changed = false;

    if (data.variable_id && !variableIds.has(data.variable_id)) {
      delete data.variable_id;
      changed = true;
    }
    if (data.message_template_id && !messageIds.has(data.message_template_id)) {
      delete data.message_template_id;
      changed = true;
    }

    if (changed) removed++;
    return { ...node, data };
  });

  return { flow: { ...flow, nodes }, removed };
}