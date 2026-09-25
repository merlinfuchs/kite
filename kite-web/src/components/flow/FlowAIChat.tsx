import {
  useFlowAIChatMutation,
  useFlowAICheckMutation,
} from "@/lib/api/mutations";
import { checkFlowAIPrompt, runFlowAIPrompt } from "@/lib/flow/ai";
import { useFlowContext } from "@/lib/flow/context";
import { NodeType } from "@/lib/flow/dataSchema";
import { useFlowAIUsage } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { FlowAIChatMessage, FlowAICheckResponse } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import { useReactFlow } from "@xyflow/react";
import {
  LoaderCircleIcon,
  SendHorizontalIcon,
  SparklesIcon,
  SquarePenIcon,
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
import { Button } from "../ui/button";
import { Textarea } from "../ui/textarea";
import FlowAICheckCard from "./FlowAICheckCard";
import FlowAIMarkdown from "./FlowAIMarkdown";
import { FlowEditorApi } from "./FlowEditor";

interface ChatEntry extends FlowAIChatMessage {
  // A request the user can send to make the change the answer suggests.
  buildPrompt?: string;
  // Problems left with the AI's changes.
  issues?: string[];
  repaired?: boolean;
  // The prompt failed, and content is the error.
  failed?: boolean;
}

export default memo(function FlowAIChat({
  editorRef,
  onClose,
}: {
  editorRef: RefObject<FlowEditorApi>;
  onClose: () => void;
}) {
  const context = useFlowContext((c) => c.type);
  const { getNodes, getEdges, fitView } = useReactFlow<NodeType>();
  const appId = useAppId();
  const chat = useFlowAIChatMutation(appId);
  const check = useFlowAICheckMutation(appId);
  const usage = useFlowAIUsage();

  const [entries, setEntries] = useState<ChatEntry[]>([]);
  const [input, setInput] = useState("");
  // What the AI is doing, or null if it's idle.
  const [status, setStatus] = useState<string | null>(null);
  const busy = status !== null;
  // A clearer version of the first prompt to confirm before it's sent.
  const [checked, setChecked] = useState<{
    prompt: string;
    check: FlowAICheckResponse;
  } | null>(null);

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
  }, []);

  const bottomRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [entries, status, checked]);

  const send = useCallback(
    async (content: string) => {
      const messages: FlowAIChatMessage[] = [
        ...entries
          .filter((e) => !e.failed)
          .map(({ role, content }) => ({ role, content })),
        { role: "user", content },
      ];
      setEntries((e) => [...e, { role: "user", content }]);
      setChecked(null);
      setStatus("Working on it...");

      // The rounds of the prompt are undone together.
      const mergeKey = `ai:${Date.now()}`;
      try {
        const res = await runFlowAIPrompt({
          context,
          messages,
          getFlow: () => ({ nodes: getNodes(), edges: getEdges() }),
          applyFlow: ({ nodes, edges }, changedNodeIds) => {
            // Selects the changed blocks, so they stand out.
            const changed = new Set(changedNodeIds);
            editorRef.current?.replaceFlow(
              nodes.map((n) =>
                !!n.selected === changed.has(n.id)
                  ? n
                  : { ...n, selected: changed.has(n.id) }
              ),
              edges,
              mergeKey
            );
          },
          send: chat.mutateAsync,
          signal: abort.current?.signal,
        });
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
            issues: res.issues,
            repaired: res.repairs > 0 && res.issues.length === 0,
          },
        ]);
      } catch (err) {
        setEntries((e) => [
          ...e,
          { role: "assistant", content: (err as Error).message, failed: true },
        ]);
      } finally {
        setStatus(null);
      }
    },
    [entries, context, getNodes, getEdges, fitView, editorRef, chat.mutateAsync]
  );

  const submit = useCallback(async () => {
    const content = input.trim();
    if (!content || busy) return;
    setInput("");

    // Only the first prompt is checked, later ones build on the chat.
    if (entries.length === 0) {
      setStatus("Checking your request...");
      const res = await checkFlowAIPrompt({
        context,
        prompt: content,
        flow: { nodes: getNodes(), edges: getEdges() },
        send: check.mutateAsync,
      });
      setStatus(null);
      if (res?.verdict === "clarify") {
        setChecked({ prompt: content, check: res });
        return;
      }
    }
    send(content);
  }, [
    input,
    busy,
    entries.length,
    context,
    getNodes,
    getEdges,
    check.mutateAsync,
    send,
  ]);

  const limit = usage?.prompts_limit;
  const left = usage ? Math.max(limit! - usage.prompts_used, 0) : undefined;
  const unavailable = limit === 0 || left === 0;

  return (
    <div className="flex-none w-96 flex flex-col bg-muted/30 border-l">
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
            disabled={busy || (entries.length === 0 && !checked)}
            onClick={() => {
              setEntries([]);
              setChecked(null);
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
        {entries.length === 0 && !checked && (
          <div className="text-muted-foreground space-y-2">
            <p>
              Describe what the flow should do, like &quot;Ban the user from the
              argument and log why&quot;. Select blocks to ask about them.
            </p>
            <p>
              Changes are applied right away. You can undo them, and nothing is
              saved until you save.
            </p>
            <p>
              Each message that changes the flow uses one prompt. Answers
              without changes, like questions the AI asks back, and fixes of its
              own changes are free. Start a new chat for unrelated changes.
            </p>
          </div>
        )}
        {entries.map((entry, i) => (
          <ChatBubble key={i} entry={entry} onBuild={fillInput} />
        ))}
        {checked && !busy && (
          <FlowAICheckCard
            prompt={checked.prompt}
            check={checked.check}
            onSend={send}
          />
        )}
        {busy && (
          <div className="flex items-center gap-2 text-muted-foreground">
            <LoaderCircleIcon className="size-4 animate-spin" />
            {status}
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <div className="flex-none p-4 space-y-2">
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
        {!!limit && (
          <div className="text-xs text-muted-foreground">
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
  onBuild,
}: {
  entry: ChatEntry;
  onBuild: (prompt: string) => void;
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
          onClick={() => onBuild(entry.buildPrompt!)}
        >
          <SparklesIcon className="size-4" />
          Build this
        </Button>
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
