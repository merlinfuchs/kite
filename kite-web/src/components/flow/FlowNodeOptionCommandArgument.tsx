import {
  Node,
  Position,
  useStore,
  useUpdateNodeInternals,
} from "@xyflow/react";
import { useEffect } from "react";
import { NodeData, NodeProps } from "../../lib/flow/dataSchema";
import FlowNodeBase from "./FlowNodeBase";
import {
  autocompleteHandle,
  hasAutocompleteOutput,
  optionColor,
} from "@/lib/flow/nodes";
import { getCommandArguments } from "@/lib/flow/commandArguments";
import FlowNodeHandle from "./FlowNodeHandle";

export default function FlowNodeOptionCommandArgument(props: NodeProps) {
  // Selects a number, so the block doesn't re-render whenever another block
  // changes.
  const position = useStore((s) => {
    const nodes = s.nodes as Node<NodeData>[];

    for (const edge of s.edges) {
      if (edge.source !== props.id || edge.targetHandle != null) continue;

      const entry = nodes.find((n) => n.id === edge.target);
      if (entry?.type !== "entry_command") continue;

      const args = getCommandArguments(entry, nodes, s.edges);
      if (args.length < 2) return null;
      return args.findIndex((n) => n.id === props.id) + 1;
    }

    return null;
  });

  // React Flow only knows about handles it measured, so the autocomplete
  // output has to be measured when it appears.
  const autocomplete = hasAutocompleteOutput(props);
  const updateNodeInternals = useUpdateNodeInternals();
  useEffect(() => {
    updateNodeInternals(props.id);
  }, [autocomplete, props.id, updateNodeInternals]);

  return (
    <div className="relative">
      <FlowNodeBase
        {...props}
        title={props.data.name}
        description={props.data.description}
        showConnectedMarker={false}
      >
        {position !== null && (
          <div
            className="absolute -top-2 -left-2 h-4 min-w-4 px-1 rounded-full flex items-center justify-center text-white text-[10px] font-medium leading-4"
            style={{ backgroundColor: optionColor }}
            title="Position in Discord"
          >
            {position}
          </div>
        )}
        <FlowNodeHandle
          type="source"
          position={Position.Bottom}
          color={optionColor}
        />
      </FlowNodeBase>

      {autocomplete && (
        <div
          className="absolute top-1/2 left-full -translate-y-1/2 ml-3"
          title="Blocks connected here run while the user types in the argument"
        >
          <div className="relative px-2 py-1 shadow-md rounded bg-muted text-xs text-foreground whitespace-nowrap">
            Autocomplete
            <FlowNodeHandle
              type="source"
              position={Position.Right}
              id={autocompleteHandle}
              size="small"
            />
          </div>
        </div>
      )}
    </div>
  );
}
