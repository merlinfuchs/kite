import {
  ArrowUturnLeftIcon,
  ArrowsRightLeftIcon,
  ChatBubbleBottomCenterIcon,
  CommandLineIcon,
  DocumentIcon,
  DocumentTextIcon,
  EnvelopeIcon,
  ExclamationCircleIcon,
  HashtagIcon,
  LanguageIcon,
  PlusIcon,
  QuestionMarkCircleIcon,
  TagIcon,
  UserIcon,
} from "@heroicons/react/24/solid";
import { ExoticComponent, useMemo } from "react";
import {
  nodeActionDataSchema,
  nodeActionLogDataSchema,
  nodeCommandDataSchema,
  nodeOptionDataSchema,
} from "./data";
import { ZodSchema } from "zod";

export const primaryColor = "#F97316";

export const actionColor = "#3b82f6";
export const entryColor = "#eab308";
export const errorColor = "#ef4444";
export const conditionColor = "#22c55e";
export const optionColor = "#a855f7";

export interface NodeValues {
  color: string;
  icon: ExoticComponent<{ className: string }>;
  defaultTitle: string;
  defaultDescription: string;
  dataSchema?: ZodSchema;
  dataFields: string[];
}

export const nodeTypes: Record<string, NodeValues> = {
  entry_command: {
    color: entryColor,
    icon: CommandLineIcon,
    defaultTitle: "Command",
    defaultDescription:
      "Command entry. Drop different actions and options here!",
    dataSchema: nodeCommandDataSchema,
    dataFields: ["name", "description"],
  },
  entry_event: {
    color: entryColor,
    icon: EnvelopeIcon,
    defaultTitle: "Listen for Event",
    defaultDescription:
      "Listens for an event to trigger the flow. Drop different actions and options here!",
    dataFields: [],
  },
  entry_error: {
    color: errorColor,
    icon: ExclamationCircleIcon,
    defaultTitle: "Error Handler",
    defaultDescription: "Handle errors that occur during execution.",
    dataFields: [],
  },
  action_response_text: {
    color: actionColor,
    icon: ArrowUturnLeftIcon,
    defaultTitle: "Text response",
    defaultDescription: "Bot replies with a plain text response",
    dataSchema: nodeActionDataSchema,
    dataFields: ["text_response", "custom_label"],
  },
  action_message_create: {
    color: actionColor,
    icon: ChatBubbleBottomCenterIcon,
    defaultTitle: "Create text message",
    defaultDescription: "Bot sends a plain text message to a channel",
    dataSchema: nodeActionDataSchema,
    dataFields: ["text_response", "custom_label"],
  },
  action_log: {
    color: actionColor,
    icon: DocumentTextIcon,
    defaultTitle: "Log Message",
    defaultDescription:
      "Log some text which is only visible in the deployment logs",
    dataSchema: nodeActionLogDataSchema,
    dataFields: ["log_level", "log_message", "custom_label"],
  },
  condition: {
    color: conditionColor,
    icon: ArrowsRightLeftIcon,
    defaultTitle: "Comparison Condition",
    defaultDescription:
      "Run actions based on the difference between two values.",
    dataFields: ["custom_label"],
  },
  option_text: {
    color: optionColor,
    icon: LanguageIcon,
    defaultTitle: "Text",
    defaultDescription: "Type in some plain text",
    dataSchema: nodeOptionDataSchema,
    dataFields: ["name", "description", "custom_label"],
  },
  option_number: {
    color: optionColor,
    icon: PlusIcon,
    defaultTitle: "Number",
    defaultDescription: "Type in a number",
    dataSchema: nodeOptionDataSchema,
    dataFields: ["name", "description", "custom_label"],
  },
  option_user: {
    color: optionColor,
    icon: UserIcon,
    defaultTitle: "User",
    defaultDescription: "Select a member from the server",
    dataSchema: nodeOptionDataSchema,
    dataFields: ["name", "description", "custom_label"],
  },
  option_channel: {
    color: optionColor,
    icon: HashtagIcon,
    defaultTitle: "Channel",
    defaultDescription: "Select a channel from the server",
    dataSchema: nodeOptionDataSchema,
    dataFields: ["name", "description", "custom_label"],
  },
  option_role: {
    color: optionColor,
    icon: TagIcon,
    defaultTitle: "Role",
    defaultDescription: "Select a role from the server",
    dataSchema: nodeOptionDataSchema,
    dataFields: ["name", "description", "custom_label"],
  },
  option_attachment: {
    color: optionColor,
    icon: DocumentIcon,
    defaultTitle: "Attachment",
    defaultDescription: "Upload a file",
    dataSchema: nodeOptionDataSchema,
    dataFields: ["name", "description", "custom_label"],
  },
};

const unknownNodeType: NodeValues = {
  color: "#ff0000",
  icon: QuestionMarkCircleIcon,
  defaultTitle: "Unknown",
  defaultDescription: "Unknown node type.",
  dataFields: [],
};

export function useNodeValues(nodeType: string): NodeValues {
  return useMemo(() => {
    const values = nodeTypes[nodeType];
    if (!values) {
      return unknownNodeType;
    }
    return values;
  }, [nodeType]);
}
