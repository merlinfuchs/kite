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
import { toast } from "sonner";

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
  const [shareCode, setShareCode] = useState("");

  const router = useRouter();
  const appId = useAppId();
  const variables = useVariables();
  const messages = useMessages();

  const commandsImportMutation = useCommandsImportMutation(appId);
  const eventListenersImportMutation = useEventListenersImportMutation(appId);

  const { label, entryNodeType, href } = kinds[kind];

  function onImport() {
    let parsed: { flow_source?: FlowData; source?: string } | undefined;
    try {
      parsed = JSON.parse(shareCode);
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

  return (
    <>
      <DialogHeader>
        <DialogTitle>Import {label}</DialogTitle>
        <DialogDescription>
          Paste a share code that was exported from another app.
        </DialogDescription>
      </DialogHeader>
      <Textarea
        value={shareCode}
        onChange={(e) => setShareCode(e.target.value)}
        className="min-h-[78px] max-h-[218px]"
      />
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <LoadingButton
          onClick={onImport}
          loading={
            commandsImportMutation.isPending ||
            eventListenersImportMutation.isPending ||
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
