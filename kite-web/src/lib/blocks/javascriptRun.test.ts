import { describe, expect, it } from "vitest";
import { z } from "zod";
import { javascriptRun } from "./javascriptRun";

const schema = javascriptRun.fields![0].schema as z.ZodType;

describe("javascriptRun", () => {
  it("accepts an input without a value", () => {
    expect(
      schema.safeParse({ code: "return 1", inputs: [{ name: "x" }] }).success
    ).toBe(true);
  });

  it("marks the second input with a name that's taken", () => {
    const res = schema.safeParse({
      code: "return 1",
      inputs: [{ name: "x" }, { name: "x" }],
    });
    expect(res.error?.issues.map((i) => i.path)).toEqual([
      ["inputs", 1, "name"],
    ]);
  });

  it("rejects placeholders in the code, not other braces", () => {
    expect(schema.safeParse({ code: 'return "{{"' }).success).toBe(true);
    expect(schema.safeParse({ code: "return '{{user.id}}'" }).success).toBe(
      false
    );
  });

  it("accepts secrets, which fetch fills in", () => {
    const code = (s: string) => schema.safeParse({ code: s }).success;
    expect(code('await fetch("https://x.com/{{secrets.API_KEY}}")')).toBe(true);
    expect(code('"{{ secrets.API_KEY }}" + "{{user.id}}"')).toBe(false);
    expect(code('"{{secrets.API_KEY.length}}"')).toBe(false);
  });
});
