import { describe, expect, it } from "vitest";
import { getReferencedSecretNames } from "./secrets";

describe("getReferencedSecretNames", () => {
  it("finds secrets in placeholders of requests and condition values", () => {
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
        {
          id: "3",
          type: "control_condition_compare",
          position: { x: 0, y: 0 },
          data: { condition_base_value: "{{secrets.BASE}}" },
        },
        {
          id: "4",
          type: "control_condition_item_compare",
          position: { x: 0, y: 0 },
          data: { condition_item_value: "{{secrets.TOKEN}}" },
        },
      ],
      edges: [],
    };
    expect(getReferencedSecretNames(flow)).toEqual([
      "API_KEY",
      "OTHER",
      "TOKEN",
    ]);
  });
});
