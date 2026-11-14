import { useCallback, useMemo, useState } from "react";
import { Node, ReactFlowState, useStore } from "@xyflow/react";
import { LoaderCircleIcon, PlayIcon } from "lucide-react";
import { useFlowTestMutation } from "@/lib/api/mutations";
import { useAppStateGuilds } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { NodeData } from "@/lib/flow/dataSchema";
import { getCommandArguments } from "@/lib/flow/commandArguments";
import { useFlowContext, useFlowContextStore } from "@/lib/flow/context";
import {
  FlowTestState,
  FlowTestTarget,
  buildFlowTestRequest,
  findTestSelectMenu,
  flowTestInputs,
  flowTestPath,
  newFlowTestSeed,
} from "@/lib/flow/flowTest";
import { useGetFlowData } from "@/lib/flow/parametersStore";
import { ComponentTypeStringSelect } from "@/lib/types/message.gen";
import {
  FlowTestAction,
  FlowTestMock,
  FlowTestRequest,
} from "@/lib/types/wire.gen";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { ScrollArea } from "../ui/scroll-area";
import { Textarea } from "../ui/textarea";
import GuildSelect from "../common/GuildSelect";
import ChannelSelect from "../common/ChannelSelect";
import FlowTestArgumentField from "./FlowTestArgumentField";
import { FlowTestCheckboxList, FlowTestField } from "./FlowTestField";
import FlowTestResult from "./FlowTestResult";
import FlowTestMockDialog, {
  FlowTestMockList,
  MockTarget,
} from "./FlowTestMockDialog";
import FlowTestVariables from "./FlowTestVariables";

// The events test runs can't fake yet.
const unsupportedEvents = ["voice_state_update", "guild_boost_remove"];

type FlowNode = Node<NodeData>;

const entryData = (s: ReactFlowState, type: string) =>
  (s.nodes.find((n) => n.type === type) as FlowNode | undefined)?.data;

// The argument blocks of the command, only new when one of them changes, so
// the panel doesn't render again while other blocks are dragged.
const selectArgs = (s: ReactFlowState) => {
  const entry = s.nodes.find((n) => n.type === "entry_command");
  return entry
    ? getCommandArguments(entry as FlowNode, s.nodes as FlowNode[], s.edges)
    : [];
};
const sameNodes = (a: FlowNode[], b: FlowNode[]) =>
  a.length === b.length && a.every((n, i) => n === b[i]);

export default function FlowTestPanel({
  target,
  hideHeader,
  onShowNode,
}: {
  target: FlowTestTarget;
  hideHeader?: boolean;
  onShowNode: (id: string) => void;
}) {
  const appId = useAppId();
  const { mutate, isPending } = useFlowTestMutation(appId);
  const getFlowData = useGetFlowData();

  const eventType =
    useStore((s) => entryData(s, "entry_event")?.event_type) ?? "";
  const commandType =
    useStore((s) => entryData(s, "entry_command")?.command_type) ??
    "chat_input";
  const args = useStore(selectArgs, sameNodes);

  const state = useFlowContext((s) => s.test);
  const store = useFlowContextStore();
  const set = useCallback(
    (patch: Partial<FlowTestState>) =>
      store.getState().setTest((s) => ({ ...s, ...patch })),
    [store]
  );

  const [mockTarget, setMockTarget] = useState<MockTarget | null>(null);

  const guilds = useAppStateGuilds();
  const guildId = state.guildId ?? guilds?.[0]?.id ?? null;
  const inputs = flowTestInputs(eventType);

  const messageData =
    target.type === "message" ? target.getMessage().data : null;
  const selectMenu = useMemo(
    () =>
      target.type === "message"
        ? findTestSelectMenu(messageData, target.flowSourceId)
        : null,
    [target, messageData]
  );

  function run() {
    const req = buildFlowTestRequest(state, {
      target,
      flow: getFlowData(),
      guildId,
      eventType,
      commandType,
      args: Object.fromEntries(
        args.map((a) => [
          a.id,
          { name: a.data.name ?? "", type: a.data.command_argument_type },
        ])
      ),
    });

    req.seed = newFlowTestSeed();
    send(req, true);
  }

  // Sends the test. When a click or submit fails, the result before it stays.
  const send = useCallback(
    (req: FlowTestRequest, isNew: boolean) => {
      const failed = (error: string) =>
        set(isNew ? { last: null, error } : { error });

      set({ error: null });
      mutate(
        { path: flowTestPath(target), req },
        {
          onSuccess(res) {
            if (res.success) {
              set({ last: { request: req, result: res.data } });
            } else {
              failed(
                res.error.code === "app_not_running"
                  ? "Your app is offline. Start it to run a test."
                  : res.error.message
              );
            }
          },
          onError(err) {
            failed(`The test couldn't run: ${err.message}`);
          },
        }
      );
    },
    [mutate, target, set]
  );

  // A changed test result runs the last test again with it, with the same
  // clicks.
  const saveMock = useCallback(
    (nodeId: string, mock: FlowTestMock | undefined) => {
      const { mocks: current, last } = store.getState().test;
      const mocks = { ...current };
      if (mock) {
        mocks[nodeId] = mock;
      } else {
        delete mocks[nodeId];
      }
      set({ mocks });
      setMockTarget(null);
      if (last) send({ ...last.request, mocks }, false);
    },
    [store, set, send]
  );

  // Clicks, picks and submits run the last test again with them.
  const act = useCallback(
    (action: FlowTestAction) => {
      const last = store.getState().test.last;
      if (!last) return;
      send(
        { ...last.request, then: [...(last.request.then ?? []), action] },
        false
      );
    },
    [store, send]
  );

  return (
    <div className="w-full h-full flex flex-col min-h-0 overflow-hidden">
      <ScrollArea className="flex-1 min-h-0 [&_[data-radix-scroll-area-viewport]>div]:!block">
        {!hideHeader && (
          <div className="p-5 pb-3">
            <div className="text-xl font-bold text-foreground mb-2">Test</div>
            <div className="text-muted-foreground text-sm">
              Run the flow as if someone used it. Nothing is sent or changed,
              and it doesn&apos;t cost credits. The bot still reads your server,
              like its members and roles.
            </div>
          </div>
        )}

        <div className="flex flex-col gap-4 px-5 pb-5">
          {unsupportedEvents.includes(eventType) ? (
            <div className="text-sm text-muted-foreground">
              Tests don&apos;t support this event yet.
            </div>
          ) : (
            <>
              {(guilds?.length ?? 0) > 1 && (
                <FlowTestField label="Server">
                  <GuildSelect
                    value={guildId}
                    onChange={(id) => set({ guildId: id, channelId: null })}
                  />
                </FlowTestField>
              )}

              {inputs.place && (
                <>
                  <FlowTestField label="Channel">
                    <ChannelSelect
                      guildId={guildId}
                      value={state.channelId}
                      onChange={(id) => set({ channelId: id })}
                      placeholder="Default channel"
                    />
                  </FlowTestField>
                  <FlowTestField label="Used by">
                    <Input
                      value={state.userId}
                      onChange={(e) => set({ userId: e.target.value })}
                      placeholder="You, or the ID of a member"
                    />
                  </FlowTestField>
                </>
              )}

              {target.type === "command" &&
                commandType === "chat_input" &&
                args.map((arg) => (
                  <FlowTestArgumentField
                    key={arg.id}
                    arg={arg}
                    guildId={guildId}
                    value={state.options[arg.id] ?? ""}
                    onChange={(v) =>
                      set({ options: { ...state.options, [arg.id]: v } })
                    }
                  />
                ))}

              {target.type === "command" && commandType !== "chat_input" && (
                <FlowTestField
                  label={commandType === "user" ? "User" : "Message"}
                >
                  <Input
                    value={state.targetId}
                    onChange={(e) => set({ targetId: e.target.value })}
                    placeholder={
                      commandType === "user"
                        ? "You, or the ID of a member"
                        : "A message of yours, or the ID of a message in the channel"
                    }
                  />
                </FlowTestField>
              )}

              {inputs.message && (
                <FlowTestField label="Message">
                  <Textarea
                    value={state.content}
                    onChange={(e) => set({ content: e.target.value })}
                    placeholder="What the message says"
                  />
                </FlowTestField>
              )}

              {inputs.reaction && (
                <FlowTestField label="Emoji">
                  <Input
                    value={state.emoji}
                    onChange={(e) => set({ emoji: e.target.value })}
                    placeholder="👍"
                  />
                </FlowTestField>
              )}

              {inputs.webhook && (
                <FlowTestField label="Request body">
                  <Textarea
                    value={state.webhookBody}
                    onChange={(e) => set({ webhookBody: e.target.value })}
                    className="font-mono"
                  />
                </FlowTestField>
              )}

              {selectMenu &&
                (selectMenu.type === ComponentTypeStringSelect ? (
                  <FlowTestField label="Picked options">
                    <FlowTestCheckboxList
                      options={(selectMenu.options ?? []).map((o) => ({
                        label: o.label || o.value || "",
                        value: o.value ?? "",
                      }))}
                      value={state.values}
                      onChange={(values) => set({ values })}
                    />
                  </FlowTestField>
                ) : (
                  <FlowTestField label="Picked">
                    <Input
                      value={state.valueIds}
                      onChange={(e) => set({ valueIds: e.target.value })}
                      placeholder="IDs, separated by commas"
                    />
                  </FlowTestField>
                ))}

              <FlowTestMockList
                mocks={state.mocks}
                onEdit={setMockTarget}
                onRemove={(nodeId) => saveMock(nodeId, undefined)}
              />

              <FlowTestVariables
                value={state.variables}
                onChange={(variables) => set({ variables })}
              />

              <Button onClick={run} disabled={isPending} className="self-start">
                {isPending ? (
                  <LoaderCircleIcon className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <PlayIcon className="h-4 w-4 mr-2" />
                )}
                Run test
              </Button>
            </>
          )}

          {state.error && (
            <div className="text-sm text-destructive">{state.error}</div>
          )}

          {state.last && (
            <>
              <FlowTestResult
                result={state.last.result}
                pending={isPending}
                mocks={state.mocks}
                onAction={act}
                onEditMock={setMockTarget}
                onShowNode={onShowNode}
              />
              <Button
                variant="ghost"
                size="sm"
                className="self-start -ml-3 text-muted-foreground"
                onClick={() => set({ last: null, error: null })}
              >
                Clear result
              </Button>
            </>
          )}
        </div>
      </ScrollArea>

      <FlowTestMockDialog
        target={mockTarget}
        mock={mockTarget ? state.mocks[mockTarget.nodeId] : undefined}
        onClose={() => setMockTarget(null)}
        onSave={(mock) => mockTarget && saveMock(mockTarget.nodeId, mock)}
      />
    </div>
  );
}
