import { useStore } from "@xyflow/react";
import { AlertTriangleIcon, ChevronRightIcon } from "lucide-react";
import { getNodeTitle, getNodeValues } from "@/lib/flow/nodes";
import { NodeData } from "@/lib/flow/dataSchema";
import {
  flowTestEffectMessage,
  flowTestEffectTitle,
} from "@/lib/flow/flowTest";
import {
  FlowTestEffect,
  FlowTestExecution,
  FlowTestResponse,
  FlowTestStep,
} from "@/lib/types/wire.gen";
import { useTheme } from "next-themes";
import DynamicIcon from "../icons/DynamicIcon";
import MessagePreview from "../message/MessagePreview";
import { ShowBlockButton } from "./FlowLogList";

export default function FlowTestResult({
  result,
  onShowNode,
}: {
  result: FlowTestResponse;
  onShowNode: (id: string) => void;
}) {
  return (
    <div className="flex flex-col gap-4">
      {result.executions.map((execution, i) => (
        <Execution key={i} execution={execution} onShowNode={onShowNode} />
      ))}
      <div className="text-sm text-muted-foreground">
        {result.credits === 1
          ? "A real run would cost 1 credit."
          : `A real run would cost ${result.credits} credits.`}
      </div>
    </div>
  );
}

function Execution({
  execution,
  onShowNode,
}: {
  execution: FlowTestExecution;
  onShowNode: (id: string) => void;
}) {
  return (
    <div className="flex flex-col gap-2">
      <div className="text-sm font-medium text-foreground">
        {execution.trigger}
      </div>

      {execution.steps.length === 0 && !execution.error && (
        <div className="text-sm text-muted-foreground">No blocks ran.</div>
      )}

      {execution.steps.map((step, i) => (
        <Step key={i} step={step} onShowNode={onShowNode} />
      ))}

      {execution.truncated && (
        <div className="text-sm text-muted-foreground">
          More blocks ran than are shown here.
        </div>
      )}

      {execution.notes?.map((note, i) => (
        <Note key={i} text={note} />
      ))}

      {execution.error && (
        <div className="text-sm text-destructive break-words">
          {execution.error}
        </div>
      )}
    </div>
  );
}

function Step({
  step,
  onShowNode,
}: {
  step: FlowTestStep;
  onShowNode: (id: string) => void;
}) {
  const node = useStore((s) => s.nodeLookup.get(step.node_id));
  const values = getNodeValues(step.node_type);
  const title = node
    ? getNodeTitle(node as { type?: string; data: NodeData })
    : values.defaultTitle;

  return (
    <div className="flex flex-col gap-2 bg-muted/40 rounded-sm px-3 py-2.5">
      <div className="flex items-center gap-2">
        <div
          className="rounded w-6 h-6 flex justify-center items-center flex-none"
          style={{ backgroundColor: values.color }}
        >
          <DynamicIcon
            name={values.icon as any}
            className="h-4 w-4 text-white"
          />
        </div>
        <div className="flex-auto text-sm font-medium truncate">{title}</div>
        <ShowBlockButton nodeId={step.node_id} onShowNode={onShowNode} />
      </div>

      {step.error && (
        <div className="text-sm text-destructive break-words">{step.error}</div>
      )}

      {step.notes?.map((note, i) => (
        <Note key={i} text={note} />
      ))}

      {step.effects?.map((effect, i) => (
        <Effect key={i} effect={effect} />
      ))}

      {step.result !== undefined && step.result !== null && (
        <Details summary="Result" value={step.result} />
      )}
    </div>
  );
}

function Effect({ effect }: { effect: FlowTestEffect }) {
  const { resolvedTheme } = useTheme();
  const message = flowTestEffectMessage(effect);

  return (
    <div className="flex flex-col gap-1.5">
      <div className="text-sm">{flowTestEffectTitle(effect)}</div>
      {effect.warnings?.map((warning, i) => (
        <div
          key={i}
          className="flex items-start gap-1.5 text-sm text-amber-600 dark:text-amber-400"
        >
          <AlertTriangleIcon className="size-4 flex-none mt-0.5" />
          {warning}
        </div>
      ))}
      {message ? (
        <div className="rounded-md overflow-hidden">
          <MessagePreview
            msg={message}
            lightTheme={resolvedTheme === "light"}
            reducePadding
          />
        </div>
      ) : (
        effect.data != null && <Details summary="Details" value={effect.data} />
      )}
    </div>
  );
}

function Note({ text }: { text: string }) {
  return <div className="text-sm text-muted-foreground">{text}</div>;
}

function Details({ summary, value }: { summary: string; value: unknown }) {
  return (
    <details className="group text-sm">
      <summary className="cursor-pointer list-none flex items-center gap-1 text-muted-foreground hover:text-foreground">
        <ChevronRightIcon className="size-3.5 transition-transform group-open:rotate-90" />
        {summary}
      </summary>
      <pre className="mt-1.5 max-h-64 overflow-auto rounded bg-muted p-2 text-xs whitespace-pre-wrap break-all">
        {JSON.stringify(value, null, 2)}
      </pre>
    </details>
  );
}
