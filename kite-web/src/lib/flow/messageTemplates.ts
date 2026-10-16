import { FlowData } from "../types/flow.gen";
import { FlowFunction, Message, MessageCreateRequest } from "../types/wire.gen";
import {
  SharedFunction,
  functionToShared,
  getReferencedFunctionIds,
} from "./functions";

// A message template as it's stored in share codes and exported JSON. The id is
// the template's id in the app it was exported from, so flows in the same share
// can be pointed at the imported copy.
export type SharedMessage = MessageCreateRequest & { id: string };

export function messageToShared(message: Message): SharedMessage {
  // Attachments are assets of the original app, so they aren't exported.
  const data = { ...message.data };
  delete data.attachments;

  return {
    id: message.id,
    name: message.name,
    description: message.description,
    data,
    flow_sources: message.flow_sources ?? {},
  };
}

function getReferencedMessageIds(flow: FlowData | null | undefined) {
  const ids = new Set<string>();
  for (const node of flow?.nodes ?? []) {
    if (node.data?.message_template_id) {
      ids.add(node.data.message_template_id);
    }
  }
  return ids;
}

// Every template and function the flows use, including those used by the
// flows of those templates and functions, e.g. a function that sends a
// template whose button calls another function. Templates and functions in
// the exclude lists, which are already part of the share, are left out.
export function collectReferencedResources(
  flows: (FlowData | null | undefined)[],
  messages: (Message | undefined)[],
  functions: (FlowFunction | undefined)[],
  excludeMessageIds: string[] = [],
  excludeFunctionIds: string[] = []
): { messages: SharedMessage[]; functions: SharedFunction[] } {
  const messagesById = new Map(
    messages.flatMap((m) => (m ? [[m.id, m] as const] : []))
  );
  const functionsById = new Map(
    functions.flatMap((f) => (f ? [[f.id, f] as const] : []))
  );
  const seenMessages = new Set(excludeMessageIds);
  const seenFunctions = new Set(excludeFunctionIds);
  const res = {
    messages: [] as SharedMessage[],
    functions: [] as SharedFunction[],
  };

  const queue = [...flows];
  while (queue.length > 0) {
    const flow = queue.shift();

    for (const id of Array.from(getReferencedMessageIds(flow))) {
      if (seenMessages.has(id)) continue;
      seenMessages.add(id);

      const message = messagesById.get(id);
      if (!message) continue;
      res.messages.push(messageToShared(message));
      queue.push(...Object.values(message.flow_sources ?? {}));
    }

    for (const id of Array.from(getReferencedFunctionIds(flow))) {
      if (seenFunctions.has(id)) continue;
      seenFunctions.add(id);

      const fn = functionsById.get(id);
      if (!fn) continue;
      res.functions.push(functionToShared(fn));
      queue.push(fn.flow_source);
    }
  }

  return res;
}

export interface ReferenceMapping {
  variableIds: Set<string>;
  // Templates that exist in the current app and can be kept as they are.
  messageIds: Set<string>;
  // Templates from the original app and the id of their copy in this app.
  messageIdMap?: Map<string, string>;
  // Functions that exist in the current app. Without it, function
  // references are kept as they are.
  functionIds?: Set<string>;
  // Functions from the original app and the id of their copy in this app.
  functionIdMap?: Map<string, string>;
}

// Variable, message template and function IDs belong to the app the flow was
// exported from. Templates and functions that were imported along with the
// flow are pointed at their copy, everything else that doesn't exist in the
// current app is cleared, so the user picks it again in the editor.
export function remapFlowReferences(
  flow: FlowData,
  {
    variableIds,
    messageIds,
    messageIdMap,
    functionIds,
    functionIdMap,
  }: ReferenceMapping
) {
  let removed = 0;

  const nodes = flow.nodes.map((node) => {
    const data = { ...node.data };
    let cleared = false;

    if (data.variable_id && !variableIds.has(data.variable_id)) {
      delete data.variable_id;
      cleared = true;
    }

    const templateId = data.message_template_id;
    if (templateId) {
      const mapped = messageIdMap?.get(templateId);
      if (mapped) {
        data.message_template_id = mapped;
      } else if (!messageIds.has(templateId)) {
        delete data.message_template_id;
        cleared = true;
      }
    }

    const functionId = data.function_id;
    if (functionId && functionIds) {
      const mapped = functionIdMap?.get(functionId);
      if (mapped) {
        data.function_id = mapped;
      } else if (!functionIds.has(functionId)) {
        delete data.function_id;
        cleared = true;
      }
    }

    if (cleared) removed++;
    return { ...node, data };
  });

  return { flow: { ...flow, nodes }, removed };
}

export function remapFlowSources(
  flowSources: { [key: string]: FlowData } | null | undefined,
  mapping: ReferenceMapping
) {
  let removed = 0;
  const res: { [key: string]: FlowData } = {};
  for (const [key, flow] of Object.entries(flowSources ?? {})) {
    const remapped = remapFlowReferences(flow, mapping);
    res[key] = remapped.flow;
    removed += remapped.removed;
  }
  return { flowSources: res, removed };
}

export function isSharedMessage(value: unknown): value is SharedMessage {
  const m = value as SharedMessage | null | undefined;
  return (
    !!m &&
    typeof m === "object" &&
    typeof m.id === "string" &&
    typeof m.name === "string" &&
    m.name.length > 0 &&
    !!m.data &&
    typeof m.data === "object" &&
    (m.flow_sources == null ||
      (typeof m.flow_sources === "object" &&
        Object.values(m.flow_sources).every(
          (f) => Array.isArray(f?.nodes) && Array.isArray(f?.edges)
        )))
  );
}
