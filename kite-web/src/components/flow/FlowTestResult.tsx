import { createContext, memo, useContext, useMemo, useState } from "react";
import { useStore } from "@xyflow/react";
import { AlertTriangleIcon, ChevronRightIcon } from "lucide-react";
import { getNodeTitle, getNodeValues } from "@/lib/flow/nodes";
import { NodeData } from "@/lib/flow/dataSchema";
import {
  flowTestEffectMessage,
  flowTestEffectTitle,
} from "@/lib/flow/flowTest";
import {
  FlowTestAction,
  FlowTestEffect,
  FlowTestExecution,
  FlowTestResponse,
  FlowTestStep,
} from "@/lib/types/wire.gen";
import DynamicIcon from "../icons/DynamicIcon";
import MessagePreviewBox from "../message/MessagePreviewBox";
import { PreviewComponentActionsContext } from "../message/MessagePreviewComponents";
import { MessageComponentSelectMenu } from "@/lib/message/schema";
import { ShowBlockButton } from "./FlowLogList";
import FlowTestModalForm from "./FlowTestModalForm";
import FlowTestSelectDialog from "./FlowTestSelectDialog";

type Actions = {
  // What the user does in the result, like clicking a button the flow sent,
  // runs the test again with it.
  onAction(action: FlowTestAction): void;
  onShowNode(id: string): void;
  pending: boolean;
};

const ActionsContext = createContext<Actions | null>(null);

// Memoized, as the panel renders on every change of its form.
export default memo(function FlowTestResult({
  result,
  onAction,
  onShowNode,
  pending,
}: { result: FlowTestResponse } & Actions) {
  const [selectMenu, setSelectMenu] =
    useState<MessageComponentSelectMenu | null>(null);

  const actions = useMemo(
    () => ({ onAction, onShowNode, pending }),
    [onAction, onShowNode, pending]
  );
  const componentActions = useMemo(
    () => ({
      onClick: (button: { flow_source_id: string }) =>
        onAction({ click: button.flow_source_id }),
      onSelect: setSelectMenu,
      disabled: pending,
    }),
    [onAction, pending]
  );

  const last = result.executions.length - 1;

  return (
    <ActionsContext.Provider value={actions}>
      <PreviewComponentActionsContext.Provider value={componentActions}>
        <div className="flex flex-col gap-4">
          {result.executions.map((execution, i) => (
            <Execution key={i} execution={execution} isLast={i === last} />
          ))}
          <div className="text-sm text-muted-foreground">
            {result.credits === 1
              ? "A real run would cost 1 credit."
              : `A real run would cost ${result.credits} credits.`}
          </div>
        </div>
        <FlowTestSelectDialog
          selectMenu={selectMenu}
          onClose={() => setSelectMenu(null)}
          onSelect={(values) => {
            if (!selectMenu) return;
            onAction({ select: selectMenu.flow_source_id, values });
            setSelectMenu(null);
          }}
        />
      </PreviewComponentActionsContext.Provider>
    </ActionsContext.Provider>
  );
});

function useActions() {
  return useContext(ActionsContext)!;
}

// isLast is set for the last run, whose modal is the one that's open.
function Execution({
  execution,
  isLast,
}: {
  execution: FlowTestExecution;
  isLast: boolean;
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
        <Step
          key={i}
          step={step}
          isLast={isLast}
          inFlow={!execution.other_flow}
        />
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

// inFlow is unset for steps of other flows, like the flow of a button on a
// message template, whose blocks can have the IDs of blocks in this one.
function Step({
  step,
  isLast,
  inFlow,
}: {
  step: FlowTestStep;
  isLast: boolean;
  inFlow: boolean;
}) {
  const { onShowNode } = useActions();
  const node = useStore((s) =>
    inFlow ? s.nodeLookup.get(step.node_id) : undefined
  );
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
        {node && (
          <ShowBlockButton nodeId={step.node_id} onShowNode={onShowNode} />
        )}
      </div>

      {step.error && (
        <div className="text-sm text-destructive break-words">{step.error}</div>
      )}

      {step.notes?.map((note, i) => (
        <Note key={i} text={note} />
      ))}

      {step.effects?.map((effect, i) => (
        <Effect key={i} effect={effect} isLast={isLast} />
      ))}

      {step.result !== undefined && step.result !== null && (
        <Details summary="Result" value={step.result} />
      )}
    </div>
  );
}

function Effect({
  effect,
  isLast,
}: {
  effect: FlowTestEffect;
  isLast: boolean;
}) {
  const { onAction, pending } = useActions();
  const message = useMemo(() => flowTestEffectMessage(effect), [effect]);
  const modal =
    effect.action === "respond" && effect.data?.type === "modal"
      ? effect.data.modal
      : null;

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
      {modal && isLast ? (
        <FlowTestModalForm
          modal={modal}
          pending={pending}
          onSubmit={(values) => onAction({ submit: values })}
        />
      ) : message ? (
        <MessagePreviewBox msg={message} />
      ) : (
        effect.data != null && <Details summary="Details" value={effect.data} />
      )}
    </div>
  );
}

function Note({ text }: { text: string }) {
  return <div className="text-sm text-muted-foreground">{text}</div>;
}

// Results can be large, so they're only formatted when opened.
function Details({ summary, value }: { summary: string; value: unknown }) {
  const [open, setOpen] = useState(false);

  return (
    <details
      className="group text-sm"
      onToggle={(e) => setOpen(e.currentTarget.open)}
    >
      <summary className="cursor-pointer list-none flex items-center gap-1 text-muted-foreground hover:text-foreground">
        <ChevronRightIcon className="size-3.5 transition-transform group-open:rotate-90" />
        {summary}
      </summary>
      {open && (
        <pre className="mt-1.5 max-h-64 overflow-auto rounded bg-muted p-2 text-xs whitespace-pre-wrap break-all">
          {JSON.stringify(value, null, 2)}
        </pre>
      )}
    </details>
  );
}
