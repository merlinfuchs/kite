import { mkdirSync, writeFileSync } from "fs";
import { join } from "path";
import { getNodeInfo } from "../src/lib/flow/nodeInfo";
import { nodeTypes } from "../src/lib/flow/nodes";

// Static exports can't serve API routes, so write what the docs fetch from
// /api/flow/nodes/<type> into the export as files.
const dir = join(__dirname, "../out/api/flow/nodes");
mkdirSync(dir, { recursive: true });

for (const nodeType of Object.keys(nodeTypes)) {
  writeFileSync(
    join(dir, `${nodeType}.json`),
    JSON.stringify(getNodeInfo(nodeType))
  );
}
