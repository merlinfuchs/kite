import { Edge, Node } from "reactflow";
import { FlatFile } from "../code/filetree";
import { FlowData } from "./data";

export function bundleFlowFiles(files: FlatFile[]): FlowData {
  const nodes: Node[] = [];
  const edges: Edge[] = [];

  for (const file of files) {
    if (file.path.endsWith(".flow")) {
      // TODO: validate with zod
      const data: FlowData = JSON.parse(file.content);
      nodes.push(...data.nodes);
      edges.push(...data.edges);
    }
  }

  return { nodes, edges };
}
