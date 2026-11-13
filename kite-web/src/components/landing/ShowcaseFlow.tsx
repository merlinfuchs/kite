import FlowExample from "@/components/flow/FlowExample";
import { nodeTypes } from "@/lib/flow/components";
import { Edge, Node } from "@xyflow/react";
import DiscordReplyNode from "./DiscordReplyNode";

const showcaseNodeTypes = { ...nodeTypes, discord_reply: DiscordReplyNode };

export default function ShowcaseFlow({
  nodes,
  edges,
}: {
  nodes: Node[];
  edges: Edge[];
}) {
  return (
    <FlowExample nodes={nodes} edges={edges} nodeTypes={showcaseNodeTypes} />
  );
}
