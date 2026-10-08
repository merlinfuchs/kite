import { FlowData } from "../types/flow.gen";
import { Message, MessageCreateRequest } from "../types/wire.gen";

// A message template as it's stored in share codes and exported JSON. The id is
// the template's id in the app it was exported from, so flows in the same share
// can be pointed at the imported copy.
export type SharedMessage = MessageCreateRequest & { id: string };

export function messageToShared(message: Message): SharedMessage {
  return {
    id: message.id,
    name: message.name,
    description: message.description,
    data: message.data,
    flow_sources: message.flow_sources ?? {},
  };
}

export function getReferencedMessageIds(flow: FlowData | null | undefined) {
  const ids = new Set<string>();
  for (const node of flow?.nodes ?? []) {
    if (node.data?.message_template_id) {
      ids.add(node.data.message_template_id);
    }
  }
  return ids;
}

// Every template the flows use, including templates used by the buttons and
// select menus of those templates. Templates in excludeIds are left out.
export function collectReferencedMessages(
  flows: (FlowData | null | undefined)[],
  messages: (Message | undefined)[],
  excludeIds: string[] = []
): SharedMessage[] {
  const byId = new Map(
    messages.flatMap((m) => (m ? [[m.id, m] as const] : []))
  );
  const seen = new Set(excludeIds);
  const res: SharedMessage[] = [];

  const queue = flows.flatMap((f) => Array.from(getReferencedMessageIds(f)));
  while (queue.length > 0) {
    const id = queue.shift()!;
    if (seen.has(id)) continue;
    seen.add(id);

    const message = byId.get(id);
    if (!message) continue;

    res.push(messageToShared(message));
    for (const flow of Object.values(message.flow_sources ?? {})) {
      queue.push(...Array.from(getReferencedMessageIds(flow)));
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
}

// Variable and message template IDs belong to the app the flow was exported
// from. Templates that were imported along with the flow are pointed at their
// copy, everything else that doesn't exist in the current app is cleared.
export function remapFlowReferences(
  flow: FlowData,
  { variableIds, messageIds, messageIdMap }: ReferenceMapping
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

export function flowSourcesReference(
  flowSources: { [key: string]: FlowData } | null | undefined,
  ids: Set<string>
) {
  return Object.values(flowSources ?? {}).some((flow) =>
    Array.from(getReferencedMessageIds(flow)).some((id) => ids.has(id))
  );
}

export function isSharedMessage(value: unknown): value is SharedMessage {
  const m = value as SharedMessage | null | undefined;
  return (
    !!m &&
    typeof m === "object" &&
    typeof m.name === "string" &&
    m.name.length > 0 &&
    !!m.data &&
    typeof m.data === "object"
  );
}
