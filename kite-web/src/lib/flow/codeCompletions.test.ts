import { CompletionContext, CompletionResult } from "@codemirror/autocomplete";
import { javascript } from "@codemirror/lang-javascript";
import { EditorState } from "@uiw/react-codemirror";
import { describe, expect, it } from "vitest";
import { codeCompletions } from "./codeCompletions";

const sources = codeCompletions(() => ({
  inputs: ["amount"],
  variables: ["xp"],
}));

function labels(code: string) {
  const state = EditorState.create({ doc: code, extensions: [javascript()] });
  const context = new CompletionContext(state, code.length, false);
  for (const source of sources) {
    const res = source(context) as CompletionResult | null;
    if (res) return res.options.map((o) => o.label);
  }
}

describe("code completions", () => {
  it("complete calls of a namespace", () => {
    expect(labels("await discord.getGuildMem")).toContain("getGuildMember");
    expect(labels("await variables.")).toEqual([
      "get",
      "set",
      "increment",
      "delete",
      "list",
    ]);
  });

  it("complete the parameters a call doesn't have yet", () => {
    const res = labels("await discord.getGuildMember({ guild_id: x, ");
    expect(res).toContain("user_id");
    expect(res).not.toContain("guild_id");
    // Values get members and globals.
    expect(labels("await discord.getChannel({ channel_id: inp")).toContain(
      "input"
    );
    expect(labels('variables.get({ name: "xp", scope: input.')).toEqual([
      "amount",
    ]);
    expect(labels('fetch(url, { method: "POST", he')).toEqual([
      "headers",
      "body",
    ]);
  });

  it("complete inputs and stored variables of the block", () => {
    expect(labels("return input.")).toEqual(["amount"]);
    expect(labels('await variables.get({ name: "')).toEqual(["xp"]);
  });

  it("complete globals, not members of other objects", () => {
    expect(labels("fet")).toContain("fetch");
    expect(labels("fet")).toContain("discord");
    expect(labels("foo.ba")).toBeUndefined();
    expect(labels("config.discord.")).toBeUndefined();
    expect(labels("res.input.")).toBeUndefined();
  });

  it("complete nothing in comments and strings, but variable names", () => {
    expect(labels("// fet")).toBeUndefined();
    expect(labels('return "Hi fet')).toBeUndefined();
    expect(labels('await variables.get({ name: "')).toEqual(["xp"]);
  });
});
