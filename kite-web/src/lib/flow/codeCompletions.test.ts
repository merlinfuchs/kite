import { CompletionContext } from "@codemirror/autocomplete";
import { EditorState } from "@uiw/react-codemirror";
import { describe, expect, it } from "vitest";
import { codeCompletions } from "./codeCompletions";

const complete = codeCompletions(() => ({
  inputs: ["amount"],
  variables: ["xp"],
}));

function labels(code: string) {
  const state = EditorState.create({ doc: code });
  const res = complete(new CompletionContext(state, code.length, false));
  return res?.options.map((o) => o.label);
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
    expect(
      labels("await discord.getChannel({ channel_id: inp")
    ).toBeUndefined();
  });

  it("complete inputs and stored variables of the block", () => {
    expect(labels("return input.")).toEqual(["amount"]);
    expect(labels('await variables.get({ name: "')).toEqual(["xp"]);
  });

  it("complete globals, not members of other objects", () => {
    expect(labels("fet")).toContain("fetch");
    expect(labels("fet")).toContain("discord");
    expect(labels("foo.ba")).toBeUndefined();
  });
});
