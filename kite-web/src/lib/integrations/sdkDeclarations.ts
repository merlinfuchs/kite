import discordTypeDeclarations from "./discord/types.json";
import {
  CodeOperation,
  codeGlobals,
  codeOperations,
  codeSignature,
} from "./sdk";

// The declarations of the types in the discord namespace, by name, generated
// from Discord's spec by scripts/discord-api.mjs. Apart from sdk.ts, so the
// editor doesn't load them.
const discordTypes: Record<string, string> = discordTypeDeclarations;

// The names of the discord types a type uses, like Message in
// discord.Message[].
function typeNames(type: string) {
  return [...type.matchAll(/discord\.([A-Za-z0-9_]+)/g)].map((m) => m[1]);
}

// The declarations of the discord types the calls use, with the types those
// use in turn if all is set.
export function codeTypeDeclarations(ops: CodeOperation[], all = true) {
  const names = new Set<string>();
  const add = (type: string, deep: boolean) => {
    for (const name of typeNames(type)) {
      if (names.has(name) || !discordTypes[name]) continue;
      names.add(name);
      if (deep) add(discordTypes[name], deep);
    }
  };
  for (const op of ops) {
    for (const p of op.params) add(p.type, all);
    add(op.result, all);
  }
  return [...names].sort().map((name) => discordTypes[name]);
}

// Declarations of everything code can use besides JavaScript itself, which
// the docs offer as kite.d.ts for editors.
export function codeDeclarations() {
  const calls = codeOperations();
  const namespaces = new Map<string, CodeOperation[]>();
  for (const op of calls) {
    namespaces.set(op.namespace, [...(namespaces.get(op.namespace) ?? []), op]);
  }

  const parts = [
    "// What the code of Kite's Run JavaScript block can use besides JavaScript",
    "// itself. See https://docs.kite.onl/reference/blocks/actions/action_javascript_run",
  ];
  for (const global of codeGlobals()) {
    parts.push("", `/** ${global.description} */`, global.declaration);
  }
  for (const [namespace, ops] of namespaces) {
    parts.push(
      "",
      `declare const ${namespace}: {`,
      ops.map((op) => codeSignature(op, "  ")).join("\n"),
      "};"
    );
  }

  // The types of Discord's API, like discord.MessageResponse, in a namespace
  // that shares the name of the calls.
  const types = codeTypeDeclarations(calls);
  parts.push(
    "",
    "/** The objects of Discord's API that its calls take and return. */",
    "declare namespace discord {",
    types
      .map((t) =>
        t
          .split("\n")
          .map((line) => `  ${line}`)
          .join("\n")
      )
      .join("\n"),
    "}"
  );
  return parts.join("\n") + "\n";
}
