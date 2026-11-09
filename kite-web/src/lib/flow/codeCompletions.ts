import {
  Completion,
  CompletionContext,
  CompletionResult,
} from "@codemirror/autocomplete";
import { CodeOperation, codeOperations } from "../integrations/sdk";

// What the code of a Run JavaScript block has besides its calls.
export interface CodeNames {
  // The names of its inputs, read as input.name.
  inputs: string[];
  // The names of its stored variables, for variables.get({ name }).
  variables: string[];
}

let operations: CodeOperation[] | undefined;
function calls() {
  return (operations ??= codeOperations());
}

function cost(op: CodeOperation) {
  if (!op.credits) return "Costs nothing.";
  return op.credits === 1 ? "Costs 1 credit." : `Costs ${op.credits} credits.`;
}

const word = /^[\w$]*$/;

// Completes the calls of Run JavaScript, their parameters, the inputs and
// stored variables of the block, and the globals the prelude defines, next
// to what the JavaScript language completes.
export function codeCompletions(names: () => CodeNames) {
  return (context: CompletionContext): CompletionResult | null => {
    const { pos, state } = context;
    const before = state.sliceDoc(Math.max(0, pos - 2000), pos);

    // In the object of a call, like discord.getChannel({ chan
    const call = /([\w$]+)\.([\w$]+)\(\s*\{([^{}()]*)$/.exec(before);
    if (call) {
      const [, namespace, name, inside] = call;
      const segment = inside.split(",").pop()!;
      if (segment.includes(":")) {
        const variable = /\bname\s*:\s*["'`]([\w$]*)$/.exec(segment);
        if (namespace !== "variables" || !variable) return null;
        return {
          from: pos - variable[1].length,
          options: names().variables.map((label) => ({
            label,
            type: "constant",
          })),
          validFor: word,
        };
      }
      const op = calls().find(
        (o) => o.namespace === namespace && o.name === name
      );
      if (!op) return null;
      const used = new Set(
        [...inside.matchAll(/([\w$]+)\s*:/g)].map((m) => m[1])
      );
      const typed = /[\w$]*$/.exec(segment)![0];
      return {
        from: pos - typed.length,
        options: op.params
          .filter((p) => !used.has(p.name))
          .map((p) => ({
            label: p.name,
            type: "property",
            detail: `${p.required ? "" : "optional "}${p.type}`,
            info: p.description,
            apply: `${p.name}: `,
            boost: p.required ? 1 : 0,
          })),
        validFor: word,
      };
    }

    // A member, like discord.get or input.am
    const member = context.matchBefore(/[\w$]+\.[\w$]*$/);
    if (member) {
      const [namespace, typed] = member.text.split(".");
      let options: Completion[];
      if (namespace === "input") {
        options = names().inputs.map((label) => ({ label, type: "property" }));
      } else {
        options = calls()
          .filter((op) => op.namespace === namespace)
          .map((op) => ({
            label: op.name,
            type: "function",
            detail: `(${op.params
              .filter((p) => p.required)
              .map((p) => p.name)
              .join(", ")})`,
            info: `${op.description} ${cost(op)}`,
          }));
      }
      if (!options.length) return null;
      return { from: pos - typed.length, options, validFor: word };
    }

    const typed = context.matchBefore(/[\w$]+$/);
    if (!typed && !context.explicit) return null;
    const from = typed?.from ?? pos;
    if (state.sliceDoc(from - 1, from) === ".") return null;
    return { from, options: globals(), validFor: word };
  };
}

function globals(): Completion[] {
  const namespaces = [...new Set(calls().map((op) => op.namespace))];
  return [
    {
      label: "input",
      type: "variable",
      info: "The values of the block's inputs, by name.",
    },
    {
      label: "fetch",
      type: "function",
      detail: "(url, { method, headers, body })",
      info: "Sends a request to a URL. Costs 3 credits.",
    },
    {
      label: "formatDate",
      type: "function",
      detail: "(date, format, timeZone)",
      info: 'Formats a date with tokens like "dddd HH:mm" in a time zone like "Europe/Berlin".',
    },
    ...namespaces.map((label) => ({ label, type: "namespace" })),
  ];
}
