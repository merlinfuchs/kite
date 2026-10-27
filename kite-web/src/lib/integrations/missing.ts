import { blockIntegrations, getBlockDefinition } from "../blocks";
import { FlowData } from "../flow/dataSchema";
import { AppIntegration } from "../types/wire.gen";
import { getIntegration, integrations } from ".";
import { Integration } from "./types";

// The IDs of the integrations the app can use, by the states of its
// integrations.
export function enabledIntegrationIds(
  states: (AppIntegration | null | undefined)[]
) {
  return new Set(
    integrations
      .filter(
        (i) =>
          i.availability === "always" ||
          states.some((s) => s?.integration_id === i.id && s.enabled)
      )
      .map((i) => i.id)
  );
}

// The integrations the blocks of the flows need that aren't enabled.
export function missingFlowIntegrations(
  flows: FlowData[],
  enabled: Set<string>
) {
  const ids = new Set(
    flows.flatMap((f) =>
      f.nodes.flatMap((n) => {
        const block = getBlockDefinition(n.type);
        return block ? blockIntegrations(block) : [];
      })
    )
  );
  return Array.from(ids)
    .map((id) => getIntegration(id))
    .filter((i): i is Integration => !!i && !enabled.has(i.id));
}
