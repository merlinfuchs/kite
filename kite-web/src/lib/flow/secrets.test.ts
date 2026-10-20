import { describe, expect, it } from "vitest";
import { getNodeValues } from "./nodes";
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

describe("comparison condition with secrets", () => {
  const schema = getNodeValues("control_condition_item_compare").dataSchema!;

  it("only allows secrets with equal or not equal", () => {
    for (const [mode, valid] of [
      ["equal", true],
      ["not_equal", true],
      ["starts_with", false],
      ["contains", false],
    ] as const) {
      const result = schema.safeParse({
        condition_item_mode: mode,
        condition_item_value: "{{secrets.TOKEN}}",
      });
      expect(result.success, mode).toBe(valid);
    }

    expect(
      schema.safeParse({
        condition_item_mode: "starts_with",
        condition_item_value: "{{webhook.data.secrets.count}}",
      }).success
    ).toBe(true);
  });

  it("only allows secrets on their own", () => {
    for (const [value, valid] of [
      ["{{secrets.TOKEN}}", true],
      ["Bearer {{ secrets.TOKEN }}", true],
      ["{{secrets.TOKEN[0:1]}}", false],
      ["{{secrets.TOKEN}}{{secrets?.TOKEN[0:1]}}", false],
    ] as const) {
      const result = schema.safeParse({
        condition_item_mode: "equal",
        condition_item_value: value,
      });
      expect(result.success, value).toBe(valid);
    }
  });
});
