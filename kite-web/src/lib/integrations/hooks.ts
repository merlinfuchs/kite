import { useMemo } from "react";
import { getBlockDefinition, blockIntegrations } from "../blocks";
import { useAppIntegrations } from "../hooks/api";
import { FlowData } from "../flow/dataSchema";
import { getIntegration, integrations } from ".";
import { Integration } from "./types";

// The IDs of the integrations the app can use. Undefined while loading.
export function useEnabledIntegrationIds() {
  const states = useAppIntegrations();
  return useMemo(() => {
    if (!states) return undefined;
    return new Set(
      integrations
        .filter(
          (i) =>
            i.availability === "always" ||
            states.some((s) => s?.integration_id === i.id && s.enabled)
        )
        .map((i) => i.id)
    );
  }, [states]);
}

// The integrations a block needs that the app didn't enable.
export function useMissingIntegrations(nodeType: string | undefined) {
  const enabled = useEnabledIntegrationIds();
  return useMemo(() => {
    const block = getBlockDefinition(nodeType);
    if (!block || !enabled) return [];
    return blockIntegrations(block)
      .map((id) => getIntegration(id))
      .filter((i): i is Integration => !!i && !enabled.has(i.id));
  }, [nodeType, enabled]);
}

// The integrations the blocks of the flows need that the app didn't enable.
export function useMissingFlowIntegrations(flows: FlowData[]) {
  const enabled = useEnabledIntegrationIds();
  if (!enabled) return [];
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
