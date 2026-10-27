import { discordEmojiUrl } from "@/tools/common/utils/discordCdn";
import { discordEventTypes } from "@/lib/flow/eventTypes";
import { auditLogActions } from "@/lib/flow/auditLogActions";
import {
  decodePermissionsBitset,
  encodePermissionsBitset,
  permissionBits,
} from "@/lib/discord/permissions";
import { getNodeCreditsCost, useNodeValues } from "@/lib/flow/nodes";
import { duplicateFlowNodes } from "@/lib/flow/clipboard";
import { activityTypeOptions, statusOptions } from "@/lib/discord/presence";
import { useAppFeature } from "@/lib/hooks/api";
import { getFlowCredits } from "@/lib/flow/credits";
import { describeList } from "@/lib/flow/list";
import {
  getCommandArguments,
  moveCommandArgument,
} from "@/lib/flow/commandArguments";
import { aiModelTiers, getAiModelTier } from "@/lib/flow/aiModels";
import { getBlockDefinition } from "@/lib/blocks";
import { parseColorValue } from "@/tools/common/utils/color";
import { aiScoreMaxLevels } from "@/lib/blocks/aiScore";
import { randomModes } from "@/lib/blocks/randomGenerate";
import {
  lookupMaxRows,
  lookupModes,
  parseLookupLines,
} from "@/lib/blocks/valueLookup";
import ColorPicker from "@/tools/common/components/ColorPicker";
import { integrations, needsCredential } from "@/lib/integrations";
import {
  discordApiOperationLabel,
  discordApiOperations,
  getDiscordApiOperation,
} from "@/lib/flow/discordApi";
import { EventTypeScheduleCron, EventTypeWebhook } from "@/lib/types/flow.gen";
import { useAppId } from "@/lib/hooks/params";
import {
  CommandArgumentChoiceData,
  DiscordAPIRequestData,
  EmojiData,
  HTTPRequestData,
  LookupRow,
  PermissionOverwriteData,
  PollAnswerData,
  PollData,
  StatusData,
} from "@/lib/types/flow.gen";
import {
  Node,
  useEdges,
  useNodes,
  useReactFlow,
  useStoreApi,
} from "@xyflow/react";
import {
  booleanEventFilterTargets,
  eventFilterTargetLabels,
  getEventFilterModes,
  getEventFilterTargets,
} from "@/lib/flow/eventFilterTargets";
import {
  CheckIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  CircleAlertIcon,
  CopyIcon,
  HelpCircleIcon,
  MinusIcon,
  PencilIcon,
  PlusIcon,
  SmileIcon,
  TrashIcon,
  XIcon,
} from "lucide-react";
import Link from "next/link";
import {
  ReactNode,
  memo,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  FlowReadOnlyContext,
  PreviewInert,
  useFlowMessages,
  useFlowVariables,
} from "@/lib/flow/context";
import { NodeData, NodeProps } from "../../lib/flow/dataSchema";
import FlowMessageViewDialog from "./FlowMessageViewDialog";
import MessageCreateDialog from "../app/MessageCreateDialog";
import VariableCreateDialog from "../app/VariableCreateDialog";
import EmojiPicker from "../common/EmojiPicker";
import ConfirmDialog from "../common/ConfirmDialog";
import EntitySelect from "../common/EntitySelect";
import JsonEditor from "../common/JsonEditor";
import PlaceholderInput from "../common/PlaceholderInput";
import WebhookUrlInput from "../common/WebhookUrlInput";
import {
  webhookVerifyAlgorithms,
  webhookVerifyEncodings,
  webhookVerifyPresets,
} from "@/lib/blocks/controlWebhookVerify";
import ScheduleCronPreview, {
  ScheduleCronHelp,
} from "../common/ScheduleCronPreview";
import Twemoji from "../common/Twemoji";
import MessageEditorDialog from "../message/MessageEditorDialog";
import { hasComponentsV2Flag } from "@/lib/message/schema";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { Input } from "../ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "../ui/command";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Switch } from "../ui/switch";
import { Textarea } from "../ui/textarea";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import FlowJsonInput from "./FlowJsonInput";
import FlowPlaceholderExplorer from "./FlowPlaceholderExplorer";
import ModalEditorDialog from "./modal/ModalEditorDialog";
import env from "@/lib/env/client";
import { ScrollArea } from "../ui/scroll-area";
import { cn } from "@/lib/utils";

interface Props {
  nodeId: string;
  className?: string;
  hideTitle?: boolean;
}

interface InputProps {
  id: string;
  type: string;
  // The setting a generic input of a block's field edits.
  name?: string;
  data: NodeData;
  updateData: (newData: Partial<NodeData>) => void;
  errors: Record<string, string>;
}

// Fields that open a dialog with more settings, which previews can open to see
// them. Their inputs make the rest of themselves read-only in previews.
export const dialogFields = new Set([
  "message_data",
  "message_template_id",
  "modal_data",
  "channel_data",
  "thread_data",
  "http_request_data",
  "discord_api_request_data",
  "lookup_rows",
]);

// The editor inputs of block settings, by the names blocks use in their
// fields. A test checks every block's inputs exist.
export const settingInputs: Record<string, any> = {
  custom_label: CustomLabelInput,
  temporary_name: TemporaryNameInput,
  name: NameInput,
  description: DescriptionInput,
  command_argument_order: CommandArgumentOrderInput,
  command_argument_type: CommandArgumentTypeInput,
  command_argument_required: CommandArgumentRequiredInput,
  command_argument_min_value: CommandArgumentMinValueInput,
  command_argument_max_value: CommandArgumentMaxValueInput,
  command_argument_max_length: CommandArgumentMaxLengthInput,
  command_argument_choices: CommandArgumentChoicesInput,
  command_disabled_contexts: CommandContextsInput,
  command_disabled_integrations: CommandIntegrationsInput,
  command_permissions: CommandPermissionsInput,
  cooldown_scope: CooldownScopeInput,
  cooldown_duration_seconds: CooldownDurationSecondsInput,
  cooldown_message: CooldownMessageInput,
  event_type: EventTypeInput,
  event_schedule_cron: EventScheduleCronInput,
  event_filter_target: EventFilterTargetInput,
  event_filter_mode: EventFilterModeInput,
  event_filter_value: EventFilterValueInput,
  event_filter_ignore_case: EventFilterIgnoreCaseInput,
  message_data: MessageDataInput,
  message_template_id: MessageTemplateInput,
  message_target: MessageTargetInput,
  emoji_data: EmojiDataInput,
  poll_data: PollDataInput,
  response_target: ResponseTargetInput,
  reply_target: ReplyTargetInput,
  message_ephemeral: MessageEphemeralInput,
  modal_data: ModalDataInput,
  channel_data: ChannelDataInput,
  thread_data: ThreadDataInput,
  forum_post_data: ForumPostDataInput,
  channel_target: ChannelTargetInput,
  forward_source_channel_target: ForwardSourceChannelTargetInput,
  forward_channel_target: ForwardChannelTargetInput,
  voice_self_mute: VoiceSelfMuteInput,
  voice_self_deaf: VoiceSelfDeafInput,
  status_data: StatusDataInput,
  role_data: RoleDataInput,
  role_target: RoleTargetInput,
  variable_id: VariableIdInput,
  variable_scope: VariableScopeInput,
  variable_operation: VariableOperationInput,
  variable_value: VariableValueInput,
  http_request_data: HttpRequestDataInput,
  discord_api_request_data: DiscordApiRequestDataInput,
  webhook_verify: WebhookVerifyInput,
  webhook_response_status: WebhookResponseStatusInput,
  webhook_response_content_type: WebhookResponseContentTypeInput,
  webhook_response_body: WebhookResponseBodyInput,
  ai_chat_completion_data: AiChatCompletionDataInput,
  ai_web_search_data: AiWebSearchDataInput,
  ai_choice_data: AiChoiceDataInput,
  ai_check_data: AiCheckDataInput,
  ai_score_data: AiScoreDataInput,
  ai_choice_name: AiChoiceNameInput,
  ai_choice_description: AiChoiceDescriptionInput,
  expression: ExpressionInput,
  random: RandomInput,
  lookup_value: LookupValueInput,
  lookup_mode: LookupModeInput,
  lookup_ignore_case: LookupIgnoreCaseInput,
  lookup_rows: LookupRowsInput,
  lookup_default: LookupDefaultInput,
  audit_log_reason: AuditLogReasonInput,
  user_target: UserTargetInput,
  guild_target: GuildTargetInput,
  member_ban_delete_message_duration_seconds:
    MemberBanDeleteMessageDurationInput,
  member_timeout_duration_seconds: MemberTimeoutDurationInput,
  member_data: MemberNickInput,
  roblox_user_target: RobloxUserTargetInput,
  roblox_lookup_mode: RobloxLookupModeInput,
  log_level: LogLevelInput,
  log_message: LogMessageInput,
  condition_compare_base_value: ConditionCompareBaseValueInput,
  condition_item_compare_mode: ConditionItemCompareModeInput,
  condition_item_compare_value: ConditionItemCompareValueInput,
  condition_item_ignore_case: ConditionItemIgnoreCaseInput,
  condition_user_base_value: ConditionUserBaseValueInput,
  condition_item_user_mode: ConditionItemUserModeInput,
  condition_item_user_value: ConditionItemUserValueInput,
  condition_channel_base_value: ConditionChannelBaseValueInput,
  condition_item_channel_mode: ConditionItemChannelModeInput,
  condition_item_channel_value: ConditionItemChannelValueInput,
  condition_role_base_value: ConditionRoleBaseValueInput,
  condition_item_role_mode: ConditionItemRoleModeInput,
  condition_item_role_value: ConditionItemRoleValueInput,
  condition_allow_multiple: ConditionAllowMultipleInput,
  loop_count: ControlLoopCountInput,
  sleep_duration_seconds: ControlSleepDurationInput,
};

function nodeTypeDocsPage(nodeType: string) {
  let group: string | null = null;
  if (nodeType.startsWith("action_")) {
    group = "actions";
  } else if (nodeType.startsWith("control_")) {
    group = "controls";
  } else if (nodeType.startsWith("option_")) {
    group = "options";
  } else if (nodeType.startsWith("entry_")) {
    group = "entries";
  }

  if (!group) {
    return null;
  }

  return (
    env.NEXT_PUBLIC_DOCS_LINK + "/reference/blocks/" + group + "/" + nodeType
  );
}

export default function FlowNodeEditor({
  nodeId,
  className,
  hideTitle,
}: Props) {
  const { setNodes, setEdges, getNodes, getEdges, deleteElements } =
    useReactFlow<Node<NodeData>>();
  const store = useStoreApi();

  function close() {
    store.getState().addSelectedNodes([]);
  }

  const nodes = useNodes<Node<NodeData>>();
  // Flows that are only shown, like the ones of a template, show the settings
  // of their blocks without letting them change.
  const readOnly = useContext(FlowReadOnlyContext);

  const node = nodes.find((n) => n.id === nodeId);
  const data = node?.data;

  const updateData = useCallback(
    (newData: Partial<NodeData>) => {
      // Nothing in a preview changes, even where an input misses it.
      if (readOnly) return;
      setNodes((nodes) =>
        nodes.map((n) => {
          if (n.id === nodeId) {
            return {
              ...n,
              data: {
                ...n.data,
                ...newData,
              },
            };
          }
          return n;
        })
      );
    },
    [setNodes, nodeId, readOnly]
  );

  function deleteNode() {
    deleteElements({
      nodes: [node!],
    });
  }

  function duplicateNode() {
    if (!node) return;

    const [newNodes, newEdges] = duplicateFlowNodes(
      [node.id],
      getNodes(),
      getEdges()
    );
    setNodes((nodes) =>
      nodes.concat(newNodes.map((n) => ({ ...n, selected: false })))
    );
    setEdges((edges) => edges.concat(newEdges));
  }

  const values = useNodeValues(node?.type!);

  const appId = useAppId();
  const premiumFeature = values.premiumFeature;
  const hasPremiumFeature = useAppFeature((f) =>
    premiumFeature ? !!f[premiumFeature] : true
  );

  const errors: Record<string, string> = useMemo(() => {
    if (!values.dataSchema) return {};

    const res = values.dataSchema.safeParse(data);
    if (res.success) {
      return {};
    }

    return Object.fromEntries(
      res.error.issues.map((issue) => [issue.path.join("."), issue.message])
    );
  }, [values.dataSchema, data]);

  if (!node || !data) return null;

  const creditsCost = getNodeCreditsCost(values, data);

  const docsPage = nodeTypeDocsPage(node.type!);

  return (
    <div
      className={cn(
        "bg-background flex flex-col min-h-0 overflow-hidden",
        className ?? "absolute top-0 left-0 w-96 h-full"
      )}
    >
      <ScrollArea className="flex-1 min-h-0 w-full">
        <div className="p-5">
          <div className="flex-none">
            {!hideTitle && (
              <div className="flex items-start justify-between mb-5">
                <div className="text-xl font-bold text-foreground">
                  Block Settings
                </div>
                <XIcon
                  className="h-6 w-6 text-muted-foreground hover:text-foreground cursor-pointer"
                  onClick={close}
                />
              </div>
            )}
            <div className="mb-5">
              <div className="flex items-center gap-1.5">
                <div className="text-lg font-bold text-foreground mb-1">
                  {values.defaultTitle}
                </div>
                {docsPage && (
                  <Link href={docsPage} target="_blank">
                    <HelpCircleIcon className="h-5 w-5 text-muted-foreground hover:text-foreground" />
                  </Link>
                )}
              </div>
              <div className="text-muted-foreground mb-3">
                {values.defaultDescription}
              </div>
              <div className="flex gap-2 flex-wrap">
                <div className="bg-muted rounded px-2 py-1 text-xs">
                  {nodeId}
                </div>
                {!!creditsCost && (
                  <div className="bg-muted rounded px-2 py-1 text-xs flex gap-1">
                    {creditsCost}
                    <div>credit{creditsCost === 1 ? "" : "s"}</div>
                  </div>
                )}
              </div>
            </div>
          </div>
          <div className="space-y-3 flex-auto">
            {hasPremiumFeature === false && (
              <div className="text-sm text-muted-foreground bg-muted rounded p-3">
                This block requires{" "}
                <Link
                  href={`/apps/${appId}/premium`}
                  target="_blank"
                  className="text-primary hover:underline"
                >
                  Premium
                </Link>{" "}
                and fails without it.
              </div>
            )}
            {values.dataFields.map((field) => {
              // Fields of a block's definition without an input of their own.
              const name = field.startsWith("field:")
                ? field.slice("field:".length)
                : undefined;
              const Input = name ? BlockFieldInput : settingInputs[field];
              if (!Input) return null;

              const input = (
                <Input
                  key={`${nodeId}-${field}`}
                  id={nodeId}
                  type={node.type}
                  name={name}
                  data={data}
                  updateData={updateData}
                  errors={errors}
                />
              );
              // Previews can't change settings, but can open the dialogs of
              // fields to see what's in them.
              return readOnly && !dialogFields.has(field) ? (
                <div
                  key={`${nodeId}-${field}`}
                  inert
                  className="space-y-3 empty:hidden"
                >
                  {input}
                </div>
              ) : (
                input
              );
            })}
          </div>
          <div className="flex-none space-y-3 mt-5">
            {!values.fixed && !readOnly && (
              <>
                <Button
                  variant="destructive"
                  onClick={deleteNode}
                  className="w-full flex gap-2"
                >
                  <TrashIcon className="h-5 w-5" />
                  <div>Delete Block</div>
                </Button>
                <Button
                  variant="secondary"
                  onClick={duplicateNode}
                  className="w-full flex gap-2"
                >
                  <CopyIcon className="h-5 w-5" />
                  <div>Duplicate Block</div>
                </Button>
              </>
            )}
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}

function CustomLabelInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="custom_label"
      title="Custom Label"
      description="Set a custom label for this block so its easier to recognize. This is optional."
      value={data.custom_label || ""}
      updateValue={(v) => updateData({ custom_label: v || undefined })}
      errors={errors}
    />
  );
}

function TemporaryNameInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="temporary_name"
      title="Temporary Variable"
      description="Store the result of this block in a temporary variable to use later."
      value={data.temporary_name || ""}
      updateValue={(v) => updateData({ temporary_name: v || undefined })}
      errors={errors}
    />
  );
}

function NameInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="name"
      title="Name"
      value={data.name || ""}
      updateValue={(v) => updateData({ name: v || undefined })}
      errors={errors}
    />
  );
}

function DescriptionInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="description"
      title="Description"
      value={data.description || ""}
      updateValue={(v) => updateData({ description: v || undefined })}
      errors={errors}
    />
  );
}

function CommandArgumentOrderInput({ id, updateData }: InputProps) {
  const nodes = useNodes<Node<NodeData>>();
  const edges = useEdges();

  const args = useMemo(() => {
    const entry = nodes.find((n) => n.id === id);
    return entry ? getCommandArguments(entry, nodes, edges) : [];
  }, [id, nodes, edges]);

  if (args.length < 2) return null;

  function move(index: number, direction: -1 | 1) {
    const order = moveCommandArgument(args, index, direction);
    if (order) {
      updateData({ command_argument_order: order });
    }
  }

  return (
    <div>
      <div className="font-medium text-foreground mb-2">Argument Order</div>
      <div className="text-muted-foreground text-sm mb-2">
        The order of the arguments in Discord. Required arguments always come
        before optional ones.
      </div>
      <div className="flex flex-col gap-2">
        {args.map((arg, i) => (
          <div
            className="flex items-center gap-2 border rounded-md pl-3 pr-1 py-1"
            key={arg.id}
          >
            <div className="text-sm text-muted-foreground w-4 flex-none">
              {i + 1}
            </div>
            <div className="text-sm text-foreground truncate flex-auto">
              {arg.data.name || "unnamed"}
            </div>
            <div className="text-xs text-muted-foreground flex-none">
              {arg.data.command_argument_required ? "Required" : "Optional"}
            </div>
            <div className="flex flex-none">
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                disabled={!moveCommandArgument(args, i, -1)}
                onClick={() => move(i, -1)}
                aria-label="Move up"
              >
                <ChevronUpIcon className="h-5 w-5" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                disabled={!moveCommandArgument(args, i, 1)}
                onClick={() => move(i, 1)}
                aria-label="Move down"
              >
                <ChevronDownIcon className="h-5 w-5" />
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function CommandArgumentTypeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="command_argument_type"
      title="Argument Type"
      type="select"
      options={[
        { value: "string", label: "Text" },
        { value: "integer", label: "Whole Number" },
        { value: "number", label: "Decimal Number" },
        { value: "boolean", label: "True/False" },
        { value: "user", label: "User" },
        { value: "channel", label: "Channel" },
        { value: "role", label: "Role" },
        { value: "mentionable", label: "Mentionable" },
        { value: "attachment", label: "Attachment" },
      ]}
      value={data.command_argument_type || ""}
      updateValue={(v) => updateData({ command_argument_type: v || undefined })}
      errors={errors}
    />
  );
}

function CommandArgumentRequiredInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseCheckbox
      field="command_argument_required"
      title="Argument Required"
      value={!!data.command_argument_required}
      updateValue={(v) =>
        updateData({ command_argument_required: v || undefined })
      }
      errors={errors}
    />
  );
}

function CommandArgumentMinValueInput({
  data,
  updateData,
  errors,
}: InputProps) {
  if (
    data.command_argument_type !== "number" &&
    data.command_argument_type !== "integer"
  )
    return null;

  return (
    <BaseInput
      type="text"
      field="command_argument_min_value"
      title="Minimum Value"
      value={data.command_argument_min_value?.toString() || ""}
      updateValue={(v) =>
        updateData({
          command_argument_min_value: parseFloat(v) || undefined,
        })
      }
      errors={errors}
    />
  );
}

function CommandArgumentMaxValueInput({
  data,
  updateData,
  errors,
}: InputProps) {
  if (
    data.command_argument_type !== "number" &&
    data.command_argument_type !== "integer"
  )
    return null;

  return (
    <BaseInput
      type="text"
      field="command_argument_max_value"
      title="Maximum Value"
      value={data.command_argument_max_value?.toString() || ""}
      updateValue={(v) =>
        updateData({
          command_argument_max_value: parseFloat(v) || undefined,
        })
      }
      errors={errors}
    />
  );
}

function CommandArgumentMaxLengthInput({
  data,
  updateData,
  errors,
}: InputProps) {
  if (data.command_argument_type !== "string") return null;

  return (
    <BaseInput
      type="text"
      field="command_argument_max_length"
      title="Maximum Length"
      value={data.command_argument_max_length?.toString() || ""}
      updateValue={(v) =>
        updateData({ command_argument_max_length: parseInt(v) || undefined })
      }
      errors={errors}
    />
  );
}

function CommandArgumentChoicesInput({ data, updateData, errors }: InputProps) {
  const addChoice = useCallback(() => {
    updateData({
      command_argument_choices: [
        ...(data.command_argument_choices || []),
        { name: "", value: "" },
      ],
    });
  }, [updateData, data]);

  const clearChoices = useCallback(() => {
    updateData({
      command_argument_choices: [],
    });
  }, [updateData]);

  const removeChoice = useCallback(
    (i: number) => {
      updateData({
        command_argument_choices: data.command_argument_choices?.filter(
          (_, j) => j !== i
        ),
      });
    },
    [updateData, data]
  );

  const updateChoice = useCallback(
    (i: number, newData: Partial<CommandArgumentChoiceData>) => {
      updateData({
        command_argument_choices: data.command_argument_choices?.map((c, j) =>
          j === i ? { ...c, ...newData } : c
        ),
      });
    },
    [updateData, data]
  );

  if (
    data.command_argument_type !== "string" &&
    data.command_argument_type !== "integer" &&
    data.command_argument_type !== "number"
  )
    return null;

  return (
    <div>
      <div className="font-medium text-foreground mb-2">Choices</div>
      <div className="text-muted-foreground text-sm mb-2">
        Let users choose from a list of options. Leave this blank to allow any
        value.
      </div>
      <div className="flex flex-col gap-3 mb-3">
        {data.command_argument_choices?.map((choice, i) => (
          <div className="flex gap-2" key={i}>
            <Input
              type="text"
              placeholder="Name"
              value={choice.name || ""}
              onChange={(e) => updateChoice(i, { name: e.target.value })}
            />
            <Input
              type="text"
              placeholder="Value"
              value={choice.value || ""}
              onChange={(e) => updateChoice(i, { value: e.target.value })}
            />
            <Button
              variant="outline"
              size="icon"
              className="flex-none"
              aria-label="Remove choice"
              onClick={() => removeChoice(i)}
            >
              <MinusIcon className="h-5 w-5" />
            </Button>
          </div>
        ))}
      </div>
      <div className="flex space-x-2">
        <Button variant="outline" size="icon" onClick={addChoice}>
          <PlusIcon className="h-5 w-5" />
        </Button>
        <Button variant="outline" size="icon" onClick={clearChoices}>
          <TrashIcon className="h-5 w-5" />
        </Button>
      </div>
    </div>
  );
}

function CommandPermissionsInput({ data, updateData, errors }: InputProps) {
  return (
    <BasePermissionInput
      field="command_permissions"
      title="Required Permissions"
      value={data.command_permissions || "0"}
      updateValue={(v) =>
        updateData({
          command_permissions: v === "0" ? undefined : v,
        })
      }
      errors={errors}
      searchable
    />
  );
}

function CooldownScopeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="cooldown_scope"
      title="Cooldown Scope"
      description="Who the cooldown applies to."
      type="select"
      options={[
        { value: "user", label: "Per User" },
        { value: "guild", label: "Per Server" },
        { value: "global", label: "Global" },
      ]}
      value={data.cooldown_scope || "user"}
      updateValue={(v) => updateData({ cooldown_scope: v || undefined })}
      errors={errors}
    />
  );
}

function CooldownDurationSecondsInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      field="cooldown_duration_seconds"
      title="Cooldown Duration"
      description="How many seconds the command is on cooldown for after it's used, up to 3600 (1 hour). Cooldowns reset when Kite restarts, so use stored variables for longer ones."
      value={data.cooldown_duration_seconds || ""}
      updateValue={(v) =>
        updateData({
          cooldown_duration_seconds: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function CooldownMessageInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="cooldown_message"
      title="Cooldown Message"
      description="Shown when someone uses the command while it's on cooldown. Use {{var('cooldown_remaining')}} to show how many seconds are left. Leave empty for a default message."
      value={data.cooldown_message || ""}
      updateValue={(v) => updateData({ cooldown_message: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

const availableCommandContextsValues = ["guild", "bot_dm", "private_channel"];

function CommandContextsInput({ data, updateData, errors }: InputProps) {
  const values = useMemo(() => {
    return availableCommandContextsValues.filter(
      (v) => !data.command_disabled_contexts?.includes(v)
    );
  }, [data.command_disabled_contexts]);

  const updateValues = useCallback(
    (values: string[]) => {
      const newValues = availableCommandContextsValues.filter(
        (v) => !values.includes(v)
      );

      updateData({
        command_disabled_contexts: newValues.length > 0 ? newValues : undefined,
      });
    },
    [updateData]
  );

  return (
    <>
      <BaseMultiSelect
        field="command_disabled_contexts"
        title="Contexts"
        values={values}
        options={[
          { value: "guild", label: "In Servers" },
          { value: "bot_dm", label: "In Bot DMs" },
          { value: "private_channel", label: "In Other DMs" },
        ]}
        updateValues={updateValues}
        errors={errors}
      />
    </>
  );
}

const availableCommandIntegrationsValues = ["guild_install", "user_install"];

function CommandIntegrationsInput({ data, updateData, errors }: InputProps) {
  const values = useMemo(() => {
    return availableCommandIntegrationsValues.filter(
      (v) => !data.command_disabled_integrations?.includes(v)
    );
  }, [data.command_disabled_integrations]);

  const updateValues = useCallback(
    (values: string[]) => {
      const newValues = availableCommandIntegrationsValues.filter(
        (v) => !values.includes(v)
      );

      updateData({
        command_disabled_integrations:
          newValues.length > 0 ? newValues : undefined,
      });
    },
    [updateData]
  );

  return (
    <BaseMultiSelect
      field="command_disabled_integrations"
      title="Integrations Types"
      values={values}
      options={[
        { value: "guild_install", label: "Server Install" },
        { value: "user_install", label: "User Install" },
      ]}
      updateValues={updateValues}
      errors={errors}
    />
  );
}

function EventTypeInput({ data, updateData, errors }: InputProps) {
  // Scheduled and webhook listeners can't become Discord listeners or the
  // other way around.
  if (data.event_type === EventTypeScheduleCron) return null;
  if (data.event_type === EventTypeWebhook) return <WebhookUrlInput />;

  return (
    <BaseInput
      type="select"
      field="event_type"
      title="Event"
      options={discordEventTypes.map((t) => ({
        value: t.value,
        label: t.label,
      }))}
      value={data.event_type || ""}
      updateValue={(v) => updateData({ event_type: v || undefined })}
      errors={errors}
    />
  );
}

function EventScheduleCronInput(props: InputProps) {
  if (props.data.event_type !== EventTypeScheduleCron) return null;
  return <ScheduleCronInput {...props} />;
}

function ScheduleCronInput({ data, updateData, errors }: InputProps) {
  const nodes = useNodes<Node<NodeData>>();
  const edges = useEdges();

  const cron = data.event_schedule_cron || "";
  const creditsPerRun = useMemo(
    () => getFlowCredits(nodes, edges).max,
    [nodes, edges]
  );

  return (
    <div className="space-y-2">
      <BaseInput
        field="event_schedule_cron"
        title="Schedule"
        description={
          <>
            <ScheduleCronHelp /> Add a leading seconds field for sub-minute
            schedules.
          </>
        }
        value={cron}
        updateValue={(v) => updateData({ event_schedule_cron: v || undefined })}
        errors={errors}
        placeholder="*/5 * * * *"
      />
      <ScheduleCronPreview cron={cron} creditsPerRun={creditsPerRun} />
    </div>
  );
}

// useConnectedEventType returns the event type of the listener an event
// filter is connected to, if any.
function useConnectedEventType(filterId: string) {
  const edges = useEdges();
  const nodes = useNodes<Node<NodeData>>();

  return useMemo(() => {
    const edge = edges.find((e) => e.source === filterId);
    const entry = edge && nodes.find((n) => n.id === edge.target);
    return entry?.type === "entry_event" ? entry.data.event_type : undefined;
  }, [edges, nodes, filterId]);
}

function EventFilterTargetInput({ id, data, updateData, errors }: InputProps) {
  const eventType = useConnectedEventType(id);
  const targets = getEventFilterTargets(eventType);
  const target = data.event_filter_target || "";

  const targetErrors =
    target && eventType && !targets.includes(target)
      ? {
          ...errors,
          event_filter_target: `${
            eventFilterTargetLabels[target] || target
          } isn't available for ${eventType
            .split("_")
            .join(" ")} events, so the listener would never run.`,
        }
      : errors;

  return (
    <BaseInput
      type="select"
      field="event_filter_target"
      title="Filter Target"
      options={targets.map((t) => ({
        value: t,
        label: eventFilterTargetLabels[t] || t,
      }))}
      value={target}
      updateValue={(v) => {
        // Some targets only take some modes, and yes or no filters start as
        // "yes". Yes or no filters have no value, emojis and actions are
        // picked, so no other value carries over to or from them.
        const isPicked = (t: string) =>
          booleanEventFilterTargets.includes(t) ||
          t === "emoji" ||
          t === "audit_log_action";

        const modes = getEventFilterModes(v);
        const mode =
          modes && !modes.includes(data.event_filter_mode || "")
            ? modes[0]
            : data.event_filter_mode;

        updateData({
          event_filter_target: v || undefined,
          event_filter_mode: mode,
          event_filter_value:
            isPicked(v) || isPicked(target)
              ? undefined
              : data.event_filter_value,
          event_filter_ignore_case:
            v === "message_content" ? data.event_filter_ignore_case : undefined,
        });
      }}
      errors={targetErrors}
    />
  );
}

const oneOfModes = [
  { value: "one_of", label: "Is One Of" },
  { value: "not_one_of", label: "Is Not One Of" },
];

function EventFilterModeInput({ data, updateData, errors }: InputProps) {
  const target = data.event_filter_target || "";
  const isBoolean = booleanEventFilterTargets.includes(target);

  const options = isBoolean
    ? [
        { value: "equal", label: "Yes" },
        { value: "not_equal", label: "No" },
      ]
    : target === "role_id"
    ? [
        { value: "equal", label: "Has Role" },
        { value: "not_equal", label: "Doesn't Have Role" },
      ]
    : target === "emoji"
    ? [
        { value: "equal", label: "Is" },
        { value: "not_equal", label: "Is Not" },
      ]
    : target === "audit_log_action"
    ? oneOfModes
    : [
        { value: "equal", label: "Equal" },
        { value: "not_equal", label: "Not Equal" },
        { value: "contains", label: "Contains" },
        { value: "starts_with", label: "Starts With" },
        { value: "ends_with", label: "Ends With" },
        ...oneOfModes,
      ];

  return (
    <BaseInput
      type="select"
      field="event_filter_mode"
      title={isBoolean ? eventFilterTargetLabels[target] : "Filter Mode"}
      description={
        target === "user_is_bot"
          ? "Choose No to ignore bots."
          : target === "message_mentions_bot"
          ? "Whether the message pings your bot."
          : undefined
      }
      options={options}
      value={data.event_filter_mode || ""}
      updateValue={(v) => updateData({ event_filter_mode: v || undefined })}
      errors={errors}
    />
  );
}

function EventFilterValueInput({ data, updateData, errors }: InputProps) {
  const target = data.event_filter_target || "";

  // Yes or no filters pick yes or no as their mode instead.
  if (booleanEventFilterTargets.includes(target)) return null;

  if (target === "audit_log_action") {
    return (
      <BaseMultiSelect
        field="event_filter_value"
        title="Actions"
        values={
          data.event_filter_value
            ?.split(",")
            .map((v) => v.trim())
            .filter(Boolean) ?? []
        }
        options={auditLogActions}
        updateValues={(v) =>
          updateData({ event_filter_value: v.join(",") || undefined })
        }
        errors={errors}
        searchable
      />
    );
  }

  if (target === "emoji") {
    return (
      <BaseEmojiPicker
        title="Emoji"
        field="event_filter_value"
        errors={errors}
        emoji={emojiFilterData(data.event_filter_value)}
        onChange={(emoji) =>
          updateData({
            event_filter_value: emoji ? emoji.id || emoji.name : undefined,
          })
        }
      />
    );
  }

  return (
    <BaseInput
      field="event_filter_value"
      title={target === "role_id" ? "Role ID" : "Filter Value"}
      description={
        target === "role_id"
          ? "The ID of the role, or the role mention."
          : data.event_filter_mode === "one_of" ||
            data.event_filter_mode === "not_one_of"
          ? "Separate the values with commas."
          : undefined
      }
      value={data.event_filter_value || ""}
      updateValue={(v) =>
        updateData({
          event_filter_value: v || undefined,
        })
      }
      errors={errors}
      placeholders="parameters"
    />
  );
}

// The value of an emoji filter is the ID of a custom emoji or a unicode emoji.
// Filters saved before the picker can also hold a custom emoji's mention.
function emojiFilterData(value?: string): EmojiData | undefined {
  if (!value) return undefined;
  const id =
    /^<a?:\w+:(\d+)>$/.exec(value)?.[1] ??
    (/^\d+$/.test(value) ? value : undefined);
  return id ? { id } : { name: value };
}

function EventFilterIgnoreCaseInput({ data, updateData, errors }: InputProps) {
  if (data.event_filter_target !== "message_content") return null;

  return (
    <BaseCheckbox
      field="event_filter_ignore_case"
      title="Ignore Case"
      description="If enabled, upper and lower case don't matter, so !help also matches !Help."
      value={!!data.event_filter_ignore_case}
      updateValue={(v) =>
        updateData({ event_filter_ignore_case: v || undefined })
      }
      errors={errors}
    />
  );
}

function LogLevelInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="log_level"
      title="Log Level"
      type="select"
      options={[
        { value: "debug", label: "Debug" },
        { value: "info", label: "Info" },
        { value: "warn", label: "Warn" },
        { value: "error", label: "Error" },
      ]}
      value={data.log_level || ""}
      updateValue={(v) => updateData({ log_level: v || undefined })}
      errors={errors}
    />
  );
}

function LogMessageInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="log_message"
      title="Log Message"
      value={data.log_message || ""}
      updateValue={(v) => updateData({ log_message: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function WebhookVerifyInput({ data, updateData, errors }: InputProps) {
  const preset = webhookVerifyPresets.find(
    (p) => p.value === data.webhook_verify_preset
  );

  return (
    <>
      <BaseInput
        type="select"
        field="webhook_verify_preset"
        title="Service"
        description="The service that sends the requests. Pick Custom for one that isn't listed."
        options={[...webhookVerifyPresets]}
        value={preset?.value ?? ""}
        updateValue={(v) =>
          updateData({ webhook_verify_preset: v || undefined })
        }
        errors={errors}
      />
      {preset && "key" in preset && (
        <BaseInput
          field="webhook_verify_key"
          title="Key"
          description={preset.key}
          value={data.webhook_verify_key || ""}
          updateValue={(v) =>
            updateData({ webhook_verify_key: v || undefined })
          }
          errors={errors}
          placeholders
        />
      )}
      {preset?.value === "custom" && (
        <>
          <BaseInput
            type="select"
            field="webhook_verify_algorithm"
            title="Algorithm"
            options={[...webhookVerifyAlgorithms]}
            value={data.webhook_verify_algorithm || "hmac_sha256"}
            updateValue={(v) =>
              updateData({ webhook_verify_algorithm: v || undefined })
            }
            errors={errors}
          />
          <BaseInput
            type="textarea"
            field="webhook_verify_message"
            title="Signed Message"
            description="What the service signs, usually the body, sometimes with a timestamp header before it. Spaces count."
            value={data.webhook_verify_message || ""}
            updateValue={(v) =>
              updateData({ webhook_verify_message: v || undefined })
            }
            errors={errors}
            placeholders
          />
          <BaseInput
            field="webhook_verify_signature"
            title="Signature"
            description="The header with the signature. Remove a prefix with an expression, like {{ trimPrefix(webhook.headers['x-signature'], 'sha256=') }}."
            value={data.webhook_verify_signature || ""}
            updateValue={(v) =>
              updateData({ webhook_verify_signature: v || undefined })
            }
            errors={errors}
            placeholders
          />
          <BaseInput
            type="select"
            field="webhook_verify_signature_encoding"
            title="Signature Encoding"
            options={[...webhookVerifyEncodings]}
            value={data.webhook_verify_signature_encoding || "hex"}
            updateValue={(v) =>
              updateData({ webhook_verify_signature_encoding: v || undefined })
            }
            errors={errors}
          />
        </>
      )}
    </>
  );
}

function WebhookResponseStatusInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="webhook_response_status"
      title="Status Code"
      description="The HTTP status code, like 200 for success or 401 when a signature is invalid. Must be between 200 and 299 or between 400 and 599."
      value={data.webhook_response_status || ""}
      updateValue={(v) =>
        updateData({ webhook_response_status: v || undefined })
      }
      errors={errors}
      placeholders
    />
  );
}

function WebhookResponseContentTypeInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      field="webhook_response_content_type"
      title="Content Type"
      type="select"
      options={[
        { value: "text/plain", label: "Text" },
        { value: "application/json", label: "JSON" },
      ]}
      value={data.webhook_response_content_type || "text/plain"}
      updateValue={(v) =>
        updateData({ webhook_response_content_type: v || undefined })
      }
      errors={errors}
    />
  );
}

function WebhookResponseBodyInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="webhook_response_body"
      title="Body"
      description="Respond within 10 seconds of the request, and before any wait that ends the run. Otherwise the request is answered with status 202."
      value={data.webhook_response_body || ""}
      updateValue={(v) => updateData({ webhook_response_body: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function AuditLogReasonInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="audit_log_reason"
      title="Audit Log Reason"
      description="This will appear in the Discord audit log."
      value={data.audit_log_reason || ""}
      updateValue={(v) => updateData({ audit_log_reason: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function HttpRequestDataInput({ data, updateData, errors }: InputProps) {
  // TODO: top level errors aren't displayed ...

  const updateField = useCallback(
    (newData: Partial<HTTPRequestData>) => {
      updateData({
        http_request_data: {
          ...data.http_request_data,
          ...newData,
        },
      });
    },
    [updateData, data]
  );

  const addHeader = useCallback(() => {
    updateData({
      http_request_data: {
        ...data.http_request_data,
        headers: [
          ...(data.http_request_data?.headers || []),
          { key: "", value: "" },
        ],
      },
    });
  }, [data, updateData]);

  const updateHeader = useCallback(
    (index: number, key: string, value: string) => {
      updateData({
        http_request_data: {
          ...data.http_request_data,
          headers: data.http_request_data?.headers?.map((h, i) =>
            i === index ? { key, value } : h
          ),
        },
      });
    },
    [data, updateData]
  );

  const removeHeader = useCallback(
    (index: number) => {
      updateData({
        http_request_data: {
          ...data.http_request_data,
          headers: data.http_request_data?.headers?.filter(
            (_, i) => i !== index
          ),
        },
      });
    },
    [data, updateData]
  );

  const credentialIntegration = urlCredentialIntegration(
    data.http_request_data?.url || ""
  );

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button className="w-full" variant="secondary">
          Configure Request
        </Button>
      </DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle>Configure HTTP Request</DialogTitle>
          <DialogDescription>
            Configure your HTTP request here to make an API call to a 3rd party
            service.
          </DialogDescription>
        </DialogHeader>
        <PreviewInert className="space-y-3">
          <BaseInput
            type="select"
            field="http_request_data.method"
            title="Method"
            description="The HTTP method to use for the request."
            options={[
              { value: "GET", label: "GET" },
              { value: "POST", label: "POST" },
              { value: "PUT", label: "PUT" },
              { value: "PATCH", label: "PATCH" },
              { value: "DELETE", label: "DELETE" },
            ]}
            value={data.http_request_data?.method || ""}
            updateValue={(v) => updateField({ method: v })}
            errors={errors}
          />
          <BaseInput
            type="text"
            field="http_request_data.url"
            title="URL"
            description="The URL to send the request to."
            value={data.http_request_data?.url || ""}
            updateValue={(v) => updateField({ url: v })}
            errors={errors}
            placeholders
          />
          {isDiscordApiUrl(data.http_request_data?.url || "") && (
            <div className="text-sm text-muted-foreground bg-muted rounded p-3">
              Use the Discord API Request block to call the Discord API. It
              sends your bot&apos;s token for you, so you don&apos;t have to
              paste it into a header.
            </div>
          )}
          {credentialIntegration && (
            <div className="text-sm text-muted-foreground bg-muted rounded p-3">
              Use the {credentialIntegration.name} blocks to call the{" "}
              {credentialIntegration.name} API. Enable{" "}
              {credentialIntegration.name} under Integrations and they send your{" "}
              {credentialIntegration.auth.label} for you, so you don&apos;t have
              to paste it into a header.
            </div>
          )}
          <div>
            <div className="font-medium text-foreground mb-1">Headers</div>
            <div className="text-muted-foreground text-sm mb-2">
              The headers to send with the request.
            </div>
            <div className="flex flex-col gap-3">
              {data.http_request_data?.headers?.map((h, i) => (
                <div className="flex gap-2" key={i}>
                  <div className="grid grid-cols-2 gap-2 flex-auto">
                    <Input
                      type="text"
                      placeholder="Key"
                      value={h.key || ""}
                      onChange={(e) => updateHeader(i, e.target.value, h.value)}
                    />
                    <BaseInput
                      type="text"
                      field={`http_request_data.headers.${i}.value`}
                      placeholder="Value"
                      value={h.value || ""}
                      updateValue={(v) => updateHeader(i, h.key, v)}
                      errors={errors}
                      placeholders
                    />
                  </div>
                  <Button
                    variant="outline"
                    size="icon"
                    className="flex-none"
                    onClick={() => removeHeader(i)}
                  >
                    <MinusIcon className="h-5 w-5" />
                  </Button>
                </div>
              ))}

              <div className="flex">
                <Button variant="outline" size="icon" onClick={addHeader}>
                  <PlusIcon className="h-5 w-5" />
                </Button>
              </div>
            </div>
          </div>
          <div>
            <div className="flex items-center justify-between mb-2">
              <div className="font-medium text-foreground">JSON Body</div>
              <Switch
                checked={!!data.http_request_data?.body_json}
                onCheckedChange={(checked) =>
                  updateField({
                    body_json: checked ? {} : undefined,
                  })
                }
              />
            </div>
            {!!data.http_request_data?.body_json && (
              <JsonEditor
                src={data.http_request_data?.body_json || {}}
                onChange={(v) => updateField({ body_json: v })}
              />
            )}
          </div>
        </PreviewInert>
      </DialogContent>
    </Dialog>
  );
}

const discordApiOperationItems = discordApiOperations.map((o) => ({
  id: o.id,
  name: discordApiOperationLabel(o.id),
  description: `${o.method} ${o.path}`,
}));

function BlockFieldInput({
  id,
  type,
  name,
  data,
  updateData,
  errors,
}: InputProps) {
  const definition = getBlockDefinition(type);
  const field = definition?.fields?.find((f) => f.name === name);
  if (!field) return null;

  const key = field.name;
  const value = String(data[key] ?? "");
  // Fields without a schema are described, a test checks it.
  const title = field.label ?? field.name;
  const description = field.description ?? "";

  function setValue(value: string) {
    updateData({ [key]: value || undefined });
  }

  if (field.widget === "permissions") {
    return (
      <BasePermissionInput
        field={key}
        title={title}
        description={description}
        value={value || "0"}
        updateValue={(v) => setValue(v === "0" ? "" : v)}
        errors={errors}
      />
    );
  }
  if (field.type === "color") {
    return (
      <BaseColorInput
        // Text mode is per block.
        key={id}
        field={key}
        title={title}
        description={description}
        value={value}
        updateValue={setValue}
        errors={errors}
      />
    );
  }
  // Edits keep what's left out, so they need "No" besides leaving it empty.
  if (
    field.type === "boolean" &&
    definition?.run.kind === "request" &&
    definition.run.method === "PATCH"
  ) {
    return (
      <BaseInput
        type="select"
        field={key}
        title={title}
        description={description}
        // Select items can't have an empty value, so "keep" stands for it.
        options={[
          { value: "keep", label: "Keep current" },
          { value: "true", label: "Yes" },
          { value: "false", label: "No" },
        ]}
        value={value || "keep"}
        updateValue={(v) => setValue(v === "keep" ? "" : v)}
        errors={errors}
      />
    );
  }
  // Stored as text like the other fields. Off leaves the setting out, which
  // the APIs these blocks call treat as false.
  if (field.type === "boolean") {
    return (
      <BaseCheckbox
        field={key}
        title={title}
        description={description}
        value={value === "true"}
        updateValue={(v) => setValue(v ? "true" : "")}
        errors={errors}
      />
    );
  }
  return (
    <BaseInput
      type="text"
      field={key}
      title={title}
      description={description}
      value={value}
      updateValue={setValue}
      errors={errors}
      placeholders
    />
  );
}

// Webhooks with a token in the URL don't need the bot's token, and the Discord
// API Request block can't call them.
function isDiscordApiUrl(url: string) {
  return (
    /^\s*(https?:\/\/)?((ptb|canary)\.)?discord(app)?\.com(\/|$)/i.test(url) &&
    !/\/webhooks\//i.test(url)
  );
}

// The integration with a credential an HTTP block's URL goes to, if any.
function urlCredentialIntegration(url: string) {
  let host: string;
  try {
    const trimmed = url.trim();
    host = new URL(
      /^https?:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`
    ).host;
  } catch {
    return;
  }
  return integrations
    .filter(needsCredential)
    .find(
      (i) =>
        (i.base_url && new URL(i.base_url).host === host) ||
        i.other_hosts?.includes(host)
    );
}

function DiscordApiRequestDataInput({ data, updateData, errors }: InputProps) {
  const request = data.discord_api_request_data;
  const op = getDiscordApiOperation(request?.operation);

  function updateRequest(newData: Partial<DiscordAPIRequestData>) {
    updateData({ discord_api_request_data: { ...request, ...newData } });
  }

  function selectOperation(id: string) {
    const newOp = getDiscordApiOperation(id);
    if (!newOp) return;

    // Keeps the values of parameters both endpoints have, like channel_id.
    const pathValues = new Map(
      request?.path_params?.map((p) => [p.key, p.value])
    );
    updateRequest({
      operation: id,
      path_params: newOp.path_params.map((p) => ({
        key: p.name,
        value: pathValues.get(p.name) ?? "",
      })),
      query: request?.query?.filter((q) =>
        newOp.query_params.some((p) => p.name === q.key)
      ),
      body_json: newOp.has_body ? request?.body_json : undefined,
    });
  }

  function setParam(
    field: "path_params" | "query",
    key: string,
    value: string
  ) {
    const params = request?.[field] ?? [];
    updateRequest({
      [field]: params.some((p) => p.key === key)
        ? params.map((p) => (p.key === key ? { key, value } : p))
        : [...params, { key, value }],
    });
  }

  const unusedQueryParams =
    op?.query_params.filter(
      (p) => !request?.query?.some((q) => q.key === p.name)
    ) ?? [];

  const operationError = errors["discord_api_request_data.operation"];
  const bodyError = errors["discord_api_request_data.body_json"];

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button className="w-full" variant="secondary">
          Configure Request
        </Button>
      </DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle>Configure Discord API Request</DialogTitle>
          <DialogDescription>
            Call an endpoint of the Discord API as your bot. Kite adds the
            bot&apos;s token, so never paste it into a flow. See{" "}
            <Link
              href="https://discord.com/developers/docs/reference"
              target="_blank"
              className="text-primary hover:underline"
            >
              Discord&apos;s API docs
            </Link>{" "}
            for what each endpoint takes and returns, and{" "}
            <Link
              href={nodeTypeDocsPage("action_discord_api_request")!}
              target="_blank"
              className="text-primary hover:underline"
            >
              how to use this block
            </Link>
            .
          </DialogDescription>
        </DialogHeader>
        <PreviewInert className="space-y-3">
          <div>
            <div className="font-medium text-foreground mb-2">Endpoint</div>
            <EntitySelect
              items={discordApiOperationItems}
              value={request?.operation ?? null}
              onChange={(id) => id && selectOperation(id)}
              placeholder="Select an endpoint"
              searchPlaceholder="Search endpoints..."
              emptyText="No endpoint found."
              wide
              modal
            />
            {op && (
              <div className="text-muted-foreground text-sm font-mono mt-2 break-all">
                {op.method} {op.path}
              </div>
            )}
            {op?.interaction && (
              <div className="text-muted-foreground text-sm mt-2">
                Kite fills in the app&apos;s ID and the token of the interaction
                that started the flow, like a command or button click.
              </div>
            )}
            {operationError && (
              <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
                <CircleAlertIcon className="h-5 w-5 flex-none" />
                <div>{operationError}</div>
              </div>
            )}
          </div>
          {op?.path_params.map((p) => (
            <BaseInput
              key={p.name}
              type="text"
              field={`discord_api_request_data.path_params.${p.name}`}
              title={p.name}
              description={p.type === "snowflake" ? "An ID" : undefined}
              value={
                request?.path_params?.find((v) => v.key === p.name)?.value || ""
              }
              updateValue={(v) => setParam("path_params", p.name, v)}
              errors={errors}
              placeholders
            />
          ))}
          {request?.query?.map((q) => (
            <div className="flex gap-2 items-end" key={q.key}>
              <BaseInput
                type="text"
                field={`discord_api_request_data.query.${q.key}`}
                title={q.key}
                description="Query parameter"
                value={q.value}
                updateValue={(v) => setParam("query", q.key, v)}
                errors={errors}
                placeholders
              />
              <Button
                variant="outline"
                size="icon"
                className="flex-none"
                onClick={() =>
                  updateRequest({
                    query: request.query?.filter((v) => v.key !== q.key),
                  })
                }
              >
                <MinusIcon className="h-5 w-5" />
              </Button>
            </div>
          ))}
          {unusedQueryParams.length > 0 && (
            <Select
              value=""
              onValueChange={(key) =>
                updateRequest({
                  query: [...(request?.query ?? []), { key, value: "" }],
                })
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Add query parameter" />
              </SelectTrigger>
              <SelectContent>
                {unusedQueryParams.map((p) => (
                  <SelectItem key={p.name} value={p.name}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
          {op?.has_body && (
            <div>
              <div className="flex items-center justify-between mb-2">
                <div className="font-medium text-foreground">JSON Body</div>
                <Switch
                  checked={!!request?.body_json}
                  onCheckedChange={(checked) =>
                    updateRequest({ body_json: checked ? {} : undefined })
                  }
                />
              </div>
              {!!request?.body_json && (
                <FlowJsonInput
                  value={request.body_json}
                  // Some endpoints take a list, which the generated type
                  // doesn't allow for.
                  onChange={(v) =>
                    updateRequest({
                      body_json: v as DiscordAPIRequestData["body_json"],
                    })
                  }
                />
              )}
              {bodyError && (
                <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
                  <CircleAlertIcon className="h-5 w-5 flex-none" />
                  <div>{bodyError}</div>
                </div>
              )}
            </div>
          )}
        </PreviewInert>
      </DialogContent>
    </Dialog>
  );
}

function AiModelInput({
  data,
  updateData,
  errors,
}: Pick<InputProps, "data" | "updateData" | "errors">) {
  return (
    <BaseInput
      type="select"
      field="ai_chat_completion_data.model"
      title="Model"
      description="How capable the AI is. More capable models cost more credits."
      options={aiModelTiers.map((t) => ({
        value: t.value,
        label: `${t.label} (${t.model})`,
      }))}
      value={getAiModelTier(data.ai_chat_completion_data?.model)?.value ?? ""}
      updateValue={(v) =>
        updateData({
          ai_chat_completion_data: {
            ...data.ai_chat_completion_data,
            model: v || undefined,
          },
        })
      }
      errors={errors}
    />
  );
}

function AiChatCompletionDataInput({ data, updateData, errors }: InputProps) {
  // TODO: top level errors aren't displayed ...

  return (
    <>
      <AiModelInput data={data} updateData={updateData} errors={errors} />
      <BaseInput
        type="textarea"
        field="ai_chat_completion_data.system_prompt"
        title="System Prompt"
        description="Tell the AI how to behave. (optional)"
        value={data.ai_chat_completion_data?.system_prompt || ""}
        updateValue={(v) =>
          updateData({
            ai_chat_completion_data: {
              ...data.ai_chat_completion_data,
              system_prompt: v || undefined,
            },
          })
        }
        errors={errors}
        placeholders
      />
      <BaseInput
        type="textarea"
        field="ai_chat_completion_data.prompt"
        title="Prompt"
        description="The message to send to the AI which it will respond to."
        value={data.ai_chat_completion_data?.prompt || ""}
        updateValue={(v) =>
          updateData({
            ai_chat_completion_data: {
              ...data.ai_chat_completion_data,
              prompt: v || undefined,
            },
          })
        }
        errors={errors}
        placeholders
      />
    </>
  );
}

function AiWebSearchDataInput({ data, updateData, errors }: InputProps) {
  // TODO: top level errors aren't displayed ...

  return (
    <>
      <AiModelInput data={data} updateData={updateData} errors={errors} />
      <BaseInput
        type="textarea"
        field="ai_chat_completion_data.prompt"
        title="Query"
        description="The search query to send to the AI."
        value={data.ai_chat_completion_data?.prompt || ""}
        updateValue={(v) =>
          updateData({
            ai_chat_completion_data: {
              ...data.ai_chat_completion_data,
              prompt: v || undefined,
            },
          })
        }
        errors={errors}
        placeholders
      />
    </>
  );
}

// The settings AI Choice, Check and Score share.
function AiDecisionInput({
  data,
  updateData,
  errors,
  field,
  title,
  description,
  textarea,
}: Pick<InputProps, "data" | "updateData" | "errors"> & {
  field:
    | "question"
    | "input"
    | "min_confidence"
    | "yes_threshold"
    | "no_threshold";
  title: string;
  description: string;
  textarea?: boolean;
}) {
  return (
    <BaseInput
      type={textarea ? "textarea" : "text"}
      field={`ai_decision_data.${field}`}
      title={title}
      description={description}
      value={data.ai_decision_data?.[field] || ""}
      updateValue={(v) =>
        updateData({
          ai_decision_data: {
            ...data.ai_decision_data,
            [field]: v || undefined,
          },
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function AiDecisionQuestionInputs(
  props: Pick<InputProps, "data" | "updateData" | "errors">
) {
  return (
    <>
      <AiDecisionInput
        {...props}
        textarea
        field="question"
        title="Question"
        description="What the AI decides, like 'Is this message spam?'."
      />
      <AiDecisionInput
        {...props}
        textarea
        field="input"
        title="Input"
        description="What the AI decides about, like the content of a message."
      />
    </>
  );
}

function AiChoiceDataInput(props: InputProps) {
  return (
    <>
      <AiDecisionQuestionInputs {...props} />
      <AiDecisionInput
        {...props}
        field="min_confidence"
        title="Min Confidence"
        description="How confident the AI has to be, from 0 to 1. Below it, the else branch runs instead. (optional)"
      />
    </>
  );
}

function AiCheckDataInput(props: InputProps) {
  return (
    <>
      <AiDecisionQuestionInputs {...props} />
      <AiDecisionInput
        {...props}
        field="yes_threshold"
        title="Yes Threshold"
        description="How likely yes has to be, from 0 to 1, for the yes branch to run. Defaults to the no threshold, or 0.5."
      />
      <AiDecisionInput
        {...props}
        field="no_threshold"
        title="No Threshold"
        description="Below this likelihood of yes, the no branch runs. In between the two thresholds, the unsure branch runs. Defaults to the yes threshold, or 0.5."
      />
    </>
  );
}

function AiScoreDataInput(props: InputProps) {
  const { data, updateData, errors } = props;
  const levels = data.ai_decision_data?.levels || [];

  function updateLevels(levels: string[]) {
    updateData({
      ai_decision_data: { ...data.ai_decision_data, levels },
    });
  }

  return (
    <>
      <AiDecisionQuestionInputs {...props} />
      <div>
        <div className="font-medium text-foreground mb-2">Levels</div>
        <div className="text-muted-foreground text-sm mb-2">
          What each level of the scale means, lowest first. The result is the
          level, from 1 to the number of levels.
        </div>
        <div className="flex flex-col gap-3 mb-3">
          {levels.map((level, i) => (
            <div className="flex gap-2" key={i}>
              <BaseInput
                field={`ai_decision_data.levels.${i}`}
                placeholder={`Level ${i + 1}`}
                value={level}
                updateValue={(v) =>
                  updateLevels(levels.map((l, j) => (j === i ? v : l)))
                }
                errors={errors}
              />
              <Button
                variant="outline"
                size="icon"
                className="flex-none"
                aria-label="Remove level"
                onClick={() => updateLevels(levels.filter((_, j) => j !== i))}
              >
                <MinusIcon className="h-5 w-5" />
              </Button>
            </div>
          ))}
        </div>
        <Button
          variant="outline"
          size="icon"
          aria-label="Add level"
          disabled={levels.length >= aiScoreMaxLevels}
          onClick={() => updateLevels([...levels, ""])}
        >
          <PlusIcon className="h-5 w-5" />
        </Button>
        {errors["ai_decision_data.levels"] && (
          <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
            <CircleAlertIcon className="h-5 w-5 flex-none" />
            <div>{errors["ai_decision_data.levels"]}</div>
          </div>
        )}
      </div>
    </>
  );
}

function AiChoiceNameInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="ai_choice_name"
      title="Name"
      description="The AI Choice block returns the name of the option it picked."
      value={data.ai_choice_name || ""}
      updateValue={(v) => updateData({ ai_choice_name: v || undefined })}
      errors={errors}
    />
  );
}

function AiChoiceDescriptionInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="ai_choice_description"
      title="Description"
      description="When to pick this option, if the name doesn't say. (optional)"
      value={data.ai_choice_description || ""}
      updateValue={(v) => updateData({ ai_choice_description: v || undefined })}
      errors={errors}
    />
  );
}

function ExpressionInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="expression"
      title="Expression"
      description="The expression to evaluate without the double curly brackets"
      value={data.expression || ""}
      updateValue={(v) =>
        updateData({
          expression: v || undefined,
        })
      }
      errors={errors}
      placeholders
      disablePlaceholderBrackets
    />
  );
}

function RandomInput({ data, updateData, errors }: InputProps) {
  const mode = data.random_mode || "number";

  return (
    <>
      <BaseInput
        type="select"
        field="random_mode"
        title="Mode"
        options={[...randomModes]}
        value={mode}
        updateValue={(v) => updateData({ random_mode: v || undefined })}
        errors={errors}
      />
      {mode === "item" || mode === "items" ? (
        <>
          <ListInput
            field="random_items"
            title="Items"
            noun="item"
            value={data.random_items || ""}
            updateValue={(v) => updateData({ random_items: v || undefined })}
            errors={errors}
          />
          {mode === "items" && (
            <BaseInput
              field="random_count"
              title="Count"
              description="How many different items to pick, from 1 to 100."
              value={data.random_count || ""}
              updateValue={(v) => updateData({ random_count: v || undefined })}
              errors={errors}
            />
          )}
        </>
      ) : (
        <>
          <BaseInput
            field="random_min"
            title="Min"
            value={data.random_min || ""}
            updateValue={(v) => updateData({ random_min: v || undefined })}
            errors={errors}
          />
          <BaseInput
            field="random_max"
            title="Max"
            value={data.random_max || ""}
            updateValue={(v) => updateData({ random_max: v || undefined })}
            errors={errors}
          />
        </>
      )}
    </>
  );
}

function LookupValueInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="lookup_value"
      title="Value"
      description="The value to look up in the table, like {{message.content}}."
      value={data.lookup_value || ""}
      updateValue={(v) => updateData({ lookup_value: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function LookupModeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="select"
      field="lookup_mode"
      title="Match Mode"
      description="How the key of a row has to match the value."
      options={[...lookupModes]}
      value={data.lookup_mode || "equal"}
      updateValue={(v) => updateData({ lookup_mode: v || undefined })}
      errors={errors}
    />
  );
}

function LookupIgnoreCaseInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseCheckbox
      field="lookup_ignore_case"
      title="Ignore Case"
      description="If enabled, upper and lower case don't matter, so hello also matches Hello."
      value={!!data.lookup_ignore_case}
      updateValue={(v) => updateData({ lookup_ignore_case: v || undefined })}
      errors={errors}
    />
  );
}

function LookupRowsInput({ data, updateData, errors }: InputProps) {
  const rows = data.lookup_rows || [];
  const [pasted, setPasted] = useState("");
  const [pasteError, setPasteError] = useState("");

  const setRows = (newRows: LookupRow[]) =>
    updateData({ lookup_rows: newRows.length ? newRows : undefined });

  // Rows are memoized with stable callbacks, so typing in a cell only
  // re-renders its row, which matters with hundreds of them.
  const latest = useRef({ rows, setRows });
  latest.current = { rows, setRows };
  const updateRow = useCallback(
    (index: number, newData: Partial<LookupRow>) => {
      const { rows, setRows } = latest.current;
      setRows(rows.map((r, i) => (i === index ? { ...r, ...newData } : r)));
    },
    []
  );
  const removeRow = useCallback((index: number) => {
    const { rows, setRows } = latest.current;
    setRows(rows.filter((_, i) => i !== index));
  }, []);

  function addPasted(replace: boolean) {
    const pastedRows = parseLookupLines(pasted);
    if (!pastedRows.length) return;
    const newRows = replace ? pastedRows : [...rows, ...pastedRows];
    if (newRows.length > lookupMaxRows) {
      setPasteError(
        `The table can have up to ${lookupMaxRows} rows, this would make it ${newRows.length}.`
      );
      return;
    }
    setRows(newRows);
    setPasted("");
    setPasteError("");
  }

  const rowError = (i: number) =>
    errors[`lookup_rows.${i}.key`] || errors[`lookup_rows.${i}.value`];
  const rowErrors = rows.filter((_, i) => rowError(i)).length;
  const error =
    errors["lookup_rows"] ||
    (rowErrors > 0 &&
      `${rowErrors} ${rowErrors === 1 ? "row has" : "rows have"} an error`);

  return (
    <div>
      <div className="font-medium text-foreground mb-1">Table</div>
      <div className="text-muted-foreground text-sm mb-2">
        The block returns the value of the first row whose key matches.
      </div>
      <Dialog>
        <DialogTrigger asChild>
          <Button className="w-full" variant="secondary">
            Edit Table ({rows.length} {rows.length === 1 ? "row" : "rows"})
          </Button>
        </DialogTrigger>
        <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Table</DialogTitle>
            <DialogDescription>
              Up to {lookupMaxRows} rows. Keys and values can use placeholders,
              like {"{{user.id}}"}.
            </DialogDescription>
          </DialogHeader>
          <PreviewInert className="space-y-5">
            <div>
              {rows.length > 0 && (
                <div className="flex gap-2 mb-2 text-sm font-medium text-foreground">
                  <div className="grid grid-cols-2 gap-2 flex-auto">
                    <div>Key</div>
                    <div>Value</div>
                  </div>
                  {/* Room for the remove buttons. */}
                  <div className="w-10 flex-none" />
                </div>
              )}
              <div className="flex flex-col gap-2">
                {rows.map((row, i) => (
                  <LookupRowInput
                    key={i}
                    index={i}
                    row={row}
                    error={rowError(i)}
                    onChange={updateRow}
                    onRemove={removeRow}
                  />
                ))}
                <div className="flex">
                  <Button
                    variant="outline"
                    size="icon"
                    aria-label="Add row"
                    onClick={() => setRows([...rows, { key: "", value: "" }])}
                    disabled={rows.length >= lookupMaxRows}
                  >
                    <PlusIcon className="h-5 w-5" />
                  </Button>
                </div>
              </div>
            </div>
            <div>
              <div className="font-medium text-foreground mb-1">Paste Rows</div>
              <div className="text-muted-foreground text-sm mb-2">
                One row per line, as key = value, or copy two columns from a
                spreadsheet.
              </div>
              <Textarea
                value={pasted}
                onChange={(e) => setPasted(e.target.value)}
                placeholder={"hello = Hi there!\nbye = See you later!"}
                className="min-h-24"
              />
              <div className="flex gap-2 mt-2">
                <Button
                  variant="secondary"
                  onClick={() => addPasted(false)}
                  disabled={!pasted.trim()}
                >
                  Add Rows
                </Button>
                <ConfirmDialog
                  title="Replace all rows?"
                  description="This removes the rows of the table and adds the pasted ones instead."
                  onConfirm={() => addPasted(true)}
                >
                  <Button variant="outline" disabled={!pasted.trim()}>
                    Replace All Rows
                  </Button>
                </ConfirmDialog>
              </div>
              {pasteError && (
                <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
                  <CircleAlertIcon className="h-5 w-5 flex-none" />
                  <div>{pasteError}</div>
                </div>
              )}
            </div>
          </PreviewInert>
        </DialogContent>
      </Dialog>
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}

const LookupRowInput = memo(function LookupRowInput({
  index,
  row,
  error,
  onChange,
  onRemove,
}: {
  index: number;
  row: LookupRow;
  error?: string;
  onChange: (index: number, newData: Partial<LookupRow>) => void;
  onRemove: (index: number) => void;
}) {
  return (
    <div>
      <div className="flex gap-2">
        <div className="grid grid-cols-2 gap-2 flex-auto">
          <Input
            type="text"
            placeholder="Key"
            value={row.key}
            onChange={(e) => onChange(index, { key: e.target.value })}
          />
          <Input
            type="text"
            placeholder="Value"
            value={row.value}
            onChange={(e) => onChange(index, { value: e.target.value })}
          />
        </div>
        <Button
          variant="outline"
          size="icon"
          className="flex-none"
          aria-label="Remove row"
          onClick={() => onRemove(index)}
        >
          <MinusIcon className="h-5 w-5" />
        </Button>
      </div>
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
});

function LookupDefaultInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="lookup_default"
      title="Default Value"
      description="The result when no row matches. Leave empty to return nothing, which a Comparison Condition can check for."
      value={data.lookup_default || ""}
      updateValue={(v) => updateData({ lookup_default: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function UserTargetInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="user_target"
      title="Target User"
      value={data.user_target || ""}
      updateValue={(v) => updateData({ user_target: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function GuildTargetInput({ type, data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="guild_target"
      title="Target Guild"
      description={
        type === "action_guild_get"
          ? undefined
          : "Leave empty to use the server the flow runs in. Required in scheduled and webhook event listeners."
      }
      value={data.guild_target || ""}
      updateValue={(v) => updateData({ guild_target: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function MemberBanDeleteMessageDurationInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      type="text"
      field="member_ban_delete_message_duration_seconds"
      title="Delete Message Duration"
      description="Number of seconds to delete messages for."
      value={data.member_ban_delete_message_duration_seconds || ""}
      updateValue={(v) =>
        updateData({
          member_ban_delete_message_duration_seconds: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function MemberTimeoutDurationInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="member_timeout_duration_seconds"
      title="Timeout Duration"
      description="Number of seconds to timeout the user for."
      value={data.member_timeout_duration_seconds || ""}
      updateValue={(v) =>
        updateData({ member_timeout_duration_seconds: v || undefined })
      }
      errors={errors}
      placeholders
    />
  );
}

function MemberNickInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="member_data"
      title="Member Nickname"
      value={data.member_data?.nick || ""}
      updateValue={(v) =>
        updateData({ member_data: v ? { nick: v } : undefined })
      }
      errors={errors}
      placeholders
    />
  );
}

function RobloxUserTargetInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="roblox_user_target"
      title="Roblox User Target"
      value={data.roblox_user_target || ""}
      updateValue={(v) => updateData({ roblox_user_target: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function RobloxLookupModeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="select"
      field="roblox_lookup_mode"
      title="Roblox Lookup Mode"
      value={data.roblox_lookup_mode || ""}
      options={[
        { label: "ID", value: "id" },
        { label: "Username", value: "username" },
      ]}
      updateValue={(v) => updateData({ roblox_lookup_mode: v || undefined })}
      errors={errors}
    />
  );
}

function MessageTemplateInput({ data, updateData, errors }: InputProps) {
  const messages = useFlowMessages();
  const readOnly = useContext(FlowReadOnlyContext);
  const template = data.message_template_id
    ? messages?.find((m) => m.id === data.message_template_id)
    : undefined;

  const appId = useAppId();

  return (
    <>
      <PreviewInert className="flex space-x-2 items-end">
        <BaseInput
          type="select"
          field="message_template_id"
          title="Message Template"
          description="Select a message template to use for the response."
          options={messages?.map((m) => ({
            value: m!.id,
            label: m!.name,
          }))}
          value={data.message_template_id || ""}
          updateValue={(v) =>
            updateData({ message_template_id: v || undefined })
          }
          errors={errors}
          clearable
        />
        {!readOnly &&
          (data.message_template_id ? (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="icon" asChild>
                  <Link
                    href={{
                      pathname: "/apps/[appId]/messages/[messageId]",
                      query: {
                        appId: appId,
                        messageId: data.message_template_id,
                      },
                    }}
                    target="_blank"
                  >
                    <PencilIcon className="h-5 w-5" />
                  </Link>
                </Button>
              </TooltipTrigger>
              <TooltipContent>Edit message template</TooltipContent>
            </Tooltip>
          ) : (
            <MessageCreateDialog
              onMessageCreated={(v) => updateData({ message_template_id: v })}
            >
              <Button variant="outline" size="icon">
                <PlusIcon className="h-5 w-5" />
              </Button>
            </MessageCreateDialog>
          ))}
      </PreviewInert>
      {readOnly && template && (
        <FlowMessageViewDialog name={template.name} message={template.data} />
      )}
    </>
  );
}

function MessageDataInput({ data, updateData, errors }: InputProps) {
  const readOnly = useContext(FlowReadOnlyContext);

  if (data.message_template_id) {
    return null;
  }

  // Components v2 messages have no content, their text lives in the components.
  const componentsV2 = hasComponentsV2Flag(data.message_data?.flags);

  return (
    <>
      {!componentsV2 && (
        <PreviewInert>
          <BaseInput
            type="textarea"
            field="message_data"
            title="Text"
            description="Edit the message content here or click below to have a full message editor with support for embeds and components."
            value={data.message_data?.content || ""}
            updateValue={(v) =>
              updateData({
                message_data: {
                  ...data.message_data,
                  content: v || undefined,
                },
              })
            }
            errors={errors}
            placeholders
          />
        </PreviewInert>
      )}

      {readOnly ? (
        <FlowMessageViewDialog message={data.message_data} />
      ) : (
        <MessageEditorDialog
          onClose={(v) => updateData({ message_data: v })}
          message={data.message_data || {}}
        >
          <Button className="w-full">Edit Message</Button>
        </MessageEditorDialog>
      )}

      <AllowedMentionsDialog
        data={data}
        updateData={updateData}
        errors={errors}
      />
    </>
  );
}

const mentionKinds = [
  {
    value: "users",
    title: "Users",
    description: "Mentioned users get a notification.",
  },
  {
    value: "roles",
    title: "Roles",
    description: "Members of mentioned roles get a notification.",
  },
  {
    value: "everyone",
    title: "@everyone and @here",
    description: "Everyone in the channel gets a notification.",
  },
];

function AllowedMentionsDialog({
  data,
  updateData,
  errors,
}: Omit<InputProps, "id" | "type">) {
  const mentions = data.message_data?.allowed_mentions;
  // Unset only pings users.
  const parse = mentions ? mentions.parse || [] : ["users"];
  const repliedUser = !!mentions?.replied_user;
  // Shown below the button, as the dialog is usually closed.
  const error = Object.entries(errors).find(([k]) =>
    k.startsWith("message_data.allowed_mentions")
  )?.[1];

  function update(newParse: string[], newRepliedUser: boolean) {
    updateData({
      message_data: {
        ...data.message_data,
        allowed_mentions: {
          parse: mentionKinds
            .map((k) => k.value)
            .filter((k) => newParse.includes(k)),
          replied_user: newRepliedUser || undefined,
        },
      },
    });
  }

  return (
    <Dialog>
      <div>
        <DialogTrigger asChild>
          <Button className="w-full" variant="secondary">
            Configure Pings
          </Button>
        </DialogTrigger>
        {error && (
          <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
            <CircleAlertIcon className="h-5 w-5 flex-none" />
            <div>{error}</div>
          </div>
        )}
      </div>
      <DialogContent className="overflow-y-auto max-h-[90dvh]">
        <DialogHeader>
          <DialogTitle>Configure Pings</DialogTitle>
          <DialogDescription>
            Choose which mentions in the message notify people. Mentions without
            a ping are still shown.
          </DialogDescription>
        </DialogHeader>
        <PreviewInert className="space-y-5">
          {mentionKinds.map((k) => (
            <BaseCheckbox
              key={k.value}
              field="message_data.allowed_mentions"
              title={k.title}
              description={k.description}
              value={parse.includes(k.value)}
              updateValue={(v) =>
                update(
                  v ? [...parse, k.value] : parse.filter((p) => p !== k.value),
                  repliedUser
                )
              }
              errors={{}}
            />
          ))}
          {data.reply_target && (
            <BaseCheckbox
              field="message_data.allowed_mentions.replied_user"
              title="Reply Author"
              description="The author of the message this replies to gets a notification."
              value={repliedUser}
              updateValue={(v) => update(parse, v)}
              errors={{}}
            />
          )}
        </PreviewInert>
      </DialogContent>
    </Dialog>
  );
}

function MessageTargetInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="message_target"
      title="Target Message"
      value={data.message_target || ""}
      updateValue={(v) => updateData({ message_target: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function ReplyTargetInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="reply_target"
      title="Reply To"
      description="The message to reply to, in the same channel. Leave blank to not reply."
      value={data.reply_target || ""}
      updateValue={(v) => updateData({ reply_target: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function ResponseTargetInput({ data, updateData, errors }: InputProps) {
  // Picking Other clears the target, so it has to be remembered until an ID
  // is entered.
  const [pickedOther, setPickedOther] = useState(false);
  const other =
    pickedOther ||
    (!!data.message_target && data.message_target !== "@original");

  return (
    <>
      <BaseInput
        type="select"
        field="message_target"
        title="Target Response"
        value={other ? "other" : data.message_target || ""}
        options={[
          { label: "Original Response", value: "@original" },
          { label: "Other Response", value: "other" },
        ]}
        updateValue={(v) => {
          setPickedOther(v === "other");
          updateData({ message_target: v === "other" ? undefined : v });
        }}
        errors={other ? {} : errors}
      />
      {other && (
        <BaseInput
          type="text"
          field="message_target"
          title="Response ID"
          description="ID of a response this interaction created after the first one, like the result of another Create Response Message block."
          value={data.message_target || ""}
          updateValue={(v) => updateData({ message_target: v || undefined })}
          errors={errors}
          placeholders
        />
      )}
    </>
  );
}

function MessageEphemeralInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseCheckbox
      field="message_ephemeral"
      title="Public Response"
      description="If enabled, the response will be visible to everyone. If disabled, the response will only be visible to the user who triggered the flow."
      value={!data.message_ephemeral}
      updateValue={(v) => updateData({ message_ephemeral: !v || undefined })}
      errors={errors}
    />
  );
}

function EmojiDataInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseEmojiPicker
      title="Emoji"
      field="emoji_data"
      errors={errors}
      emoji={data.emoji_data}
      onChange={(emoji) =>
        updateData({
          emoji_data: emoji,
        })
      }
    />
  );
}

const pollMaxAnswers = 10;

function PollDataInput({ data, updateData, errors }: InputProps) {
  const updateField = useCallback(
    (newData: Partial<PollData>) => {
      updateData({ poll_data: { ...data.poll_data, ...newData } });
    },
    [updateData, data]
  );

  const answers = useMemo(
    () => data.poll_data?.answers || [],
    [data.poll_data?.answers]
  );

  const addAnswer = useCallback(() => {
    if (answers.length >= pollMaxAnswers) return;
    updateField({ answers: [...answers, { text: "" }] });
  }, [updateField, answers]);

  const updateAnswer = useCallback(
    (index: number, newData: Partial<PollAnswerData>) => {
      updateField({
        answers: answers.map((a, i) =>
          i === index ? { ...a, ...newData } : a
        ),
      });
    },
    [updateField, answers]
  );

  const removeAnswer = useCallback(
    (index: number) => {
      updateField({ answers: answers.filter((_, i) => i !== index) });
    },
    [updateField, answers]
  );

  const answersError = errors["poll_data.answers"];

  return (
    <>
      <BaseInput
        type="textarea"
        field="poll_data.question"
        title="Question"
        description="The question shown at the top of the poll. Up to 300 characters."
        value={data.poll_data?.question || ""}
        updateValue={(v) => updateField({ question: v || undefined })}
        errors={errors}
        placeholders
      />
      <div>
        <div className="font-medium text-foreground mb-1">Answers</div>
        <div className="text-muted-foreground text-sm mb-2">
          Up to {pollMaxAnswers} answers of 55 characters each. Answers that are
          empty after placeholders are filled in are skipped.
        </div>
        <div className="flex flex-col gap-3">
          {answers.map((answer, i) => {
            const error = errors[`poll_data.answers.${i}.text`];

            return (
              <div key={i}>
                <div className="flex gap-2">
                  <EmojiPicker
                    onEmojiSelect={(emoji) =>
                      updateAnswer(i, {
                        emoji: emoji.native
                          ? { name: emoji.name }
                          : { id: emoji.id, name: emoji.name },
                      })
                    }
                  >
                    <Button size="icon" variant="outline" className="flex-none">
                      {answer.emoji?.id ? (
                        <img
                          src={discordEmojiUrl(answer.emoji.id)}
                          alt=""
                          className="h-6 w-6"
                        />
                      ) : answer.emoji ? (
                        <Twemoji options={{ className: "h-6 w-6" }}>
                          {answer.emoji.name}
                        </Twemoji>
                      ) : (
                        <SmileIcon className="h-6 w-6 text-foreground/80" />
                      )}
                    </Button>
                  </EmojiPicker>
                  {answer.emoji && (
                    <div
                      className="flex items-center cursor-pointer text-muted-foreground hover:text-foreground"
                      onClick={() => updateAnswer(i, { emoji: undefined })}
                    >
                      <XIcon className="h-5 w-5" />
                    </div>
                  )}
                  <div className="flex-auto">
                    <PlaceholderInput
                      value={answer.text || ""}
                      onChange={(v) => updateAnswer(i, { text: v })}
                      placeholder={`Answer ${i + 1}`}
                    />
                  </div>
                  <Button
                    variant="outline"
                    size="icon"
                    className="flex-none"
                    onClick={() => removeAnswer(i)}
                  >
                    <MinusIcon className="h-5 w-5" />
                  </Button>
                </div>
                {error && (
                  <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
                    <CircleAlertIcon className="h-5 w-5 flex-none" />
                    <div>{error}</div>
                  </div>
                )}
              </div>
            );
          })}

          <div className="flex">
            <Button
              variant="outline"
              size="icon"
              onClick={addAnswer}
              disabled={answers.length >= pollMaxAnswers}
            >
              <PlusIcon className="h-5 w-5" />
            </Button>
          </div>
        </div>
        {answersError && (
          <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
            <CircleAlertIcon className="h-5 w-5 flex-none" />
            <div>{answersError}</div>
          </div>
        )}
      </div>
      <BaseInput
        type="text"
        field="poll_data.duration_hours"
        title="Duration"
        description="Number of hours the poll is open for, between 1 and 768 (32 days). Leave empty for 24 hours."
        value={data.poll_data?.duration_hours || ""}
        updateValue={(v) => updateField({ duration_hours: v || undefined })}
        errors={errors}
        placeholders
      />
      <BaseCheckbox
        field="poll_data.allow_multiselect"
        title="Allow Multiple Answers"
        description="If enabled, people can vote for more than one answer."
        value={!!data.poll_data?.allow_multiselect}
        updateValue={(v) => updateField({ allow_multiselect: v || undefined })}
        errors={errors}
      />
    </>
  );
}

function ModalDataInput({ data, updateData, errors }: InputProps) {
  return (
    <ModalEditorDialog
      modal={data.modal_data}
      onChange={(modal_data) => updateData({ modal_data })}
      errors={errors}
    />
  );
}

// Which of the optional channel fields Discord accepts for a given channel
// type. An unset type means "Text", matching the select's default.
//
// These must stay in sync with channelTypeSupportsTopic/Voice/RateLimitPerUser in
// kite-service/pkg/flow/data.go, which drops the same fields before sending
// them to Discord. Hidden here but sent there just means a confusing error.
function channelTypeSupportsNSFW(type = 0) {
  return type === 0;
}

function channelTypeSupportsTopic(type = 0) {
  // Forum (15) and media (16) use the topic as their post guidelines.
  return type === 0 || type === 5 || type === 15 || type === 16;
}

function channelTypeSupportsVoice(type = 0) {
  return type === 2 || type === 13;
}

function channelTypeSupportsRateLimitPerUser(type = 0) {
  // Everything but announcement (5) and category (4) channels.
  return type === 0 || type === 2 || type === 13 || type === 15 || type === 16;
}

function ChannelDataInput({ data, updateData, errors }: InputProps) {
  const addOverwrite = useCallback(() => {
    updateData({
      channel_data: {
        ...data.channel_data,
        permission_overwrites: [
          ...(data.channel_data?.permission_overwrites || []),
          { type: 0, allow: "0", deny: "0" },
        ],
      },
    });
  }, [updateData, data]);

  const clearOverwrites = useCallback(() => {
    updateData({
      channel_data: { ...data.channel_data, permission_overwrites: [] },
    });
  }, [updateData, data]);

  const removeOverwrite = useCallback(
    (i: number) => {
      updateData({
        channel_data: {
          ...data.channel_data,
          permission_overwrites:
            data.channel_data?.permission_overwrites?.filter((_, j) => j !== i),
        },
      });
    },
    [updateData, data]
  );

  const updateOverwrite = useCallback(
    (i: number, newData: Partial<PermissionOverwriteData>) => {
      updateData({
        channel_data: {
          ...data.channel_data,
          permission_overwrites: data.channel_data?.permission_overwrites?.map(
            (o, j) => (j === i ? { ...o, ...newData } : o)
          ),
        },
      });
    },
    [updateData, data]
  );

  // Changing the type only hides the inputs that don't apply to it. Without
  // also dropping their values, a topic entered while the type was "Text"
  // would still be sent for a voice channel, which Discord rejects.
  const updateType = useCallback(
    (type: number) => {
      const channelData = { ...data.channel_data, type };

      if (!channelTypeSupportsNSFW(type)) channelData.nsfw = undefined;
      if (!channelTypeSupportsTopic(type)) channelData.topic = undefined;
      if (!channelTypeSupportsRateLimitPerUser(type))
        channelData.rate_limit_per_user = undefined;
      if (!channelTypeSupportsVoice(type)) {
        channelData.bitrate = undefined;
        channelData.user_limit = undefined;
      }

      updateData({ channel_data: channelData });
    },
    [updateData, data]
  );

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button className="w-full">Configure Channel</Button>
      </DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle>Configure Channel</DialogTitle>
          <DialogDescription>
            Configure your channel here! A channel must have one of the
            available types and a name. All other fields are optional.
          </DialogDescription>
        </DialogHeader>
        <PreviewInert className="space-y-3">
          <div className="flex space-x-3">
            <BaseInput
              type="select"
              field={`channel_data.type`}
              title="Type"
              value={data.channel_data?.type?.toString() || "0"}
              options={[
                {
                  label: "Text",
                  value: "0",
                },
                {
                  label: "Voice",
                  value: "2",
                },
                {
                  label: "Category",
                  value: "4",
                },
                {
                  label: "Announcement",
                  value: "5",
                },
                {
                  label: "Stage",
                  value: "13",
                },
                {
                  label: "Forum",
                  value: "15",
                },
                {
                  label: "Media",
                  value: "16",
                },
              ]}
              updateValue={(v) => updateType(parseInt(v) || 0)}
              errors={errors}
            />

            {channelTypeSupportsNSFW(data.channel_data?.type) && (
              <BaseCheckbox
                field="channel_data.nsfw"
                title="NSFW"
                value={data.channel_data?.nsfw || false}
                updateValue={(v) =>
                  updateData({
                    channel_data: {
                      ...data.channel_data,
                      nsfw: v,
                    },
                  })
                }
                errors={errors}
              />
            )}
          </div>

          <BaseInput
            type="text"
            field="channel_data.name"
            title="Name"
            description="The name for the channel."
            value={data.channel_data?.name || ""}
            updateValue={(v) =>
              updateData({
                channel_data: { ...data.channel_data, name: v || undefined },
              })
            }
            errors={errors}
            placeholders
          />

          {channelTypeSupportsTopic(data.channel_data?.type) && (
            <BaseInput
              type="text"
              field="channel_data.topic"
              title="Topic"
              description="The topic for the channel."
              value={data.channel_data?.topic || ""}
              updateValue={(v) =>
                updateData({
                  channel_data: {
                    ...data.channel_data,
                    topic: v || undefined,
                  },
                })
              }
              errors={errors}
              placeholders
            />
          )}

          {channelTypeSupportsRateLimitPerUser(data.channel_data?.type) && (
            <BaseInput
              type="text"
              field="channel_data.rate_limit_per_user"
              title="Slowmode"
              description="How many seconds users must wait between messages (0-21600, max 6 hours). 0 turns slowmode off. Leave empty to skip it."
              value={data.channel_data?.rate_limit_per_user || ""}
              updateValue={(v) =>
                updateData({
                  channel_data: {
                    ...data.channel_data,
                    rate_limit_per_user: v || undefined,
                  },
                })
              }
              errors={errors}
              placeholders
            />
          )}

          {channelTypeSupportsVoice(data.channel_data?.type) && (
            <>
              <BaseInput
                type="text"
                field="channel_data.bitrate"
                title="Bitrate"
                description="The bitrate for the channel."
                value={data.channel_data?.bitrate || ""}
                updateValue={(v) =>
                  updateData({
                    channel_data: {
                      ...data.channel_data,
                      bitrate: v || undefined,
                    },
                  })
                }
                errors={errors}
                placeholders
              />
              <BaseInput
                type="text"
                field="channel_data.user_limit"
                title="User Limit"
                description="The user limit for the channel."
                value={data.channel_data?.user_limit?.toString() || ""}
                updateValue={(v) =>
                  updateData({
                    channel_data: {
                      ...data.channel_data,
                      user_limit: v || undefined,
                    },
                  })
                }
                errors={errors}
                placeholders
              />
            </>
          )}

          {data.channel_data?.type !== 4 && (
            <BaseInput
              type="text"
              field="channel_data.parent"
              title="Category"
              description="The category that the channel will be created in."
              value={data.channel_data?.parent || ""}
              updateValue={(v) =>
                updateData({
                  channel_data: {
                    ...data.channel_data,
                    parent: v || undefined,
                  },
                })
              }
              errors={errors}
              placeholders
            />
          )}

          <BaseInput
            type="text"
            field="channel_data.position"
            title="Position"
            description="The position for the channel."
            value={data.channel_data?.position?.toString() || ""}
            updateValue={(v) =>
              updateData({
                channel_data: {
                  ...data.channel_data,
                  position: v || undefined,
                },
              })
            }
            errors={errors}
            placeholders
          />

          <div className="space-y-3">
            <div className="font-medium text-foreground">
              Permission Overwrites
            </div>

            {data.channel_data?.permission_overwrites?.map((overwrite, i) => (
              <Card className="space-y-3 p-3 -mx-1" key={i}>
                <BaseInput
                  type="select"
                  field={`channel_data.permission_overwrites.${i}.type`}
                  title="Type"
                  value={overwrite.type?.toString() || "0"}
                  options={[
                    {
                      label: "Role",
                      value: "0",
                    },
                    {
                      label: "User",
                      value: "1",
                    },
                  ]}
                  updateValue={(v) =>
                    updateOverwrite(i, {
                      type: parseInt(v) || 0,
                    })
                  }
                  errors={errors}
                />
                <BaseInput
                  type="text"
                  field={`channel_data.permission_overwrites.${i}.id`}
                  title={overwrite.type === 0 ? "Role" : "User"}
                  description="Used to identify the input in your flow."
                  value={overwrite.id || ""}
                  updateValue={(v) =>
                    updateOverwrite(i, {
                      id: v || undefined,
                    })
                  }
                  errors={errors}
                  placeholders
                />
                <BasePermissionInput
                  field={`channel_data.permission_overwrites.${i}.allow`}
                  title="Allow"
                  value={overwrite.allow || ""}
                  updateValue={(v) =>
                    updateOverwrite(i, {
                      allow: v || undefined,
                    })
                  }
                  errors={errors}
                />
                <BasePermissionInput
                  field={`channel_data.permission_overwrites.${i}.deny`}
                  title="Deny"
                  value={overwrite.deny || ""}
                  updateValue={(v) =>
                    updateOverwrite(i, {
                      deny: v || undefined,
                    })
                  }
                  errors={errors}
                />
                <Button
                  variant="outline"
                  size="sm"
                  className="flex gap-2"
                  onClick={() => removeOverwrite(i)}
                >
                  <TrashIcon className="h-4 w-4" />
                  <div>Remove Overwrite</div>
                </Button>
              </Card>
            ))}
          </div>

          <div className="flex space-x-3">
            <Button onClick={addOverwrite}>Add Overwrite</Button>
            <Button variant="outline" onClick={clearOverwrites}>
              Clear Overwrites
            </Button>
          </div>
        </PreviewInert>
      </DialogContent>
    </Dialog>
  );
}

function ForumPostDataInput({ data, updateData, errors }: InputProps) {
  return (
    <>
      <BaseInput
        type="text"
        field="channel_data.name"
        title="Post Title"
        value={data.channel_data?.name || ""}
        updateValue={(v) =>
          updateData({
            channel_data: { ...data.channel_data, name: v || undefined },
          })
        }
        errors={errors}
        placeholders
      />
      <BaseInput
        type="text"
        field="channel_data.applied_tags"
        title="Tags"
        description="Names of up to 5 of the forum's tags, separated by commas."
        value={data.channel_data?.applied_tags || ""}
        updateValue={(v) =>
          updateData({
            channel_data: {
              ...data.channel_data,
              applied_tags: v || undefined,
            },
          })
        }
        errors={errors}
        placeholders
      />
    </>
  );
}

function ThreadDataInput({ data, updateData, errors }: InputProps) {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button className="w-full">Configure Thread</Button>
      </DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle>Configure Thread</DialogTitle>
          <DialogDescription>
            Configure your thread here! A thread must have a name and can be
            associated with a message or be independent.
          </DialogDescription>
        </DialogHeader>
        <PreviewInert className="space-y-3">
          <BaseInput
            type="text"
            field="channel_data.parent"
            title="Parent Channel"
            description="The parent channel that the thread will be created in."
            value={data.channel_data?.parent || ""}
            updateValue={(v) =>
              updateData({
                channel_data: {
                  ...data.channel_data,
                  parent: v || undefined,
                },
              })
            }
            errors={errors}
            placeholders
          />

          <BaseInput
            type="text"
            field="message_target"
            title="Message Target"
            description="The message to start the thread from. Leave blank to create an independent thread."
            value={data.message_target || ""}
            updateValue={(v) =>
              updateData({
                message_target: v || undefined,
              })
            }
            errors={errors}
            placeholders
          />

          {!data.message_target && (
            <BaseInput
              type="select"
              field={`channel_data.type`}
              title="Type"
              value={data.channel_data?.type?.toString() || "11"}
              options={[
                {
                  label: "Public Thread",
                  value: "11",
                },
                {
                  label: "Private Thread",
                  value: "12",
                },
                {
                  label: "Announcement Thread",
                  value: "10 ",
                },
              ]}
              updateValue={(v) =>
                updateData({
                  channel_data: {
                    ...data.channel_data,
                    type: parseInt(v) || undefined,
                  },
                })
              }
              errors={errors}
            />
          )}

          <BaseInput
            type="text"
            field="channel_data.name"
            title="Name"
            description="The name for the channel."
            value={data.channel_data?.name || ""}
            updateValue={(v) =>
              updateData({
                channel_data: { ...data.channel_data, name: v || undefined },
              })
            }
            errors={errors}
            placeholders
          />

          {(!data.channel_data?.type || data.channel_data.type === 0) && (
            <BaseCheckbox
              field="channel_data.invitable"
              title="Invitable"
              description="Whether non-moderators can add new members to the thread."
              value={data.channel_data?.invitable || false}
              updateValue={(v) =>
                updateData({
                  channel_data: {
                    ...data.channel_data,
                    invitable: v,
                  },
                })
              }
              errors={errors}
            />
          )}
        </PreviewInert>
      </DialogContent>
    </Dialog>
  );
}

function ChannelTargetInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="channel_target"
      title="Target Channel"
      value={data.channel_target || ""}
      updateValue={(v) => updateData({ channel_target: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function ForwardSourceChannelTargetInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      type="text"
      field="channel_target"
      title="Message Channel"
      description="The channel the message to forward is in. Leave blank to use the channel the flow runs in."
      value={data.channel_target || ""}
      updateValue={(v) => updateData({ channel_target: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function ForwardChannelTargetInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="forward_channel_target"
      title="Forward To"
      description="The channel to forward the message to."
      value={data.forward_channel_target || ""}
      updateValue={(v) =>
        updateData({ forward_channel_target: v || undefined })
      }
      errors={errors}
      placeholders
    />
  );
}

function VoiceSelfMuteInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseCheckbox
      field="voice_self_mute"
      title="Mute Self"
      description="If enabled, the bot joins the voice channel muted."
      value={!!data.voice_self_mute}
      updateValue={(v) => updateData({ voice_self_mute: v || undefined })}
      errors={errors}
    />
  );
}

function VoiceSelfDeafInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseCheckbox
      field="voice_self_deaf"
      title="Deafen Self"
      description="If enabled, the bot joins the voice channel deafened."
      value={!!data.voice_self_deaf}
      updateValue={(v) => updateData({ voice_self_deaf: v || undefined })}
      errors={errors}
    />
  );
}

function StatusDataInput({ data, updateData, errors }: InputProps) {
  const updateField = (newData: Partial<StatusData>) =>
    updateData({ status_data: { ...data.status_data, ...newData } });

  return (
    <>
      <BaseInput
        type="select"
        field="status_data.status"
        title="Status"
        options={statusOptions}
        value={data.status_data?.status || "online"}
        updateValue={(v) => updateField({ status: v || undefined })}
        errors={errors}
      />
      <BaseInput
        type="select"
        field="status_data.activity_type"
        title="Activity Type"
        options={activityTypeOptions}
        value={(data.status_data?.activity_type || 0).toString()}
        updateValue={(v) =>
          updateField({
            activity_type: parseInt(v) || undefined,
            // Custom statuses have no second line
            activity_state:
              v !== "4" ? data.status_data?.activity_state : undefined,
            // Only streaming activities have a URL
            activity_url:
              v === "1" ? data.status_data?.activity_url : undefined,
          })
        }
        errors={errors}
      />
      <BaseInput
        field="status_data.activity_name"
        title="Activity Name"
        value={data.status_data?.activity_name || ""}
        updateValue={(v) => updateField({ activity_name: v || undefined })}
        errors={errors}
        placeholders
      />
      {data.status_data?.activity_type !== 4 && (
        <BaseInput
          field="status_data.activity_state"
          title="Second Line"
          description="Optional text shown under the activity name."
          value={data.status_data?.activity_state || ""}
          updateValue={(v) => updateField({ activity_state: v || undefined })}
          errors={errors}
          placeholders
        />
      )}
      {data.status_data?.activity_type === 1 && (
        <BaseInput
          field="status_data.activity_url"
          title="Stream URL"
          description="Twitch or YouTube URL shown for the streaming activity."
          value={data.status_data?.activity_url || ""}
          updateValue={(v) => updateField({ activity_url: v || undefined })}
          errors={errors}
          placeholders
        />
      )}
    </>
  );
}

function RoleDataInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="role_data"
      title="Role Name"
      value={data.role_data?.name || ""}
      updateValue={(v) =>
        updateData({ role_data: v ? { name: v } : undefined })
      }
      errors={errors}
      placeholders
    />
  );
}

function RoleTargetInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="role_target"
      title="Target Role"
      value={data.role_target || ""}
      updateValue={(v) => updateData({ role_target: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function VariableIdInput({ data, updateData, errors }: InputProps) {
  const variables = useFlowVariables();
  const readOnly = useContext(FlowReadOnlyContext);

  const appId = useAppId();

  return (
    <div className="flex space-x-2 items-end">
      <BaseInput
        type="select"
        field="variable_id"
        title="Variable"
        options={variables?.map((v) => ({
          value: v!.id,
          label: v!.name,
        }))}
        value={data.variable_id || ""}
        updateValue={(v) => updateData({ variable_id: v || undefined })}
        errors={errors}
        clearable
      />
      {!readOnly &&
        (data.variable_id ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="outline" size="icon" asChild>
                <Link
                  href={{
                    pathname: "/apps/[appId]/variables/[variableId]",
                    query: { appId: appId, variableId: data.variable_id },
                  }}
                  target="_blank"
                >
                  <PencilIcon className="h-5 w-5" />
                </Link>
              </Button>
            </TooltipTrigger>
            <TooltipContent>Manage variable</TooltipContent>
          </Tooltip>
        ) : (
          <VariableCreateDialog
            onVariableCreated={(v) => updateData({ variable_id: v })}
          >
            <Button variant="outline" size="icon">
              <PlusIcon className="h-5 w-5" />
            </Button>
          </VariableCreateDialog>
        ))}
    </div>
  );
}

function VariableScopeInput({ data, updateData, errors }: InputProps) {
  const variables = useFlowVariables();
  const readOnly = useContext(FlowReadOnlyContext);

  const scoped = useMemo(() => {
    const variable = variables?.find((v) => v?.id === data.variable_id);
    return variable?.scoped;
  }, [variables, data]);

  const hasScope = !!data.variable_scope?.trim();
  // Shown for unscoped variables too while a scope is left, so it can be
  // cleared.
  if (!scoped && !hasScope) return null;

  // Previews don't show validation.
  let error: string | undefined;
  if (!readOnly && scoped === false) {
    error = "The variable isn't scoped, so the scope is ignored.";
  } else if (!readOnly && scoped && !hasScope) {
    error = "Scoped variables need a scope.";
  }

  return (
    <BaseInput
      type="text"
      field="variable_scope"
      title="Scope"
      value={data.variable_scope || ""}
      updateValue={(v) => updateData({ variable_scope: v || undefined })}
      errors={error ? { ...errors, variable_scope: error } : errors}
      placeholders
    />
  );
}

function VariableOperationInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="select"
      field="variable_operation"
      title="Operation"
      value={data.variable_operation || ""}
      updateValue={(v) => updateData({ variable_operation: v || undefined })}
      options={[
        { value: "overwrite", label: "Overwrite" },
        { value: "append", label: "Append" },
        { value: "prepend", label: "Prepend" },
        { value: "increment", label: "Increment" },
        { value: "decrement", label: "Decrement" },
      ]}
      errors={errors}
    />
  );
}

function VariableValueInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="variable_value"
      title="Value"
      value={data.variable_value || ""}
      updateValue={(v) => updateData({ variable_value: v || undefined })}
      errors={errors}
      placeholders
    />
  );
}

function ConditionCompareBaseValueInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      field="condition_base_value"
      title="Base Value"
      value={data.condition_base_value || ""}
      updateValue={(v) =>
        updateData({
          condition_base_value: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

const numericComparisonModes = [
  "greater_than",
  "less_than",
  "greater_than_or_equal",
  "less_than_or_equal",
];

function ConditionItemCompareModeInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      type="select"
      field="condition_item_mode"
      title="Comparison Mode"
      options={[
        { value: "equal", label: "Equal" },
        { value: "not_equal", label: "Not Equal" },
        { value: "greater_than", label: "Greater Than" },
        { value: "less_than", label: "Less Than" },
        { value: "greater_than_or_equal", label: "Greater Than or Equal" },
        { value: "less_than_or_equal", label: "Less Than or Equal" },
        { value: "contains", label: "Contains" },
        { value: "starts_with", label: "Starts With" },
        { value: "ends_with", label: "Ends With" },
        ...oneOfModes,
        { value: "contains_any", label: "Contains Any Of" },
      ]}
      value={data.condition_item_mode || ""}
      updateValue={(v) =>
        updateData({
          condition_item_mode: v || undefined,
          // Number modes hide Ignore Case, so it can't stay set unseen.
          condition_item_ignore_case: numericComparisonModes.includes(v)
            ? undefined
            : data.condition_item_ignore_case,
        })
      }
      errors={errors}
    />
  );
}

const conditionListModes = ["one_of", "not_one_of", "contains_any"];

function ConditionItemCompareValueInput({
  data,
  updateData,
  errors,
}: InputProps) {
  const value = data.condition_item_value || "";
  const updateValue = (v: string) =>
    updateData({ condition_item_value: v || undefined });

  if (conditionListModes.includes(data.condition_item_mode || "")) {
    return (
      <ListInput
        field="condition_item_value"
        title="Comparison Values"
        noun="value"
        value={value}
        updateValue={updateValue}
        errors={errors}
      />
    );
  }

  return (
    <BaseInput
      // A list stays multi-line after switching to another mode, so its lines
      // aren't hidden in a single-line input.
      type={value.includes("\n") ? "textarea" : "text"}
      field="condition_item_value"
      title="Comparison Value"
      value={value}
      updateValue={updateValue}
      errors={errors}
      placeholders
    />
  );
}

// ListInput edits a setting that takes a list: one item per line, or a
// placeholder that holds a list.
function ListInput({
  field,
  title,
  noun,
  value,
  updateValue,
  errors,
}: {
  field: string;
  title: string;
  noun: string;
  value: string;
  updateValue: (value: string) => void;
  errors: Record<string, string>;
}) {
  return (
    <BaseInput
      type="textarea"
      field={field}
      title={title}
      description={`One ${noun} per line, or a placeholder that holds a list. ${describeList(
        value,
        noun
      )}.`}
      value={value}
      updateValue={updateValue}
      errors={errors}
      placeholders
    />
  );
}

function ConditionItemIgnoreCaseInput({
  data,
  updateData,
  errors,
}: InputProps) {
  if (numericComparisonModes.includes(data.condition_item_mode || "")) {
    return null;
  }

  return (
    <BaseCheckbox
      field="condition_item_ignore_case"
      title="Ignore Case"
      description="If enabled, upper and lower case don't matter, so gtg also matches GTG."
      value={!!data.condition_item_ignore_case}
      updateValue={(v) =>
        updateData({ condition_item_ignore_case: v || undefined })
      }
      errors={errors}
    />
  );
}

function ConditionUserBaseValueInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="condition_base_value"
      title="Base User"
      value={data.condition_base_value || ""}
      updateValue={(v) =>
        updateData({
          condition_base_value: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function ConditionItemUserModeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="select"
      field="condition_item_mode"
      title="Comparison Mode"
      options={[
        { value: "equal", label: "Equal" },
        { value: "not_equal", label: "Not Equal" },
        { value: "has_role", label: "Has Role" },
        { value: "not_has_role", label: "Does Not Have Role" },
        { value: "has_permission", label: "Has Permission" },
        { value: "not_has_permission", label: "Does Not Have Permission" },
      ]}
      value={data.condition_item_mode || ""}
      updateValue={(v) => updateData({ condition_item_mode: v || undefined })}
      errors={errors}
    />
  );
}

function ConditionItemUserValueInput({ data, updateData, errors }: InputProps) {
  if (
    data.condition_item_mode === "has_role" ||
    data.condition_item_mode === "not_has_role"
  ) {
    return (
      <BaseInput
        field="condition_item_value"
        title="Comparison Role"
        value={data.condition_item_value || ""}
        updateValue={(v) =>
          updateData({
            condition_item_value: v || undefined,
          })
        }
        errors={errors}
        placeholders
      />
    );
  } else if (
    data.condition_item_mode === "has_permission" ||
    data.condition_item_mode === "not_has_permission"
  ) {
    return (
      <BasePermissionInput
        field="condition_item_value"
        title="Comparison Permission"
        value={data.condition_item_value || "0"}
        updateValue={(v) =>
          updateData({
            condition_item_value: v === "0" ? undefined : v,
          })
        }
        errors={errors}
      />
    );
  }

  return (
    <BaseInput
      field="condition_item_value"
      title="Comparison User"
      value={data.condition_item_value || ""}
      updateValue={(v) =>
        updateData({
          condition_item_value: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function ConditionChannelBaseValueInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      field="condition_base_value"
      title="Base Channel"
      value={data.condition_base_value || ""}
      updateValue={(v) =>
        updateData({
          condition_base_value: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function ConditionItemChannelModeInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      type="select"
      field="condition_item_mode"
      title="Comparison Mode"
      options={[
        { value: "equal", label: "Equal" },
        { value: "not_equal", label: "Not Equal" },
      ]}
      value={data.condition_item_mode || ""}
      updateValue={(v) => updateData({ condition_item_mode: v || undefined })}
      errors={errors}
    />
  );
}

function ConditionItemChannelValueInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      field="condition_item_value"
      title="Comparison Channel"
      value={data.condition_item_value || ""}
      updateValue={(v) =>
        updateData({
          condition_item_value: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function ConditionRoleBaseValueInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="condition_base_value"
      title="Base Role"
      value={data.condition_base_value || ""}
      updateValue={(v) =>
        updateData({
          condition_base_value: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function ConditionItemRoleModeInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="select"
      field="condition_item_mode"
      title="Comparison Mode"
      options={[
        { value: "equal", label: "Equal" },
        { value: "not_equal", label: "Not Equal" },
      ]}
      value={data.condition_item_mode || ""}
      updateValue={(v) => updateData({ condition_item_mode: v || undefined })}
      errors={errors}
    />
  );
}

function ConditionItemRoleValueInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="condition_item_value"
      title="Comparison Role"
      value={data.condition_item_value || ""}
      updateValue={(v) =>
        updateData({
          condition_item_value: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function ConditionAllowMultipleInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseCheckbox
      field="condition_allow_multiple"
      title="Allow Multiple"
      description="Allow multiple conditions to be met. If disabled, only the first condition that is met will be executed."
      value={data.condition_allow_multiple || false}
      updateValue={(v) =>
        updateData({ condition_allow_multiple: v || undefined })
      }
      errors={errors}
    />
  );
}

function ControlLoopCountInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="loop_count"
      title="Loop Count"
      description="The number of times to run the loop."
      value={data.loop_count || ""}
      updateValue={(v) =>
        updateData({
          loop_count: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function ControlSleepDurationInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="sleep_duration_seconds"
      title="Wait Duration"
      description="The number of seconds to wait before continuing, up to 30 days. Waits over 5 seconds end the current run and continue later, so responses only work if the flow continues within 15 minutes."
      value={data.sleep_duration_seconds || ""}
      updateValue={(v) =>
        updateData({
          sleep_duration_seconds: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function BaseInput({
  type,
  field,
  options,
  title,
  description,
  errors,
  value,
  placeholder,
  updateValue,
  placeholders,
  disablePlaceholderBrackets,
  clearable,
}: {
  type?: "text" | "textarea" | "select";
  field: string;
  options?: { value: string; label: string }[];
  // Left out for an input beside others, like the value of a header.
  title?: string;
  description?: ReactNode;
  errors: Record<string, string>;
  value: string;
  placeholder?: string;
  updateValue: (value: string) => void;
  // "parameters" offers only the flow's parameters.
  placeholders?: boolean | "parameters";
  disablePlaceholderBrackets?: boolean;
  clearable?: boolean;
}) {
  const error = errors[field];

  const inputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const onPlaceholderSelect = useCallback(
    (placeholder: string) => {
      const value = disablePlaceholderBrackets
        ? placeholder
        : `{{${placeholder}}}`;

      const element =
        type === "textarea" ? textareaRef.current : inputRef.current;

      if (!element) return;

      const start = element.selectionStart ?? 0;
      const end = element.selectionEnd ?? 0;

      const newValue =
        element.value.substring(0, start) +
        value +
        element.value.substring(end);

      updateValue(newValue);
    },
    [inputRef, textareaRef, type, updateValue, disablePlaceholderBrackets]
  );

  return (
    <div className="flex-auto">
      {title && <div className="font-medium text-foreground">{title}</div>}
      {description ? (
        <div className="text-muted-foreground text-sm mt-1">{description}</div>
      ) : null}
      <div className={cn("relative", title && "mt-2")}>
        {type === "textarea" ? (
          <Textarea
            value={value}
            onChange={(e) => updateValue(e.target.value)}
            ref={textareaRef}
            placeholder={placeholder}
            // Room for the placeholder explorer in the corner.
            className={cn(placeholders && "pr-10")}
          />
        ) : type === "select" ? (
          <Select value={value} onValueChange={(v) => updateValue(v)}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {options?.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
              {clearable && (
                <>
                  <SelectSeparator />
                  <Button
                    className="w-full px-2"
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      updateValue("");
                    }}
                  >
                    Clear Selection
                  </Button>
                </>
              )}
            </SelectContent>
          </Select>
        ) : placeholders ? (
          <PlaceholderInput
            value={value}
            onChange={(v) => updateValue(v)}
            ref={inputRef}
            placeholder={placeholder}
          />
        ) : (
          <Input
            type="text"
            value={value}
            onChange={(e) => updateValue(e.target.value)}
            ref={inputRef}
            placeholder={placeholder}
          />
        )}
        {placeholders && (
          <FlowPlaceholderExplorer
            onSelect={onPlaceholderSelect}
            hideBrackets={disablePlaceholderBrackets}
            field={field}
            onlyParameters={placeholders === "parameters"}
          />
        )}
      </div>
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}

function BaseCheckbox({
  field,
  title,
  description,
  errors,
  value,
  updateValue,
}: {
  field: string;
  title: string;
  description?: string;
  errors: Record<string, string>;
  value: boolean;
  updateValue: (value: boolean) => void;
}) {
  const error = errors[field];

  return (
    <div>
      <div className="font-medium text-foreground mb-2">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mb-2">{description}</div>
      ) : null}
      <Switch checked={value} onCheckedChange={updateValue} />
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}

function BaseMultiSelect({
  field,
  title,
  description,
  errors,
  options,
  values,
  updateValues,
  searchable,
}: {
  field: string;
  title: string;
  description?: string;
  errors: Record<string, string>;
  options: { value: string; label: string }[];
  values: string[];
  updateValues: (value: string[]) => void;
  // Shows a search bar at the top of the list, for long lists of options.
  searchable?: boolean;
}) {
  const error = errors[field];

  function toggleValue(value: string) {
    if (values.includes(value)) {
      updateValues(values.filter((val) => val !== value));
    } else {
      updateValues([...values, value]);
    }
  }

  const trigger = (
    <Button variant="outline" className="w-full flex items-center">
      <div>{values.length} selected</div>
      <ChevronDownIcon className="h-4 w-4 ml-auto" />
    </Button>
  );

  return (
    <div>
      <div className="font-medium text-foreground mb-2">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mb-2">{description}</div>
      ) : null}
      {searchable ? (
        <Popover modal>
          <PopoverTrigger asChild>{trigger}</PopoverTrigger>
          <PopoverContent className="w-64 p-0">
            <Command>
              <CommandInput placeholder="Search..." />
              <CommandList>
                <CommandEmpty>No results</CommandEmpty>
                <CommandGroup>
                  {options.map((o) => (
                    <CommandItem
                      key={o.value}
                      value={o.value}
                      keywords={[o.label]}
                      onSelect={() => toggleValue(o.value)}
                      // Keeps focus in the search bar to pick several options.
                      onMouseDown={(e) => e.preventDefault()}
                    >
                      <CheckIcon
                        className={cn(
                          "mr-2 h-4 w-4 flex-none",
                          values.includes(o.value) ? "opacity-100" : "opacity-0"
                        )}
                      />
                      {o.label}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
      ) : (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>{trigger}</DropdownMenuTrigger>
          <DropdownMenuContent className="w-56 max-h-[320px] overflow-y-auto">
            {options.map((o) => (
              <DropdownMenuCheckboxItem
                key={o.value}
                checked={values.includes(o.value)}
                onCheckedChange={() => toggleValue(o.value)}
              >
                {o.label}
              </DropdownMenuCheckboxItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      )}
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}

// A color setting, picked like the color of an embed and stored as a number,
// or a placeholder.
function BaseColorInput({
  field,
  title,
  description,
  errors,
  value,
  updateValue,
}: {
  field: string;
  title: string;
  description?: string;
  errors: Record<string, string>;
  value: string;
  updateValue: (value: string) => void;
}) {
  const parsed = parseColorValue(value);
  const color = parsed !== undefined && parsed <= 0xffffff ? parsed : undefined;
  // Values that aren't colors, like placeholders, are shown as text, so they
  // can be fixed instead of being lost. Once typed in, text stays text.
  const [textMode, setTextMode] = useState(false);

  if (textMode || (!!value && color === undefined)) {
    return (
      <div>
        <BaseInput
          type="text"
          field={field}
          title={title}
          description={description}
          value={value}
          updateValue={(v) => {
            setTextMode(true);
            updateValue(v);
          }}
          errors={errors}
          placeholder="#ff0000 or {{ placeholder }}"
          placeholders
        />
        <Button
          variant="link"
          size="sm"
          className="px-0 h-auto mt-1"
          onClick={() => {
            setTextMode(false);
            if (color === undefined) updateValue("");
          }}
        >
          Use the color picker
        </Button>
      </div>
    );
  }

  const error = errors[field];

  return (
    <div className="flex-auto">
      <div className="font-medium text-foreground">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mt-1">{description}</div>
      ) : null}
      <div className="mt-2">
        <ColorPicker
          value={color}
          onChange={(v) => updateValue(v === undefined ? "" : v.toString())}
        />
      </div>
      <Button
        variant="link"
        size="sm"
        className="px-0 h-auto mt-1"
        onClick={() => setTextMode(true)}
      >
        Use a placeholder instead
      </Button>
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}

function BasePermissionInput({
  field,
  title,
  description,
  errors,
  value,
  updateValue,
  searchable,
}: {
  field: string;
  title: string;
  description?: string;
  errors: Record<string, string>;
  value: string;
  updateValue: (value: string) => void;
  searchable?: boolean;
}) {
  const availablePermissions = useMemo(
    () =>
      permissionBits.map((p) => ({
        value: p.bit.toString(),
        label: p.label,
      })),
    []
  );

  const enabledPermissions = useMemo(
    () => decodePermissionsBitset(value || "0").map((p) => p.bit.toString()),
    [value]
  );

  const setPermissions = useCallback(
    (v: string[]) => {
      const newPerms = encodePermissionsBitset(v.map((p) => parseInt(p)));

      updateValue(newPerms);
    },
    [updateValue]
  );

  return (
    <BaseMultiSelect
      field={field}
      title={title}
      description={description}
      values={enabledPermissions}
      options={availablePermissions}
      updateValues={setPermissions}
      errors={errors}
      searchable={searchable}
    />
  );
}

function BaseEmojiPicker({
  title,
  description,
  field,
  errors,
  emoji,
  onChange,
}: {
  title: string;
  description?: string;
  field: string;
  errors: Record<string, string>;
  emoji: EmojiData | undefined;
  onChange: (emoji: EmojiData | undefined) => void;
}) {
  const error = errors[field];

  return (
    <div className="flex-none">
      <div className="font-medium text-foreground mb-2">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mb-2">{description}</div>
      ) : null}
      <div className="flex">
        <EmojiPicker onEmojiSelect={onChange}>
          <Button size="icon" variant="outline">
            {emoji?.id ? (
              <img src={discordEmojiUrl(emoji.id)} alt="" className="h-6 w-6" />
            ) : emoji ? (
              <Twemoji
                options={{
                  className: "h-6 w-6",
                }}
              >
                {emoji.name}
              </Twemoji>
            ) : (
              <SmileIcon className="h-6 w-6 text-foreground/80" />
            )}
          </Button>
        </EmojiPicker>
        {emoji && (
          <div
            className="flex items-center cursor-pointer pr-1 text-muted-foreground hover:text-foreground"
            onClick={() => onChange(undefined)}
          >
            <XIcon className="h-5 w-5" />
          </div>
        )}
      </div>
      {error && (
        <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-2">
          <CircleAlertIcon className="h-5 w-5 flex-none" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
}
