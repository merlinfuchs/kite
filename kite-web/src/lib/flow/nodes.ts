import {
  ArrowsRightLeftIcon,
  ChatBubbleBottomCenterIcon,
  CommandLineIcon,
  EnvelopeIcon,
  LanguageIcon,
  QuestionMarkCircleIcon,
} from "@heroicons/react/24/solid";
import { ExoticComponent, useMemo } from "react";
import {
  nodeActionDataSchema,
  nodeCommandDataSchema,
  nodeOptionDataSchema,
} from "./data";
import { ZodSchema } from "zod";

export const primaryColor = "#5457f0";

export const actionColor = "#3b82f6";
export const entryColor = "#eab308";
export const conditionColor = "#22c55e";
export const optionColor = "#a855f7";

export interface NodeValues {
  color: string;
  icon: ExoticComponent<{ className: string }>;
  defaultTitle: string;
  defaultDescription: string;
  schema?: ZodSchema;
}

export const nodeTypes: Record<string, NodeValues> = {
  entry_command: {
    color: entryColor,
    icon: CommandLineIcon,
    defaultTitle: "Command",
    defaultDescription:
      "Command entry. Drop different actions and options here!",
    schema: nodeCommandDataSchema,
  },
  entry_event: {
    color: entryColor,
    icon: EnvelopeIcon,
    defaultTitle: "Listen for Event",
    defaultDescription:
      "Listens for an event to trigger the flow. Drop different actions and options here!",
  },
  action: {
    color: actionColor,
    icon: ChatBubbleBottomCenterIcon,
    defaultTitle: "Plain text response",
    defaultDescription: "Bot replies with a plain text response",
    schema: nodeActionDataSchema,
  },
  condition: {
    color: conditionColor,
    icon: ArrowsRightLeftIcon,
    defaultTitle: "Comparison Condition",
    defaultDescription:
      "Run actions based on the difference between two values.",
  },
  option: {
    color: optionColor,
    icon: LanguageIcon,
    defaultTitle: "Text",
    defaultDescription: "A plain text option",
    schema: nodeOptionDataSchema,
  },
};

const unknownNodeType: NodeValues = {
  color: "#ff0000",
  icon: QuestionMarkCircleIcon,
  defaultTitle: "Unknown",
  defaultDescription: "Unknown node type.",
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
