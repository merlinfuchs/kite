import EmbedBase from "./EmbedBase";

export default function EmbedFlowNode({ type }: { type: string }) {
  return <EmbedBase src={`http://localhost:3000/embed/flow/nodes/${type}`} />;
}
