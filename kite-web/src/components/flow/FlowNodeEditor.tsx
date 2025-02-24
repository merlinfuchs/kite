import {
  decodePermissionsBitset,
  encodePermissionsBitset,
  permissionBits,
} from "@/lib/discord/permissions";
import { getNodeId, useNodeValues } from "@/lib/flow/nodes";
import { useMessages, useVariables } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { Node, useNodes, useReactFlow, useStoreApi } from "@xyflow/react";
import {
  ChevronDownIcon,
  CircleAlertIcon,
  CopyIcon,
  HelpCircleIcon,
  MinusIcon,
  PencilIcon,
  PlusIcon,
  TrashIcon,
  XIcon,
} from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useRef } from "react";
import { NodeData, NodeProps } from "../../lib/flow/data";
import MessageCreateDialog from "../app/MessageCreateDialog";
import VariableCreateDialog from "../app/VariableCreateDialog";
import PlaceholderInput from "../common/PlaceholderInput";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { Input } from "../ui/input";
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
import FlowPlaceholderExplorer from "./FlowPlaceholderExplorer";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Separator } from "../ui/separator";
import { Card } from "../ui/card";
import { HTTPRequestData, ModalComponentData } from "@/lib/types/flow.gen";
import JsonEditor from "../common/JsonEditor";

interface Props {
  nodeId: string;
}

interface InputProps {
  type: string;
  data: NodeData;
  updateData: (newData: Partial<NodeData>) => void;
  errors: Record<string, string>;
}

const intputs: Record<string, any> = {
  custom_label: CustomLabelInput,
  name: NameInput,
  description: DescriptionInput,
  command_argument_type: CommandArgumentTypeInput,
  command_argument_required: CommandArgumentRequiredInput,
  command_contexts: CommandContextsInput,
  command_integrations: CommandIntegrationsInput,
  command_permissions: CommandPermissionsInput,
  event_type: EventTypeInput,
  message_data: MessageDataInput,
  message_template_id: MessageTemplateInput,
  message_target: MessageTargetInput,
  response_target: ResponseTargetInput,
  message_ephemeral: MessageEphemeralInput,
  modal_data: ModalDataInput,
  channel_data: ChannelDataInput,
  channel_target: ChannelTargetInput,
  role_data: RoleDataInput,
  role_target: RoleTargetInput,
  variable_id: VariableIdInput,
  variable_scope: VariableScopeInput,
  variable_operation: VariableOperationInput,
  variable_value: VariableValueInput,
  http_request_data: HttpRequestDataInput,
  ai_chat_completion_data: AiChatCompletionDataInput,
  expression: ExpressionInput,
  random_min: RandomMinInput,
  random_max: RandomMaxInput,
  audit_log_reason: AuditLogReasonInput,
  user_target: UserTargetInput,
  member_ban_delete_message_duration_seconds:
    MemberBanDeleteMessageDurationInput,
  member_timeout_duration_seconds: MemberTimeoutDurationInput,
  member_nick: MemberNickInput,
  log_level: LogLevelInput,
  log_message: LogMessageInput,
  condition_compare_base_value: ConditionCompareBaseValueInput,
  condition_item_compare_mode: ConditionItemCompareModeInput,
  condition_item_compare_value: ConditionItemCompareValueInput,
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

export default function FlowNodeEditor({ nodeId }: Props) {
  const { setNodes } = useReactFlow<Node<NodeProps>>();
  const store = useStoreApi();

  function close() {
    store.getState().addSelectedNodes([]);
  }

  const nodes = useNodes<Node<NodeProps>>();

  const node = nodes.find((n) => n.id === nodeId);

  const data = node?.data;

  function updateData(newData: Partial<NodeData>) {
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
  }

  function deleteNode() {
    setNodes((nodes) => nodes.filter((n) => n.id !== nodeId));
  }

  function duplicateNode() {
    if (!node) return;

    const newNode = {
      ...node,
      id: getNodeId(),
      selected: false,
      position: {
        x: node?.position.x! + 100,
        y: node?.position.y! + 100,
      },
    };
    setNodes((nodes) => nodes.concat(newNode));
  }

  const values = useNodeValues(node?.type!);

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

  return (
    <div className="absolute top-0 left-0 bg-background w-96 h-full p-5 flex flex-col overflow-y-auto">
      <div className="flex-none">
        <div className="flex items-start justify-between mb-5">
          <div className="text-xl font-bold text-foreground">
            Block Settings
          </div>
          <XIcon
            className="h-6 w-6 text-muted-foreground hover:text-foreground cursor-pointer"
            onClick={close}
          />
        </div>
        <div className="mb-5">
          <div className="flex items-center gap-1.5">
            <div className="text-lg font-bold text-foreground mb-1">
              {values.defaultTitle}
            </div>
            {values.helpUrl && (
              <Link href={values.helpUrl} target="_blank">
                <HelpCircleIcon className="h-5 w-5 text-muted-foreground hover:text-foreground" />
              </Link>
            )}
          </div>
          <div className="text-muted-foreground mb-3">
            {values.defaultDescription}
          </div>
          <div className="flex">
            <div className="bg-muted rounded px-2 py-1 text-xs">{nodeId}</div>
          </div>
        </div>
      </div>
      <div className="space-y-3 flex-auto">
        {values.dataFields.map((field) => {
          const Input = intputs[field];
          if (!Input) return null;

          return (
            <Input
              key={field}
              type={node.type}
              data={data}
              updateData={updateData}
              errors={errors}
            />
          );
        })}
      </div>
      <div className="flex-none space-y-3 mt-5">
        {!values.fixed && (
          <>
            <button
              className="bg-red-500 hover:bg-red-600 px-3 py-2 w-full rounded text-white font-medium flex space-x-2 justify-center items-center"
              onClick={deleteNode}
            >
              <TrashIcon className="h-5 w-5" />
              <div>Delete Block</div>
            </button>
            <button
              className="bg-dark-5 hover:bg-dark-4 px-3 py-2 w-full rounded text-white font-medium flex space-x-2 justify-center items-center"
              onClick={duplicateNode}
            >
              <CopyIcon className="h-5 w-5" />
              <div>Duplicate Block</div>
            </button>
          </>
        )}
      </div>
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
  return (
    <BaseInput
      type="select"
      field="event_type"
      title="Event"
      options={[
        { value: "message_create", label: "Message Create" },
        { value: "message_update", label: "Message Update" },
        { value: "message_delete", label: "Message Delete" },
        { value: "guild_member_add", label: "Server Member Add" },
        { value: "guild_member_remove", label: "Server Member Remove" },
      ]}
      value={data.event_type || ""}
      updateValue={(v) => updateData({ event_type: v || undefined })}
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
        <div className="space-y-3">
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
          <div>
            <div className="font-medium text-foreground mb-1">Headers</div>
            <div className="text-muted-foreground text-sm mb-2">
              The headers to send with the request.
            </div>
            <div className="flex flex-col gap-3">
              {data.http_request_data?.headers?.map((h, i) => (
                <div className="flex gap-2" key={i}>
                  <Input
                    type="text"
                    placeholder="Key"
                    value={h.key || ""}
                    onChange={(e) => updateHeader(i, e.target.value, h.value)}
                  />
                  <Input
                    type="text"
                    placeholder="Value"
                    value={h.value || ""}
                    onChange={(e) => updateHeader(i, h.key, e.target.value)}
                  />
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
        </div>
      </DialogContent>
    </Dialog>
  );
}

function AiChatCompletionDataInput({ data, updateData, errors }: InputProps) {
  // TODO: top level errors aren't displayed ...

  return (
    <>
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

function ExpressionInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="textarea"
      field="expression"
      title="Expression"
      description="The expression to evaluate"
      value={data.expression || ""}
      updateValue={(v) =>
        updateData({
          expression: v || undefined,
        })
      }
      errors={errors}
      placeholders
    />
  );
}

function RandomMinInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="random_min"
      title="Min"
      value={data.random_min || ""}
      updateValue={(v) => updateData({ random_min: v || undefined })}
      errors={errors}
    />
  );
}

function RandomMaxInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      field="random_max"
      title="Max"
      value={data.random_max || ""}
      updateValue={(v) => updateData({ random_max: v || undefined })}
      errors={errors}
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

function MessageTemplateInput({ data, updateData, errors }: InputProps) {
  const messages = useMessages();

  const appId = useAppId();

  return (
    <div className="flex space-x-2 items-end">
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
        updateValue={(v) => updateData({ message_template_id: v || undefined })}
        errors={errors}
        clearable
      />
      {data.message_template_id ? (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="outline" size="icon" asChild>
              <Link
                href={{
                  pathname: "/apps/[appId]/messages/[messageId]",
                  query: { appId: appId, messageId: data.message_template_id },
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
      )}
    </div>
  );
}

function MessageDataInput({ data, updateData, errors }: InputProps) {
  if (data.message_template_id) {
    return null;
  }

  return (
    <BaseInput
      type="textarea"
      field="message_data"
      title="Text Response"
      description="Right now only text responses are supported, but more options will be added in the future."
      value={data.message_data?.content || ""}
      updateValue={(v) =>
        updateData({ message_data: v ? { content: v } : undefined })
      }
      errors={errors}
      placeholders
    />
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

function ResponseTargetInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="select"
      field="message_target"
      title="Target Response"
      value={data.message_target || ""}
      options={[
        {
          label: "Original Response",
          value: "@original",
        },
      ]}
      updateValue={(v) => updateData({ message_target: v || undefined })}
      errors={errors}
    />
  );
}

function MessageEphemeralInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseCheckbox
      field="message_ephemeral"
      title="Public Response"
      value={!data.message_ephemeral}
      updateValue={(v) => updateData({ message_ephemeral: !v || undefined })}
      errors={errors}
    />
  );
}

function ModalDataInput({ data, updateData, errors }: InputProps) {
  const addInput = useCallback(() => {
    updateData({
      modal_data: {
        title: data.modal_data?.title,
        components: [
          ...(data.modal_data?.components || []),
          { components: [{ style: 1 }] },
        ],
      },
    });
  }, [updateData, data]);

  const clearInputs = useCallback(() => {
    updateData({
      modal_data: {
        title: data.modal_data?.title,
        components: [],
      },
    });
  }, [updateData, data]);

  const updateComponentField = useCallback(
    (r: number, c: number, newData: Partial<ModalComponentData>) => {
      const current = data.modal_data || {};
      if (!current.components) return;

      const row = current.components[r];
      if (!row || !row.components) return;

      const component = row.components[c];
      if (!component) return;

      Object.assign(component, newData);

      updateData({
        modal_data: current,
      });
    },
    [updateData, data]
  );

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button className="w-full" variant="secondary">
          Configure Modal
        </Button>
      </DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle>Configure Modal</DialogTitle>
          <DialogDescription>
            Configure your modal here! A modal must have a title and at least
            one input component.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <BaseInput
            type="text"
            field="modal_data"
            title="Title"
            value={data.modal_data?.title || ""}
            updateValue={(v) =>
              updateData({
                modal_data: {
                  title: v || undefined,
                  components: data.modal_data?.components,
                },
              })
            }
            errors={errors}
            placeholders
          />

          <div className="space-y-3">
            <div className="font-medium text-foreground">Inputs</div>
            {data.modal_data?.components?.map((row, r) =>
              row?.components?.map((component, c) => (
                <Card className="space-y-3 p-3 -mx-1" key={`${r}-${c}`}>
                  <div className="flex space-x-3">
                    <BaseInput
                      type="select"
                      field={`modal_data.components.${r}.components.${c}.type`}
                      title="Type"
                      value={component.style?.toString() || "1"}
                      options={[
                        {
                          label: "Short",
                          value: "1",
                        },
                        {
                          label: "Paragraph",
                          value: "2",
                        },
                      ]}
                      updateValue={(v) =>
                        updateComponentField(r, c, {
                          style: parseInt(v) || 1,
                        })
                      }
                      errors={errors}
                    />
                    <BaseCheckbox
                      field={`modal_data.components.${r}.components.${c}.required`}
                      title="Required"
                      value={component?.required || false}
                      updateValue={(v) =>
                        updateComponentField(r, c, {
                          required: v,
                        })
                      }
                      errors={errors}
                    />
                  </div>
                  <BaseInput
                    type="text"
                    field={`modal_data.components.${r}.components.${c}.custom_id`}
                    title="Identifier"
                    description="Used to identify the input in your flow."
                    value={component?.custom_id || ""}
                    updateValue={(v) =>
                      updateComponentField(r, c, {
                        custom_id: v || undefined,
                      })
                    }
                    errors={errors}
                    placeholders
                  />
                  <BaseInput
                    type="text"
                    field={`modal_data.components.${r}.components.${c}.label`}
                    title="Label"
                    value={component?.label || ""}
                    updateValue={(v) =>
                      updateComponentField(r, c, {
                        label: v || undefined,
                      })
                    }
                    errors={errors}
                    placeholders
                  />
                  <BaseInput
                    type="text"
                    field={`modal_data.components.${r}.components.${c}.placeholder`}
                    title="Placeholder"
                    value={component?.placeholder || ""}
                    updateValue={(v) =>
                      updateComponentField(r, c, {
                        placeholder: v || undefined,
                      })
                    }
                    errors={errors}
                    placeholders
                  />
                </Card>
              ))
            )}
          </div>

          <div className="flex space-x-3">
            <Button
              onClick={addInput}
              disabled={(data.modal_data?.components?.length || 0) >= 5}
            >
              Add Input
            </Button>
            <Button variant="outline" onClick={clearInputs}>
              Clear Inputs
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function ChannelDataInput({ data, updateData, errors }: InputProps) {
  return (
    <BaseInput
      type="text"
      field="channel_data"
      title="Channel Name"
      value={data.channel_data?.name || ""}
      updateValue={(v) =>
        updateData({ channel_data: v ? { name: v } : undefined })
      }
      errors={errors}
    />
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
  const variables = useVariables();

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
      {data.variable_id ? (
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
      )}
    </div>
  );
}

function VariableScopeInput({ data, updateData, errors }: InputProps) {
  const variables = useVariables();

  const scoped = useMemo(() => {
    const variable = variables?.find((v) => v?.id === data.variable_id);
    return variable?.scoped;
  }, [variables, data]);

  useEffect(() => {
    if (scoped === false) {
      updateData({ variable_scope: undefined });
    }
  }, [scoped, updateData]);

  if (!scoped) return null;

  return (
    <BaseInput
      type="text"
      field="variable_scope"
      title="Scope"
      value={data.variable_scope || ""}
      updateValue={(v) => updateData({ variable_scope: v || undefined })}
      errors={errors}
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
      ]}
      value={data.condition_item_mode || ""}
      updateValue={(v) => updateData({ condition_item_mode: v || undefined })}
      errors={errors}
    />
  );
}

function ConditionItemCompareValueInput({
  data,
  updateData,
  errors,
}: InputProps) {
  return (
    <BaseInput
      field="condition_item_value"
      title="Comparison Value"
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
      description="The number of seconds to wait before continuing."
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
  updateValue,
  placeholders,
  clearable,
}: {
  type?: "text" | "textarea" | "select";
  field: string;
  options?: { value: string; label: string }[];
  title: string;
  description?: string;
  errors: Record<string, string>;
  value: string;
  updateValue: (value: string) => void;
  placeholders?: boolean;
  clearable?: boolean;
}) {
  const error = errors[field];

  const inputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const onPlaceholderSelect = useCallback(
    (placeholder: string) => {
      const value = `{{${placeholder}}}`;

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
    [inputRef, textareaRef, type, updateValue]
  );

  return (
    <div className="flex-auto">
      <div className="font-medium text-foreground">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mt-1">{description}</div>
      ) : null}
      <div className="relative mt-2">
        {type === "textarea" ? (
          <Textarea
            value={value}
            onChange={(e) => updateValue(e.target.value)}
            ref={textareaRef}
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
          />
        ) : (
          <Input
            type="text"
            value={value}
            onChange={(e) => updateValue(e.target.value)}
            ref={inputRef}
          />
        )}
        {placeholders && (
          <FlowPlaceholderExplorer onSelect={onPlaceholderSelect} />
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
}: {
  field: string;
  title: string;
  description?: string;
  errors: Record<string, string>;
  options: { value: string; label: string }[];
  values: string[];
  updateValues: (value: string[]) => void;
}) {
  const error = errors[field];

  return (
    <div>
      <div className="font-medium text-foreground mb-2">{title}</div>
      {description ? (
        <div className="text-muted-foreground text-sm mb-2">{description}</div>
      ) : null}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="w-full flex items-center">
            <div>{values.length} selected</div>
            <ChevronDownIcon className="h-4 w-4 ml-auto" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-56 max-h-[320px] overflow-y-auto">
          {options.map((o) => (
            <DropdownMenuCheckboxItem
              key={o.value}
              checked={values.includes(o.value)}
              onCheckedChange={(v) => {
                if (v) {
                  updateValues([...values, o.value]);
                } else {
                  updateValues(values.filter((val) => val !== o.value));
                }
              }}
            >
              {o.label}
            </DropdownMenuCheckboxItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
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
}: {
  field: string;
  title: string;
  description?: string;
  errors: Record<string, string>;
  value: string;
  updateValue: (value: string) => void;
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
    />
  );
}
