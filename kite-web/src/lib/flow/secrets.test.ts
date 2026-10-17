import { describe, expect, it } from "vitest";
import { getReferencedSecretNames } from "./secrets";

describe("getReferencedSecretNames", () => {
  it("finds secrets in placeholders of requests", () => {
    const flow = {
      nodes: [
        {
          id: "1",
          type: "action_http_request",
          position: { x: 0, y: 0 },
          data: {
            http_request_data: {
              url: "https://example.com/secrets.txt?key={{secrets.API_KEY}}",
              headers: [{ key: "X", value: "{{ secrets.OTHER }}" }],
            },
          },
        },
        {
          id: "2",
          type: "action_log",
          position: { x: 0, y: 0 },
          data: { log_message: "{{secrets.LOGGED}}" },
        },
      ],
      edges: [],
    };
    expect(getReferencedSecretNames(flow)).toEqual(["API_KEY", "OTHER"]);
  });
});
