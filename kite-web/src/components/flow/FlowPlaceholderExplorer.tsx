import { getBlockDefinition } from "@/lib/blocks";
import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { getAvailablePlaceholders } from "@/lib/flow/placeholders";
import { useAppSecrets } from "@/lib/hooks/api";
import { syncParameters } from "@/lib/flow/parameters";
import { useFlowParameters } from "@/lib/flow/parametersStore";
import { Node, useEdges, useNodes } from "@xyflow/react";
import { VariableIcon } from "lucide-react";
import { useMemo } from "react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";

export default function FlowPlaceholderExplorer({
  onSelect,
  hideBrackets,
  field,
}: {
  onSelect: (value: string) => void;
  hideBrackets?: boolean;
  // The setting it inserts into, like http_request_data.url. The app's
  // secrets are only offered for settings that can use them.
  field?: string;
}) {
  const nodes = useNodes<Node<NodeData>>();
  const edges = useEdges();
  const contextType = useFlowContext((c) => c.type);
  const selected = nodes.find((n) => n.selected);

  const allowsSecrets = !!getBlockDefinition(selected?.type)?.fields?.find(
    (f) => f.name === field?.split(".")[0]
  )?.secrets;
  const secrets = useAppSecrets();
  const storedParameters = useFlowParameters();

  // TODO: only compute when explorer is open
  const placeholders = useMemo(() => {
    let res = getAvailablePlaceholders(selected?.id, nodes, edges, contextType);
    const parameters = syncParameters({ nodes }, storedParameters);
    if (parameters.length) {
      res = [
        ...res,
        {
          label: "Parameters",
          placeholders: parameters.map((p) => ({
            label: p.label || p.key,
            value: `params.${p.key}`,
          })),
        },
      ];
    }

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
  }, [
    selected?.id,
    nodes,
    edges,
    contextType,
    allowsSecrets,
    secrets,
    storedParameters,
  ]);

  return (
    <div className="absolute top-1.5 right-1.5 z-20">
      <PlaceholderExplorer
        onSelect={onSelect}
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
