import { ReactNode, useState } from "react";
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
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { ShareCodeDisplay, ShareCodePanel } from "./ShareCode";
import { useShareCodeCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useFunctions, useMessages } from "@/lib/hooks/api";
import { collectReferencedResources } from "@/lib/flow/messageTemplates";
import { FlowData } from "@/lib/types/flow.gen";
import {
  BracesIcon,
  CopyIcon,
  KeyRoundIcon,
  SquareFunctionIcon,
} from "lucide-react";
import { getFunctionName } from "@/lib/flow/functions";
import { toast } from "sonner";

type ExportProps = {
  title: string;
  type: "command" | "event_listener" | "message" | "function";
  shareData: Record<string, unknown>;
  // Message templates and functions used by these flows are exported along
  // with them, so they don't have to be recreated in the other app.
  flows: (FlowData | null | undefined)[];
  // Templates and functions that are already part of shareData.
  excludeMessageIds?: string[];
  excludeFunctionIds?: string[];
};

export default function FlowExportDialog({
  children,
  ...props
}: ExportProps & { children: ReactNode }) {
  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-lg">
        <ExportForm {...props} />
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so its state resets every time it's opened.
function ExportForm({
  title,
  type,
  shareData: baseShareData,
  flows,
  excludeMessageIds,
  excludeFunctionIds,
}: ExportProps) {
  const [showJson, setShowJson] = useState(false);
  const [code, setCode] = useState<string | null>(null);
  // Whether to export the functions the flows call, asked before anything is
  // exported if they call any.
  const [includeFunctions, setIncludeFunctions] = useState<boolean | null>(
    null
  );

  const appId = useAppId();
  const createMutation = useShareCodeCreateMutation(appId);
  const messages = useMessages();
  const functions = useFunctions();

  const referenced =
    messages && functions
      ? collectReferencedResources(
          flows,
          messages,
          functions,
          excludeMessageIds,
          excludeFunctionIds
        )
      : null;
  // Without the functions, the templates only they use are left out too.
  const bundled =
    referenced && includeFunctions === false
      ? collectReferencedResources(flows, messages!, [], excludeMessageIds)
      : referenced;
  // Without the templates and functions the export would be incomplete, so
  // it's only offered once they're loaded.
  const shareData = {
    ...baseShareData,
    ...(bundled && bundled.messages.length > 0
      ? { messages: bundled.messages }
      : {}),
    ...(bundled && bundled.functions.length > 0
      ? { functions: bundled.functions }
      : {}),
  };

  const bundledParts = [
    bundled && bundled.messages.length > 0
      ? `${bundled.messages.length} message template(s)`
      : null,
    bundled && bundled.functions.length > 0
      ? `${bundled.functions.length} function(s)`
      : null,
  ].filter(Boolean);
  const bundledNote =
    bundledParts.length > 0
      ? ` The ${bundledParts.join(" and ")} it uses are included.`
      : "";

  function copy(text: string) {
    navigator.clipboard
      .writeText(text)
      .then(() => toast.success("Copied to clipboard!"))
      .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
  }

  function generateCode() {
    createMutation.mutate(
      { type, data: shareData },
      {
        onSuccess: (res) => {
          if (res.success) {
            setCode(res.data.code);
          } else {
            toast.error(
              `Failed to generate share code: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  if (
    referenced &&
    referenced.functions.length > 0 &&
    includeFunctions === null
  ) {
    const count = referenced.functions.length;
    return (
      <>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>
            This uses {count === 1 ? "a function" : `${count} functions`}.
            Export {count === 1 ? "it" : "them"} too, so whoever imports this
            gets {count === 1 ? "it" : "them"} already connected?
          </DialogDescription>
        </DialogHeader>
        <ul className="space-y-2">
          {referenced.functions.map((f) => (
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
          Without them, the Call function blocks have no function selected after
          importing, unless the app already has the same functions.
        </p>
        <DialogFooter className="gap-2 sm:space-x-0">
          <Button variant="outline" onClick={() => setIncludeFunctions(false)}>
            Export without
          </Button>
          <Button onClick={() => setIncludeFunctions(true)}>
            Export with functions
          </Button>
        </DialogFooter>
      </>
    );
  }

  if (showJson) {
    const json = bundled ? JSON.stringify(shareData) : "";

    return (
      <>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>
            Copy the JSON below and import it into another app.{bundledNote}
          </DialogDescription>
        </DialogHeader>
        <Textarea
          readOnly
          value={json}
          placeholder="Loading..."
          minRows={8}
          maxRows={8}
          className="resize-none break-all font-mono md:text-xs"
          onFocus={(e) => e.target.select()}
        />
        <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
          <Button variant="ghost" onClick={() => setShowJson(false)}>
            <KeyRoundIcon className="mr-2 h-4 w-4" />
            Use share code
          </Button>
          <Button onClick={() => copy(json)} disabled={!bundled}>
            <CopyIcon className="mr-2 h-4 w-4" />
            Copy JSON
          </Button>
        </DialogFooter>
      </>
    );
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>
          Anyone with the code can import this into their app. Codes expire
          after 90 days without use.{bundledNote}
        </DialogDescription>
      </DialogHeader>
      <ShareCodePanel>
        <ShareCodeDisplay code={code} />
      </ShareCodePanel>
      <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
        <Button variant="ghost" onClick={() => setShowJson(true)}>
          <BracesIcon className="mr-2 h-4 w-4" />
          Use JSON
        </Button>
        {code ? (
          <Button onClick={() => copy(code)}>
            <CopyIcon className="mr-2 h-4 w-4" />
            Copy code
          </Button>
        ) : (
          <LoadingButton
            onClick={generateCode}
            loading={createMutation.isPending || !bundled}
          >
            Generate code
          </LoadingButton>
        )}
      </DialogFooter>
    </>
  );
}
