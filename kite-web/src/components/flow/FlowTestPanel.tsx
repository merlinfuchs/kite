import { Dispatch, SetStateAction, useMemo } from "react";
import { Node, useEdges, useNodes } from "@xyflow/react";
import { LoaderCircleIcon, PlayIcon } from "lucide-react";
import { useFlowTestMutation } from "@/lib/api/mutations";
import { useAppStateGuilds } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { NodeData } from "@/lib/flow/dataSchema";
import { getCommandArguments } from "@/lib/flow/commandArguments";
import {
  FlowTestState,
  FlowTestTarget,
  flowTestPath,
  newFlowTestSeed,
  withFlowTestAction,
} from "@/lib/flow/flowTest";
import { useGetFlowData } from "@/lib/flow/parametersStore";
import { FlowTestAction, FlowTestRequest } from "@/lib/types/wire.gen";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { ScrollArea } from "../ui/scroll-area";
import { Textarea } from "../ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import GuildSelect from "../common/GuildSelect";
import ChannelSelect from "../common/ChannelSelect";
import RoleSelect from "../common/RoleSelect";
import FlowTestResult from "./FlowTestResult";

// Text, voice, announcement, stage, forum, media
const allChannelTypes = [0, 2, 5, 13, 15, 16];

const messageEvents = [
  "message_create",
  "direct_message_create",
  "message_update",
  "message_delete",
];
const reactionEvents = ["message_reaction_add", "message_reaction_remove"];
// The events test runs can't fake yet.
const unsupportedEvents = ["voice_state_update", "guild_boost_remove"];

export default function FlowTestPanel({
  target,
  state,
  onStateChange,
  hideHeader,
  onShowNode,
}: {
  target: FlowTestTarget;
  state: FlowTestState;
  onStateChange: Dispatch<SetStateAction<FlowTestState>>;
  hideHeader?: boolean;
  onShowNode: (id: string) => void;
}) {
  const appId = useAppId();
  const mutation = useFlowTestMutation(appId);
  const getFlowData = useGetFlowData();

  const nodes = useNodes<Node<NodeData>>();
  const edges = useEdges();
  const commandEntry = nodes.find((n) => n.type === "entry_command");
  const eventEntry = nodes.find((n) => n.type === "entry_event");
  const eventType = eventEntry?.data.event_type ?? "";
  const commandType = commandEntry?.data.command_type ?? "chat_input";
  const args = useMemo(
    () => (commandEntry ? getCommandArguments(commandEntry, nodes, edges) : []),
    [commandEntry, nodes, edges]
  );

  const {
    guildId,
    channelId,
    userId,
    options,
    targetId,
    content,
    emoji,
    webhookBody,
    values,
    valueIds,
    result,
    error,
  } = state;
  const set = (patch: Partial<FlowTestState>) =>
    onStateChange((s) => ({ ...s, ...patch }));

  const guilds = useAppStateGuilds();
  const currentGuildId = guildId ?? guilds?.[0]?.id ?? null;

  const isWebhook = eventType === "webhook";
  const isSchedule = eventType === "schedule_cron";
  const hasPlace = !isWebhook && !isSchedule;
  const selectMenu =
    target.kind === "message"
      ? findSelectMenu(target.getMessage().data, target.flowSourceId)
      : null;

  function run() {
    const req: FlowTestRequest = {
      user_id: userId.trim() || undefined,
      guild_id: currentGuildId ?? undefined,
      channel_id: hasPlace ? channelId ?? undefined : undefined,
    };

    if (target.kind === "message") {
      // The flow of the dialog is only in the message once it's closed.
      const message = target.getMessage();
      req.message = {
        ...message,
        flow_sources: {
          ...message.flow_sources,
          [target.flowSourceId]: getFlowData(),
        },
      };
      req.flow_source_id = target.flowSourceId;
      if (selectMenu?.options) {
        req.values = values;
      } else if (selectMenu) {
        req.values = valueIds
          .split(",")
          .map((v) => v.trim())
          .filter(Boolean);
      }
    } else {
      req.flow_source = getFlowData();
    }

    if (target.kind === "command") {
      if (commandType === "chat_input") {
        req.options = {};
        for (const arg of args) {
          const name = arg.data.name;
          const value = options[arg.id]?.trim();
          if (!name || !value) continue;
          req.options[name] = optionValue(
            arg.data.command_argument_type,
            value
          );
        }
      } else if (targetId.trim()) {
        req.options = { target: targetId.trim() };
      }
    }

    if (messageEvents.includes(eventType) && content) {
      req.event = { content };
    } else if (reactionEvents.includes(eventType) && emoji.trim()) {
      req.event = { emoji: { name: emoji.trim() } };
    } else if (isWebhook) {
      req.webhook = { body: webhookBody };
    }

    req.seed = newFlowTestSeed();
    send(req, true);
  }

  // Clicks, picks and submits run the test again with them. When that fails,
  // the result before them stays.
  function act(action: FlowTestAction) {
    if (state.request) send(withFlowTestAction(state.request, action), false);
  }

  function send(req: FlowTestRequest, isNew: boolean) {
    set({ error: null });
    mutation.mutate(
      { path: flowTestPath(target), req },
      {
        onSuccess(res) {
          if (res.success) {
            set({ request: req, result: res.data });
          } else {
            set({
              ...(isNew && { request: null, result: null }),
              error:
                res.error.code === "app_not_running"
                  ? "Your app is offline. Start it to run a test."
                  : res.error.message,
            });
          }
        },
        onError(err) {
          set({
            ...(isNew && { request: null, result: null }),
            error: `The test couldn't run: ${err.message}`,
          });
        },
      }
    );
  }

  // Isolated, as message previews stack above the block settings otherwise.
  return (
    <div className="w-full h-full flex flex-col min-h-0 overflow-hidden isolate">
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
                <Field label="Server">
                  <GuildSelect
                    value={currentGuildId}
                    onChange={(id) => set({ guildId: id, channelId: null })}
                  />
                </Field>
              )}

              {hasPlace && (
                <>
                  <Field label="Channel">
                    <ChannelSelect
                      guildId={currentGuildId}
                      value={channelId}
                      onChange={(id) => set({ channelId: id })}
                      placeholder="Default channel"
                    />
                  </Field>
                  <Field label="Used by">
                    <Input
                      value={userId}
                      onChange={(e) => set({ userId: e.target.value })}
                      placeholder="You, or the ID of a member"
                    />
                  </Field>
                </>
              )}

              {target.kind === "command" &&
                commandType === "chat_input" &&
                args.map((arg) => (
                  <ArgumentField
                    key={arg.id}
                    arg={arg}
                    guildId={currentGuildId}
                    value={options[arg.id] ?? ""}
                    onChange={(v) =>
                      set({ options: { ...options, [arg.id]: v } })
                    }
                  />
                ))}

              {target.kind === "command" && commandType !== "chat_input" && (
                <Field label={commandType === "user" ? "User" : "Message"}>
                  <Input
                    value={targetId}
                    onChange={(e) => set({ targetId: e.target.value })}
                    placeholder={
                      commandType === "user"
                        ? "You, or the ID of a member"
                        : "A message of yours, or the ID of a message in the channel"
                    }
                  />
                </Field>
              )}

              {messageEvents.includes(eventType) && (
                <Field label="Message">
                  <Textarea
                    value={content}
                    onChange={(e) => set({ content: e.target.value })}
                    placeholder="What the message says"
                  />
                </Field>
              )}

              {reactionEvents.includes(eventType) && (
                <Field label="Emoji">
                  <Input
                    value={emoji}
                    onChange={(e) => set({ emoji: e.target.value })}
                    placeholder="👍"
                  />
                </Field>
              )}

              {isWebhook && (
                <Field label="Request body">
                  <Textarea
                    value={webhookBody}
                    onChange={(e) => set({ webhookBody: e.target.value })}
                    className="font-mono"
                  />
                </Field>
              )}

              {selectMenu && !selectMenu.options && (
                <Field label="Picked">
                  <Input
                    value={valueIds}
                    onChange={(e) => set({ valueIds: e.target.value })}
                    placeholder="IDs, separated by commas"
                  />
                </Field>
              )}

              {selectMenu?.options && (
                <Field label="Picked options">
                  <div className="flex flex-col gap-1.5">
                    {selectMenu.options.map((o) => (
                      <label
                        key={o.value}
                        className="flex items-center gap-2 text-sm"
                      >
                        <input
                          type="checkbox"
                          checked={values.includes(o.value)}
                          onChange={(e) =>
                            set({
                              values: e.target.checked
                                ? [...values, o.value]
                                : values.filter((x) => x !== o.value),
                            })
                          }
                        />
                        {o.label}
                      </label>
                    ))}
                  </div>
                </Field>
              )}

              <Button
                onClick={run}
                disabled={mutation.isPending}
                className="self-start"
              >
                {mutation.isPending ? (
                  <LoaderCircleIcon className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <PlayIcon className="h-4 w-4 mr-2" />
                )}
                Run test
              </Button>
            </>
          )}

          {error && <div className="text-sm text-destructive">{error}</div>}

          {result && (
            <>
              <FlowTestResult
                result={result}
                pending={mutation.isPending}
                onAction={act}
                onShowNode={onShowNode}
              />
              <Button
                variant="ghost"
                size="sm"
                className="self-start -ml-3 text-muted-foreground"
                onClick={() =>
                  set({ request: null, result: null, error: null })
                }
              >
                Clear result
              </Button>
            </>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label className="text-sm">{label}</Label>
      {children}
    </div>
  );
}

function ArgumentField({
  arg,
  guildId,
  value,
  onChange,
}: {
  arg: Node<NodeData>;
  guildId: string | null;
  value: string;
  onChange: (value: string) => void;
}) {
  const type = arg.data.command_argument_type;
  const choices = (arg.data.command_argument_choices ?? []).flatMap((c) =>
    c.name && c.value ? [{ name: c.name, value: c.value }] : []
  );
  const label = arg.data.command_argument_required
    ? arg.data.name ?? ""
    : `${arg.data.name ?? ""} (optional)`;

  let input: React.ReactNode;
  if (choices.length > 0) {
    input = (
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger>
          <SelectValue placeholder="Pick a choice" />
        </SelectTrigger>
        <SelectContent>
          {choices.map((c) => (
            <SelectItem key={c.value} value={c.value}>
              {c.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  } else if (type === "boolean") {
    input = (
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger>
          <SelectValue placeholder="Pick one" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="true">True</SelectItem>
          <SelectItem value="false">False</SelectItem>
        </SelectContent>
      </Select>
    );
  } else if (type === "channel") {
    input = (
      <ChannelSelect
        guildId={guildId}
        value={value || null}
        onChange={(v) => onChange(v ?? "")}
        types={allChannelTypes}
      />
    );
  } else if (type === "role") {
    input = (
      <RoleSelect
        guildId={guildId}
        value={value || null}
        onChange={(v) => onChange(v ?? "")}
      />
    );
  } else {
    input = (
      <Input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        type={type === "integer" || type === "number" ? "number" : "text"}
        placeholder={argumentPlaceholders[type ?? ""]}
      />
    );
  }

  return <Field label={label}>{input}</Field>;
}

const argumentPlaceholders: Record<string, string> = {
  user: "ID of a member",
  mentionable: "ID of a member or role",
  attachment: "Link to a file",
};

function optionValue(type: string | undefined, value: string): unknown {
  switch (type) {
    case "integer":
    case "number":
      return Number(value);
    case "boolean":
      return value === "true";
    default:
      return value;
  }
}

// The select menu with the flow, with its options if they are fixed, or null
// if the flow is of a button.
function findSelectMenu(
  data: unknown,
  flowSourceId: string
): { options?: { label: string; value: string }[] } | null {
  if (Array.isArray(data)) {
    for (const item of data) {
      const res = findSelectMenu(item, flowSourceId);
      if (res) return res;
    }
    return null;
  }
  if (!data || typeof data !== "object") return null;

  const obj = data as Record<string, any>;
  if (obj.flow_source_id === flowSourceId) {
    // 3 is a select menu of options, 5 to 8 of users, roles or channels.
    if (obj.type === 3) {
      return {
        options: (obj.options ?? []).map((o: any) => ({
          label: o.label || o.value,
          value: o.value,
        })),
      };
    }
    return obj.type >= 5 && obj.type <= 8 ? {} : null;
  }
  for (const value of Object.values(obj)) {
    const res = findSelectMenu(value, flowSourceId);
    if (res) return res;
  }
  return null;
}
