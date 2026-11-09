import { getBlockDefinition } from "@/lib/blocks";
import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { getAvailablePlaceholders } from "@/lib/flow/placeholders";
import { useAppSecrets } from "@/lib/hooks/api";
import { useSyncedParameters } from "@/lib/flow/parametersStore";
import { Node, useStoreApi } from "@xyflow/react";
import { VariableIcon } from "lucide-react";
import { useMemo, useState } from "react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";

export default function FlowPlaceholderExplorer({
  onSelect,
  hideBrackets,
  field,
  onlyParameters,
}: {
  onSelect: (value: string) => void;
  hideBrackets?: boolean;
  // For settings that aren't templates, like the value of an event filter.
  onlyParameters?: boolean;
  // The setting it inserts into, like http_request_data.url. The app's
  // secrets are only offered for settings that can use them.
  field?: string;
}) {
  const [open, setOpen] = useState(false);
  const store = useStoreApi();
  const contextType = useFlowContext((c) => c.type);
  const secrets = useAppSecrets();
  const parameters = useSyncedParameters();

  const placeholders = useMemo(() => {
    if (!open) return [];

    const parameterGroup = parameters.length
      ? [
          {
            label: "Parameters",
            placeholders: parameters.map((p) => ({
              label: p.label || p.key,
              value: `params.${p.key}`,
            })),
          },
        ]
      : [];
    if (onlyParameters) return parameterGroup;

    const { nodes, edges } = store.getState();
    const typedNodes = nodes as Node<NodeData>[];
    const selected = typedNodes.find((n) => n.selected);
    const isResponseTarget =
      field === "message_target" &&
      !!selected?.type?.startsWith("action_response_");
    const allowsSecrets = !!getBlockDefinition(selected?.type)?.fields?.find(
      (f) => f.name === field?.split(".")[0]
    )?.secrets;

    const res = [
      ...getAvailablePlaceholders(
        selected?.id,
        typedNodes,
        edges,
        contextType,
        isResponseTarget
      ),
      ...parameterGroup,
    ];

    const names = allowsSecrets
      ? secrets?.flatMap((s) => (s ? [s.name] : [])) ?? []
      : [];
    if (names.length === 0) return res;

    return [
      ...res,
      {
        label: "Secrets",
        placeholders: names.map((name) => ({
          label: name,
          value: `secrets.${name}`,
        })),
      },
    ];
  }, [open, store, field, contextType, secrets, parameters, onlyParameters]);

  if (onlyParameters && !parameters.length) return null;

  return (
    <div className="absolute top-1.5 right-1.5 z-20">
      <PlaceholderExplorer
        onSelect={onSelect}
        onOpenChange={setOpen}
        placeholders={placeholders}
        hideBrackets={hideBrackets}
      >
        <VariableIcon
          className="h-5.5 w-5.5 text-muted-foreground hover:text-foreground cursor-pointer"
          role="button"
        />
      </PlaceholderExplorer>
    </div>
  );
}
