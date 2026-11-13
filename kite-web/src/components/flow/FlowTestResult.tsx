import { createContext, useContext, useMemo, useState } from "react";
import { useStore } from "@xyflow/react";
import { AlertTriangleIcon, ChevronRightIcon } from "lucide-react";
import { getNodeTitle, getNodeValues } from "@/lib/flow/nodes";
import { NodeData } from "@/lib/flow/dataSchema";
import {
  canSetTestResult,
  flowTestEffectMessage,
  flowTestEffectTitle,
  flowTestNoteText,
} from "@/lib/flow/flowTest";
import {
  FlowTestAction,
  FlowTestEffect,
  FlowTestMock,
  FlowTestExecution,
  FlowTestResponse,
  FlowTestStep,
} from "@/lib/types/wire.gen";
import { useTheme } from "next-themes";
import DynamicIcon from "../icons/DynamicIcon";
import MessagePreview from "../message/MessagePreview";
import { PreviewComponentActionsContext } from "../message/MessagePreviewComponents";
import { MessageComponentSelectMenu } from "@/lib/message/schema";
import { ShowBlockButton } from "./FlowLogList";
import FlowTestModalForm from "./FlowTestModalForm";
import { MockTarget } from "./FlowTestMockDialog";
import { Button } from "../ui/button";
import FlowTestSelectDialog from "./FlowTestSelectDialog";

// What the user does in the result, like clicking a button the flow sent,
// runs the test again with it. isLast is set within the last run, the only
// one whose modal can still be submitted.
const TestActionContext = createContext<{
  onAction(action: FlowTestAction): void;
  onEditMock(target: MockTarget): void;
  mocks: Record<string, FlowTestMock>;
  pending: boolean;
  isLast: boolean;
}>({
  onAction: () => {},
  onEditMock: () => {},
  mocks: {},
  pending: false,
  isLast: false,
});

export default function FlowTestResult({
  result,
  pending,
  mocks,
  onAction,
  onEditMock,
  onShowNode,
}: {
  result: FlowTestResponse;
  pending: boolean;
  mocks: Record<string, FlowTestMock>;
  onAction: (action: FlowTestAction) => void;
  onEditMock: (target: MockTarget) => void;
  onShowNode: (id: string) => void;
}) {
  const [selectMenu, setSelectMenu] =
    useState<MessageComponentSelectMenu | null>(null);

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
    <PreviewComponentActionsContext.Provider value={componentActions}>
      <div className="flex flex-col gap-4">
        {result.executions.map((execution, i) => (
          <TestActionContext.Provider
            key={i}
            value={{
              onAction,
              onEditMock,
              mocks,
              pending,
              isLast: i === last,
            }}
          >
            <Execution execution={execution} onShowNode={onShowNode} />
          </TestActionContext.Provider>
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
  const { onEditMock, mocks } = useContext(TestActionContext);
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
        <Note key={i} text={flowTestNoteText(note)} />
      ))}

      {node && (canSetTestResult(step) || mocks[step.node_id]) && (
        <Button
          variant="secondary"
          size="sm"
          className="self-start"
          onClick={() =>
            onEditMock({
              nodeId: step.node_id,
              nodeType: step.node_type,
              title,
            })
          }
        >
          {mocks[step.node_id] ? "Change test result" : "Set test result"}
        </Button>
      )}

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
  const { onAction, pending, isLast } = useContext(TestActionContext);
  const message = flowTestEffectMessage(effect);
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
