import { describe, expect, it } from "vitest";
import {
  formatJsonTemplate,
  legacyJsonBodyToTemplate,
  scanJsonTemplate,
  validateJsonTemplate,
} from "./jsonTemplate";

describe("scanJsonTemplate", () => {
  it("knows which placeholders are inside strings", () => {
    const scan = scanJsonTemplate(
      '{"a": "hi {{user.name}}", "b": {{result(\'x\').status_code}}}'
    );
    expect(scan.placeholders.map((p) => [p.expression, p.inString])).toEqual([
      ["user.name", true],
      ["result('x').status_code", false],
    ]);
  });

  it("handles escaped quotes", () => {
    const scan = scanJsonTemplate('{"a": "x\\"", "b": {{num}}}');
    expect(scan.placeholders[0].inString).toBe(false);
  });

  it("ignores quotes inside placeholders", () => {
    const scan = scanJsonTemplate('{"a": {{arg("x")}}, "b": "{{y}}"}');
    expect(scan.placeholders.map((p) => p.inString)).toEqual([false, true]);
  });

  it("reports unclosed placeholders", () => {
    expect(scanJsonTemplate('{"a": "{{x"}').unclosedAt).toBe(7);
  });
});

describe("validateJsonTemplate", () => {
  it("accepts placeholders anywhere a value can go", () => {
    expect(
      validateJsonTemplate(
        '{"a": "{{x}}", "b": {{y}}, "c": [{{z}}, 1], "{{k}}": true}'
      )
    ).toBeNull();
  });

  it("accepts an empty body", () => {
    expect(validateJsonTemplate("  ")).toBeNull();
  });

  it("rejects invalid JSON", () => {
    expect(validateJsonTemplate('{"a": 1,}')).not.toBeNull();
    expect(validateJsonTemplate('{"a": }')).not.toBeNull();
  });

  it("rejects unclosed placeholders", () => {
    expect(validateJsonTemplate('{"a": {{x}')?.message).toMatch(/closing/);
  });
});

describe("formatJsonTemplate", () => {
  it("formats while keeping placeholders", () => {
    expect(
      formatJsonTemplate('{"a":"hi {{user.name}}","b":{{arg("n")}},"c":[1,2]}')
    ).toBe(
      '{\n  "a": "hi {{user.name}}",\n  "b": {{arg("n")}},\n  "c": [\n    1,\n    2\n  ]\n}'
    );
  });

  it("keeps $ in placeholders", () => {
    expect(formatJsonTemplate('{"a":{{x + "$&"}}}')).toBe(
      '{\n  "a": {{x + "$&"}}\n}'
    );
  });

  it("returns null for invalid JSON", () => {
    expect(formatJsonTemplate('{"a":')).toBeNull();
  });
});

describe("legacyJsonBodyToTemplate", () => {
  it("converts old object bodies", () => {
    expect(legacyJsonBodyToTemplate({ content: "{{user.name}}" })).toBe(
      '{\n  "content": "{{user.name}}"\n}'
    );
    expect(legacyJsonBodyToTemplate(null)).toBe("");
  });
});
