import { describe, expect, it } from "vitest";
import { getReferencedSecretNames } from "./secrets";

describe("getReferencedSecretNames", () => {
  it("finds secrets in any setting", () => {
    const flow = {
      nodes: [
        {
          id: "1",
          type: "action_http_request",
          position: { x: 0, y: 0 },
          data: {
            http_request_data: {
              url: "https://example.com/?key={{secrets.API_KEY}}",
              headers: [{ key: "X", value: "{{ secrets.OTHER }}" }],
            },
          },
        },
      ],
      edges: [],
    };
    expect(getReferencedSecretNames(flow)).toEqual(["API_KEY", "OTHER"]);
  });
});
