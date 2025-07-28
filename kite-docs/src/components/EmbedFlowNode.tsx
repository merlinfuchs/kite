import EmbedBase from "./EmbedBase";

export default function EmbedFlowNode({ type }: { type: string }) {
  return <EmbedBase src={`/embed/flow/nodes/${type}`} />;
}
