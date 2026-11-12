import { describe, expect, it } from "vitest";
import { remapFlowReferences } from "./messageTemplates";
import { nodeVariableIds } from "./variableIds";

describe("variable IDs", () => {
  const flow = {
    nodes: [
      {
        id: "1",
        position: { x: 0, y: 0 },
        data: { variable_id: "kept" },
      },
      {
        id: "2",
        position: { x: 0, y: 0 },
        data: {
          javascript_data: {
            code: "return 1",
            variables: [
              { name: "a", variable_id: "kept" },
              { name: "b", variable_id: "gone" },
            ],
          },
        },
      },
    ],
    edges: [],
  };

  it("include those of the code", () => {
    expect(flow.nodes.flatMap((n) => nodeVariableIds(n.data))).toEqual([
      "kept",
      "kept",
      "gone",
    ]);
  });

  it("are cleared on import when the app doesn't have them", () => {
    const res = remapFlowReferences(flow, {
      variableIds: new Set(["kept"]),
      messageIds: new Set(),
    });
    expect(res.removed).toBe(1);
    expect(res.flow.nodes[0].data.variable_id).toBe("kept");
    expect(res.flow.nodes[1].data.javascript_data?.variables).toEqual([
      { name: "a", variable_id: "kept" },
      { name: "b", variable_id: undefined },
    ]);
  });
});
