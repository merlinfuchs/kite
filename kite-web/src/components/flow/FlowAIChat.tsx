import { track } from "@/lib/analytics";
import { useFlowAIChatMutation } from "@/lib/api/mutations";
import {
  diffFlow,
  FlowChanges,
  hasFlowChanges,
  runFlowAIPrompt,
} from "@/lib/flow/ai";
import { useFlowContext } from "@/lib/flow/context";
import { NodeType } from "@/lib/flow/dataSchema";
import { FlowSnapshot } from "@/lib/flow/history";
import { getNodeTitle } from "@/lib/flow/nodes";
import { useBillingPlans, useFlowAIUsage } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import {
  FlowAIChatMessage,
  FlowAIField,
  FlowAIUsage,
} from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { useReactFlow, useStoreApi } from "@xyflow/react";
import {
  CheckIcon,
  CrownIcon,
  LoaderCircleIcon,
  MinusIcon,
  PencilIcon,
  PlusIcon,
  SendHorizontalIcon,
  SparklesIcon,
  SquarePenIcon,
  Undo2Icon,
  XIcon,
} from "lucide-react";
import {
  memo,
  RefObject,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import Link from "next/link";
import { Button } from "../ui/button";
import { Textarea } from "../ui/textarea";
import FlowAIFieldsCard from "./FlowAIFieldsCard";
import FlowAIMarkdown from "./FlowAIMarkdown";
import { FlowEditorApi } from "./FlowEditor";

interface ChatEntry extends FlowAIChatMessage {
  // A request the user can send to make the change the answer suggests.
  buildPrompt?: string;
  // What the AI asks the user to fill in.
  fields?: FlowAIField[];
  // Problems left with the AI's changes.
  issues?: string[];
  repaired?: boolean;
  // The prompt failed, and content is the error.
  failed?: boolean;
  change?: AIChange;
}

// What an answer changed in the flow. The changes stay pending, and their
// blocks highlighted, until the user keeps or undoes them. Anything else that
// changes the flow keeps them, so undoing can't lose the user's own edits.
interface AIChange extends FlowChanges {
  before: FlowSnapshot;
  after: FlowSnapshot;
  status: "pending" | "kept" | "undone";
}

// How the user kept or undid the changes. "editor" is its undo button or
// shortcut, "edit" any other change to the flow.
type ResolveReason =
  | "button"
  | "prompt"
  | "new_chat"
  | "save"
  | "edit"
  | "editor";

export default memo(function FlowAIChat({
  editorRef,
  saveCount,
  onClose,
}: {
  editorRef: RefObject<FlowEditorApi>;
  // Goes up when the flow is saved, which keeps the pending changes.
  saveCount?: number;
  onClose: () => void;
}) {
  const context = useFlowContext((c) => c.type);
  const setAIChangedNodeIds = useFlowContext((c) => c.setAIChangedNodeIds);
  const { getNodes, getEdges, fitView } = useReactFlow<NodeType>();
  const store = useStoreApi<NodeType>();
  const appId = useAppId();
  const { mutateAsync: sendChat } = useFlowAIChatMutation(appId);
  const usage = useFlowAIUsage();

  const [entries, setEntries] = useState<ChatEntry[]>([]);
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);

  // Stops a running prompt when the editor is closed.
  // Created in the effect, as React may unmount and mount it again.
  const abort = useRef<AbortController>();
  useEffect(() => {
    const controller = new AbortController();
    abort.current = controller;
    return () => controller.abort();
  }, []);

  // "Build this" puts the suggested request into the input, so the user can
  // fill in what only they know before it uses a prompt.
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const fillInput = useCallback((prompt: string) => {
    setInput(prompt);
    inputRef.current?.focus();
    track("flow_ai_build_clicked");
  }, []);

  // The editor may show the changes a moment after they're applied, so edits
  // only count once it has.
  const pending = useRef<{ change: AIChange; shown: boolean } | null>(null);

  const resolve = useCallback(
    (status: "kept" | "undone", reason: ResolveReason) => {
      const change = pending.current?.change;
      if (!change) return;
      pending.current = null;
      setEntries((e) =>
        e.map((entry) =>
          entry.change === change
            ? { ...entry, change: { ...change, status } }
            : entry
        )
      );
      setAIChangedNodeIds([]);
      track(
        status === "kept" ? "flow_ai_changes_kept" : "flow_ai_changes_undone",
        {
          context,
          reason,
          added_blocks: change.added.length,
          changed_blocks: change.changed.length,
          removed_blocks: change.removed.length,
        }
      );
      return change;
    },
    [context, setAIChangedNodeIds]
  );

  const keep = useCallback(() => resolve("kept", "button"), [resolve]);
  const undo = useCallback(() => {
    const change = resolve("undone", "button");
    if (change) {
      editorRef.current?.replaceFlow(change.before.nodes, change.before.edges);
    }
  }, [resolve, editorRef]);

  // Undoing in the editor undoes the changes, any other edit keeps them.
  useEffect(() => {
    let last: FlowSnapshot | undefined;
    return store.subscribe(({ nodes, edges }) => {
      const p = pending.current;
      if (!p || (nodes === last?.nodes && edges === last?.edges)) return;
      last = { nodes, edges };
      if (!hasFlowChanges(diffFlow(p.change.after, last))) {
        p.shown = true;
      } else if (p.shown) {
        if (hasFlowChanges(diffFlow(p.change.before, last))) {
          resolve("kept", "edit");
        } else {
          resolve("undone", "editor");
        }
      }
    });
  }, [store, resolve]);

  const lastSaveCount = useRef(saveCount);
  useEffect(() => {
    if (saveCount === lastSaveCount.current) return;
    lastSaveCount.current = saveCount;
    resolve("kept", "save");
  }, [saveCount, resolve]);

  const showBlock = useCallback(
    (id: string) => fitView({ nodes: [{ id }], duration: 300, maxZoom: 1 }),
    [fitView]
  );

  const bottomRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [entries, busy]);

  const send = useCallback(
    async (content: string) => {
      const messages: FlowAIChatMessage[] = [
        ...entries
          .filter((e) => !e.failed)
          .map(({ role, content }) => ({ role, content })),
        { role: "user", content },
      ];
      resolve("kept", "prompt");
      setEntries((e) => [...e, { role: "user", content }]);
      setBusy(true);
      track("flow_ai_prompt_sent", {
        context,
        turn: messages.filter((m) => m.role === "user").length,
      });
      setAIChangedNodeIds([]);

      // The rounds of the prompt are undone together.
      const mergeKey = `ai:${Date.now()}`;
      let before: FlowSnapshot | undefined;
      let applied: FlowSnapshot | undefined;
      try {
        const res = await runFlowAIPrompt({
          context,
          messages,
          getFlow: () => ({ nodes: getNodes(), edges: getEdges() }),
          applyFlow: ({ nodes, edges }, changedNodeIds) => {
            before ??= { nodes: getNodes(), edges: getEdges() };
            applied = { nodes, edges };
            editorRef.current?.replaceFlow(nodes, edges, mergeKey);
            setAIChangedNodeIds(changedNodeIds);
          },
          send: (req, signal) => sendChat({ req, signal }),
          signal: abort.current?.signal,
        });
        track("flow_ai_prompt_answered", {
          context,
          changed_blocks: res.changedNodeIds.length,
          repairs: res.repairs,
          issues: res.issues.length,
          asked_fields: res.fields?.length ?? 0,
          suggested_build: !!res.buildPrompt,
        });
        let change: AIChange | undefined;
        if (before && applied) {
          const changes = diffFlow(before, applied);
          for (const node of changes.added) {
            track("block_added", { type: node.type!, source: "ai" });
          }
          if (hasFlowChanges(changes)) {
            change = { ...changes, before, after: applied, status: "pending" };
            const { nodes, edges } = store.getState();
            pending.current = {
              change,
              shown: !hasFlowChanges(diffFlow(applied, { nodes, edges })),
            };
          }
        }
        if (res.changedNodeIds.length > 0) {
          // Once the new blocks have been measured.
          setTimeout(() => {
            fitView({
              nodes: res.changedNodeIds.map((id) => ({ id })),
              duration: 300,
              maxZoom: 1,
            });
          }, 50);
        }
        setEntries((e) => [
          ...e,
          {
            role: "assistant",
            content: res.message,
            buildPrompt: res.buildPrompt,
            fields: res.fields,
            issues: res.issues,
            repaired: res.repairs > 0 && res.issues.length === 0,
            change,
          },
        ]);
      } catch (err) {
        // Not when the prompt was stopped by closing the editor.
        if (!abort.current?.signal.aborted) {
          track("flow_ai_prompt_failed", { context });
        }
        setEntries((e) => [
          ...e,
          { role: "assistant", content: (err as Error).message, failed: true },
        ]);
      } finally {
        setBusy(false);
      }
    },
    [
      entries,
      context,
      getNodes,
      getEdges,
      fitView,
      editorRef,
      sendChat,
      setAIChangedNodeIds,
      resolve,
      store,
    ]
  );

  const submit = useCallback(() => {
    const content = input.trim();
    if (!content || busy) return;
    setInput("");
    send(content);
  }, [input, busy, send]);

  const limit = usage?.prompts_limit;
  const left = usage ? Math.max(limit! - usage.prompts_used, 0) : undefined;
  const exhausted =
    !!limit && (left === 0 || usage!.answers_used >= usage!.answers_limit);
  const unavailable = limit === 0 || exhausted;

  return (
    <div className="flex-none w-full md:w-96 h-full flex flex-col bg-background md:bg-muted/30 md:border-l">
      {/* Everything is on the left, where it doesn't cover the close button of
          dialogs the editor is shown in. */}
      <div className="flex-none flex items-center gap-2 px-4 h-12">
        <div className="flex items-center gap-2 font-medium">
          <SparklesIcon className="size-4" />
          Flow AI
        </div>
        <div className="flex items-center">
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            title="New chat"
            disabled={busy || entries.length === 0}
            onClick={() => {
              resolve("kept", "new_chat");
              setEntries([]);
            }}
          >
            <SquarePenIcon className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-8"
            title="Close"
            onClick={onClose}
          >
            <XIcon className="size-4" />
          </Button>
        </div>
      </div>

      <div className="flex-auto overflow-y-auto px-4 py-2 space-y-3 text-sm">
        {entries.length === 0 && (
          <div className="text-muted-foreground space-y-2">
            <p>
              Describe what the flow should do, like &quot;Ban the user from the
              argument and log why&quot;. Select blocks to ask about them.
            </p>
            <p>
              Changes show up in the flow right away, for you to keep or undo.
              Nothing is saved until you save.
            </p>
            <p>
              Each message that changes the flow uses one prompt. Answers
              without changes, like questions the AI asks back, and fixes of its
              own changes are free. Start a new chat for unrelated changes.
            </p>
          </div>
        )}
        {entries.map((entry, i) => (
          <ChatBubble
            key={i}
            entry={entry}
            canBuild={!unavailable}
            onBuild={fillInput}
            onKeep={keep}
            onUndo={undo}
            onShowBlock={showBlock}
          />
        ))}
        {!busy && !!entries.at(-1)?.fields?.length && (
          <FlowAIFieldsCard
            key={entries.length}
            fields={entries.at(-1)!.fields!}
            disabled={unavailable}
            onSend={send}
          />
        )}
        {busy && (
          <div className="flex items-center gap-2 text-muted-foreground">
            <LoaderCircleIcon className="size-4 animate-spin" />
            Working on it...
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <div className="flex-none p-4 space-y-2">
        {exhausted ? (
          <LimitNotice usage={usage!} />
        ) : (
          <div className="relative">
            <Textarea
              ref={inputRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  submit();
                }
              }}
              placeholder={
                limit === 0
                  ? "Your plan doesn't include the flow AI."
                  : "Ask a question or for a change..."
              }
              maxLength={4000}
              minRows={2}
              maxRows={8}
              disabled={unavailable}
              className="pr-12 resize-none"
            />
            <Button
              size="icon"
              className="absolute bottom-2 right-2 size-8"
              title="Send"
              disabled={busy || unavailable || !input.trim()}
              onClick={submit}
            >
              <SendHorizontalIcon className="size-4" />
            </Button>
          </div>
        )}
        {!!limit && (
          <div
            className={cn(
              "text-xs text-muted-foreground",
              left! <= 3 && "text-amber-600 dark:text-amber-500"
            )}
          >
            {left} of {limit} prompts left this month
          </div>
        )}
      </div>
    </div>
  );
});

// Memoized, as the chat re-renders on every keystroke in the input.
const ChatBubble = memo(function ChatBubble({
  entry,
  canBuild,
  onBuild,
  onKeep,
  onUndo,
  onShowBlock,
}: {
  entry: ChatEntry;
  canBuild: boolean;
  onBuild: (prompt: string) => void;
  onKeep: () => void;
  onUndo: () => void;
  onShowBlock: (id: string) => void;
}) {
  if (entry.role === "user") {
    return (
      <div className="ml-8 rounded-lg bg-primary/15 px-3 py-2 whitespace-pre-wrap">
        {entry.content}
      </div>
    );
  }

  return (
    <div className={cn("mr-8 space-y-2", entry.failed && "text-destructive")}>
      {entry.failed ? (
        <p className="whitespace-pre-wrap">{entry.content}</p>
      ) : (
        entry.content && <FlowAIMarkdown text={entry.content} />
      )}
      {entry.buildPrompt && (
        <Button
          size="sm"
          variant="outline"
          className="gap-2"
          title={entry.buildPrompt}
          disabled={!canBuild}
          onClick={() => onBuild(entry.buildPrompt!)}
        >
          <SparklesIcon className="size-4" />
          Build this
        </Button>
      )}
      {entry.change && (
        <ChangeCard
          change={entry.change}
          onKeep={onKeep}
          onUndo={onUndo}
          onShowBlock={onShowBlock}
        />
      )}
      {entry.repaired && (
        <p className="text-xs text-muted-foreground">
          Fixed problems with its changes.
        </p>
      )}
      {!!entry.issues?.length && (
        <div className="text-xs text-muted-foreground">
          <p>Some problems are left:</p>
          <ul className="list-disc pl-4">
            {entry.issues.map((issue, i) => (
              <li key={i}>{issue}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
});

const maxChangeRows = 8;

function ChangeCard({
  change,
  onKeep,
  onUndo,
  onShowBlock,
}: {
  change: AIChange;
  onKeep: () => void;
  onUndo: () => void;
  onShowBlock: (id: string) => void;
}) {
  const rows = [
    ...change.added.map((node) => ({ node, kind: "added" as const })),
    ...change.changed.map((node) => ({ node, kind: "changed" as const })),
    ...change.removed.map((node) => ({ node, kind: "removed" as const })),
  ];
  // Blocks of an undone change aren't in the flow anymore.
  const canShow = change.status !== "undone";

  return (
    <div className="rounded-lg border bg-background text-xs">
      <div className="p-2 space-y-0.5">
        {rows.slice(0, maxChangeRows).map(({ node, kind }) => {
          const Icon = {
            added: PlusIcon,
            changed: PencilIcon,
            removed: MinusIcon,
          }[kind];
          return (
            <button
              key={node.id}
              type="button"
              className="flex w-full items-center gap-2 rounded px-1.5 py-1 text-left enabled:hover:bg-muted"
              title={kind === "removed" ? "Removed" : `Show ${kind} block`}
              disabled={kind === "removed" || !canShow}
              onClick={() => onShowBlock(node.id)}
            >
              <Icon
                className={cn(
                  "size-3.5 flex-none",
                  kind === "added" && "text-green-600 dark:text-green-500",
                  kind === "changed" && "text-amber-600 dark:text-amber-500",
                  kind === "removed" && "text-destructive"
                )}
              />
              <span
                className={cn(
                  "truncate",
                  kind === "removed" && "line-through text-muted-foreground"
                )}
              >
                {getNodeTitle(node)}
              </span>
            </button>
          );
        })}
        {rows.length > maxChangeRows && (
          <p className="px-1.5 py-1 text-muted-foreground">
            and {rows.length - maxChangeRows} more
          </p>
        )}
        {rows.length === 0 && (
          <p className="px-1.5 py-1 text-muted-foreground">
            Changed how blocks are connected
          </p>
        )}
      </div>
      <div className="flex items-center gap-2 border-t px-2 py-1.5">
        {change.status === "pending" ? (
          <>
            <Button size="sm" className="h-7 gap-1.5" onClick={onKeep}>
              <CheckIcon className="size-3.5" />
              Keep
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="h-7 gap-1.5"
              onClick={onUndo}
            >
              <Undo2Icon className="size-3.5" />
              Undo
            </Button>
          </>
        ) : (
          <span className="flex items-center gap-1.5 px-1.5 py-1 text-muted-foreground">
            {change.status === "kept" ? (
              <>
                <CheckIcon className="size-3.5" />
                Kept
              </>
            ) : (
              <>
                <Undo2Icon className="size-3.5" />
                Undone
              </>
            )}
          </span>
        )}
      </div>
    </div>
  );
}

// Shown instead of the input when the month's prompts or answers are used up.
function LimitNotice({ usage }: { usage: FlowAIUsage }) {
  const appId = useAppId();
  const plans = useBillingPlans();

  useEffect(() => {
    track("flow_ai_limit_reached", { prompts_limit: usage.prompts_limit });
  }, [usage.prompts_limit]);

  const canUpgrade = plans?.some(
    (p) => p && p.feature_max_ai_prompts_per_month > usage.prompts_limit
  );

  // Usage is counted per calendar month in UTC.
  const now = new Date();
  const reset = new Date(
    Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 1)
  ).toLocaleDateString(undefined, {
    month: "long",
    day: "numeric",
    timeZone: "UTC",
  });

  return (
    <div className="rounded-lg border bg-background p-3 space-y-3 text-sm">
      <p>
        {usage.prompts_used >= usage.prompts_limit
          ? `You've used all ${usage.prompts_limit} AI prompts for this month.`
          : "You've asked the AI too many questions for this month."}{" "}
        They reset on {reset}.
      </p>
      {canUpgrade && (
        <Button size="sm" className="gap-2" asChild>
          <Link
            href={{ pathname: "/apps/[appId]/premium", query: { appId } }}
            onClick={() => track("upsell_clicked", { from: "flow_ai_limit" })}
          >
            <CrownIcon className="size-4" />
            Get more prompts
          </Link>
        </Button>
      )}
    </div>
  );
}
