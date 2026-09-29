import { useMemo } from "react";
import { getBlockDefinition, blockIntegrations } from "../blocks";
import { useAppIntegrations } from "../hooks/api";
import { getIntegration, integrations, needsCredential } from ".";
import { Integration } from "./types";

// The IDs of the integrations the app can use: those that are always
// connected and those the app connected. Undefined while loading.
export function useConnectedIntegrationIds() {
  const connected = useAppIntegrations();
  return useMemo(() => {
    if (!connected) return undefined;
    const ids = new Set(connected.map((c) => c?.integration_id));
    return new Set(
      integrations
        .filter((i) => !needsCredential(i) || ids.has(i.id))
        .map((i) => i.id)
    );
  }, [connected]);
}

// The integrations a block needs that the app didn't connect.
export function useMissingIntegrations(nodeType: string | undefined) {
  const connected = useConnectedIntegrationIds();
  return useMemo(() => {
    const block = getBlockDefinition(nodeType);
    if (!block || !connected) return [];
    return blockIntegrations(block)
      .map((id) => getIntegration(id))
      .filter((i): i is Integration => !!i && !connected.has(i.id));
  }, [nodeType, connected]);
}
