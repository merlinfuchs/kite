import { useState } from "react";
import { Node, useStore, useStoreApi } from "@xyflow/react";
import { XIcon } from "lucide-react";
import { NodeData } from "@/lib/flow/dataSchema";
import {
  blockReturnsText,
  jsonOrText,
  mockTestResult,
} from "@/lib/flow/flowTest";
import { getNodeTitle } from "@/lib/flow/nodes";
import { FlowTestMock } from "@/lib/types/wire.gen";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Textarea } from "../ui/textarea";
import { RadioGroup, RadioGroupItem } from "../ui/radio-group";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import { FlowTestField as Field } from "./FlowTestField";

// The block whose test result is edited.
export type MockTarget = { nodeId: string; nodeType: string; title: string };

// Sets what a block does in tests instead of running, like the answer of AI
// Check or the response of a request.
export default function FlowTestMockDialog({
  target,
  mock,
  onClose,
  onSave,
}: {
  target: MockTarget | null;
  mock: FlowTestMock | undefined;
  onClose: () => void;
  // Undefined removes the test result.
  onSave: (mock: FlowTestMock | undefined) => void;
}) {
  return (
    <Dialog open={!!target} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-md">
        {target && (
          <MockForm
            key={target.nodeId}
            target={target}
            mock={mock}
            onSave={onSave}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

function MockForm({
  target,
  mock,
  onSave,
}: {
  target: MockTarget;
  mock: FlowTestMock | undefined;
  onSave: (mock: FlowTestMock | undefined) => void;
}) {
  const kind = mockKind(target.nodeType);
  const [fails, setFails] = useState(!!mock?.error);
  const [error, setError] = useState(mock?.error ?? "");
  const [value, setValue] = useState(() => initialValue(kind, mock?.result));
  const [statusCode, setStatusCode] = useState(
    String(mock?.result?.status_code ?? 200)
  );
  const choices = useAIChoiceOptions(target);

  function result(): unknown {
    switch (kind) {
      case "http":
        return {
          status_code: Number(statusCode) || 200,
          body: jsonOrText(value),
        };
      case "text":
        return mockTestResult(target.nodeType, value);
      default:
        return value;
    }
  }

  const canSave = fails ? !!error.trim() : !!value.trim() || kind === "http";

  return (
    <>
      <DialogHeader>
        <DialogTitle>Test result of {target.title}</DialogTitle>
        <DialogDescription>
          In tests, the block does this instead of running. The test runs again
          with it.
        </DialogDescription>
      </DialogHeader>

      <RadioGroup
        value={fails ? "fails" : "works"}
        onValueChange={(v) => setFails(v === "fails")}
        className="flex gap-4"
      >
        <label className="flex items-center gap-2 text-sm">
          <RadioGroupItem value="works" />
          {kind === "text" || kind === "http" ? "Returns" : "Answers"}
        </label>
        <label className="flex items-center gap-2 text-sm">
          <RadioGroupItem value="fails" />
          Fails
        </label>
      </RadioGroup>

      {fails ? (
        <Field label="Error">
          <Input value={error} onChange={(e) => setError(e.target.value)} />
        </Field>
      ) : kind === "check" || kind === "verify" ? (
        <RadioGroup value={value} onValueChange={setValue} className="gap-2">
          {(kind === "check"
            ? [
                ["yes", "Yes"],
                ["no", "No"],
              ]
            : [
                ["valid", "Valid"],
                ["invalid", "Invalid"],
              ]
          ).map(([v, label]) => (
            <label key={v} className="flex items-center gap-2 text-sm">
              <RadioGroupItem value={v} />
              {label}
            </label>
          ))}
        </RadioGroup>
      ) : kind === "choice" && choices.length > 0 ? (
        <RadioGroup value={value} onValueChange={setValue} className="gap-2">
          {choices.map((name) => (
            <label key={name} className="flex items-center gap-2 text-sm">
              <RadioGroupItem value={name} />
              {name}
            </label>
          ))}
        </RadioGroup>
      ) : kind === "http" ? (
        <>
          <Field label="Status code">
            <Input
              type="number"
              value={statusCode}
              onChange={(e) => setStatusCode(e.target.value)}
            />
          </Field>
          <Field label="Body">
            <Textarea
              value={value}
              onChange={(e) => setValue(e.target.value)}
              className="font-mono"
              placeholder='{"name": "Kite"}'
            />
          </Field>
        </>
      ) : (
        <Field label="Result">
          <Textarea
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder={
              blockReturnsText(target.nodeType)
                ? "The text the block returns"
                : "The result of the block as JSON"
            }
          />
        </Field>
      )}

      <DialogFooter className="gap-2">
        {mock && (
          <Button variant="ghost" onClick={() => onSave(undefined)}>
            Remove
          </Button>
        )}
        <Button
          disabled={!canSave}
          onClick={() =>
            onSave(fails ? { error: error.trim() } : { result: result() })
          }
        >
          Save and run again
        </Button>
      </DialogFooter>
    </>
  );
}

type MockKind = "check" | "choice" | "verify" | "http" | "text";

function mockKind(nodeType: string): MockKind {
  switch (nodeType) {
    case "control_ai_check":
      return "check";
    case "control_condition_ai_choice":
      return "choice";
    case "control_webhook_verify":
      return "verify";
    case "action_http_request":
      return "http";
    default:
      return "text";
  }
}

function initialValue(kind: MockKind, result: unknown): string {
  if (result === undefined) {
    return kind === "check" ? "yes" : kind === "verify" ? "invalid" : "";
  }
  if (kind === "http") {
    const body = (result as { body?: unknown })?.body;
    return body === undefined
      ? ""
      : typeof body === "string"
      ? body
      : JSON.stringify(body, null, 2);
  }
  return typeof result === "string" ? result : JSON.stringify(result, null, 2);
}

// The names of the options of an AI Choice block, which its answer is one of.
function useAIChoiceOptions(target: MockTarget) {
  const store = useStoreApi<Node<NodeData>>();
  if (target.nodeType !== "control_condition_ai_choice") return [];

  const { nodeLookup, edges } = store.getState();
  return edges
    .filter((e) => e.source === target.nodeId)
    .map((e) => nodeLookup.get(e.target))
    .flatMap((n) =>
      n?.type === "control_condition_item_ai_choice" && n.data.ai_choice_name
        ? [n.data.ai_choice_name]
        : []
    );
}

// The test results that are set, which later tests use too.
export function FlowTestMockList({
  mocks,
  onEdit,
  onRemove,
}: {
  mocks: Record<string, FlowTestMock>;
  onEdit: (target: MockTarget) => void;
  onRemove: (nodeId: string) => void;
}) {
  const ids = Object.keys(mocks);
  if (!ids.length) return null;

  return (
    <div className="flex flex-col gap-1.5">
      <div className="text-sm font-medium">Test results</div>
      {ids.map((nodeId) => (
        <MockListEntry
          key={nodeId}
          nodeId={nodeId}
          fails={!!mocks[nodeId].error}
          onEdit={onEdit}
          onRemove={onRemove}
        />
      ))}
    </div>
  );
}

function MockListEntry({
  nodeId,
  fails,
  onEdit,
  onRemove,
}: {
  nodeId: string;
  fails: boolean;
  onEdit: (target: MockTarget) => void;
  onRemove: (nodeId: string) => void;
}) {
  const node = useStore((s) => s.nodeLookup.get(nodeId)) as
    | Node<NodeData>
    | undefined;
  // Blocks deleted since aren't in the test anymore.
  if (!node) return null;
  const title = getNodeTitle(node);

  return (
    <div className="flex items-center gap-2 text-sm">
      <button
        className="flex-auto text-left truncate hover:underline"
        onClick={() => onEdit({ nodeId, nodeType: node.type!, title })}
      >
        {title}: {fails ? "fails" : "returns your result"}
      </button>
      <Button
        variant="ghost"
        size="icon"
        className="flex-none size-7"
        onClick={() => onRemove(nodeId)}
      >
        <XIcon className="size-4" />
        <span className="sr-only">Remove</span>
      </Button>
    </div>
  );
}
