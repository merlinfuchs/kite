import { useMemo } from "react";
import { getBlockDefinition, blockIntegrations } from "../blocks";
import { useAppIntegrations } from "../hooks/api";
import { FlowData } from "../flow/dataSchema";
import { getIntegration } from ".";
import { enabledIntegrationIds, missingFlowIntegrations } from "./missing";
import { Integration } from "./types";

// The IDs of the integrations the app can use. Undefined while loading.
export function useEnabledIntegrationIds() {
  const states = useAppIntegrations();
  return useMemo(
    () => (states ? enabledIntegrationIds(states) : undefined),
    [states]
  );
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
  return missingFlowIntegrations(flows, enabled);
}
