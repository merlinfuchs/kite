import {
  Completion,
  CompletionContext,
  CompletionResult,
  CompletionSource,
  ifNotIn,
} from "@codemirror/autocomplete";
import {
  CodeParam,
  codeGlobals,
  codeOperations,
  costText,
} from "../integrations/sdk";

// What the code of a Run JavaScript block has besides its calls.
export interface CodeNames {
  // The names of its inputs, read as input.name.
  inputs: string[];
  // The names of its stored variables, for variables.get({ name }).
  variables: string[];
}

const word = /^[\w$]*$/;
// Not after a dot or in a longer name, so config.discord isn't Kite's.
const start = "(?<![\\w$.])";
const objectCallRe = new RegExp(
  `${start}([\\w$]+)\\.([\\w$]+)\\(\\s*\\{([^{}()]*)$`
);
const fetchInitRe = new RegExp(`${start}fetch\\([^(){}]*,\\s*\\{([^{}()]*)$`);
const memberRe = new RegExp(`${start}[\\w$]+\\.[\\w$]*$`);
const variableNameRe = new RegExp(
  `${start}variables\\.[\\w$]+\\(\\s*\\{[^{}()]*\\bname\\s*:\\s*["'\`]([\\w$]*)$`
);

// Completes the calls of Run JavaScript, their parameters, the inputs and
// stored variables of the block, and the globals the prelude defines, next
// to what the JavaScript language completes. Only the names of stored
// variables are completed in strings, and nothing in comments.
export function codeCompletions(names: () => CodeNames): CompletionSource[] {
  const variableNames = (context: CompletionContext) => {
    const m = variableNameRe.exec(before(context));
    if (!m) return null;
    return {
      from: context.pos - m[1].length,
      options: names().variables.map((label) => ({ label, type: "constant" })),
      validFor: word,
    };
  };
  const code = (context: CompletionContext) =>
    parameters(context) ?? members(context, names) ?? globalNames(context);
  return [
    variableNames,
    ifNotIn(["String", "TemplateString", "LineComment", "BlockComment"], code),
  ];
}

function before(context: CompletionContext) {
  return context.state.sliceDoc(Math.max(0, context.pos - 2000), context.pos);
}

// The parameters a call doesn't have yet, in its object, like
// discord.getChannel({ chan, or the options of fetch.
function parameters(context: CompletionContext): CompletionResult | null {
  const text = before(context);
  let params: CodeParam[] | undefined;
  let inside: string;
  const call = objectCallRe.exec(text);
  const init = fetchInitRe.exec(text);
  if (call) {
    const [, namespace, name] = call;
    inside = call[3];
    params = codeOperations().find(
      (o) => o.namespace === namespace && o.name === name
    )?.params;
  } else if (init) {
    inside = init[1];
    params = codeOperations()
      .find((o) => o.call === "http.request")
      ?.params.filter((p) => p.name !== "url");
  } else {
    return null;
  }

  // In a value, where members and globals are completed instead.
  const segment = inside.split(",").pop()!;
  if (!params || segment.includes(":")) return null;
  const used = new Set([...inside.matchAll(/([\w$]+)\s*:/g)].map((m) => m[1]));
  const typed = /[\w$]*$/.exec(segment)![0];
  return {
    from: context.pos - typed.length,
    options: params
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
function members(
  context: CompletionContext,
  names: () => CodeNames
): CompletionResult | null {
  const member = context.matchBefore(memberRe);
  if (!member) return null;
  const [namespace, typed] = member.text.split(".");
  const options: Completion[] =
    namespace === "input"
      ? names().inputs.map((label) => ({ label, type: "property" }))
      : codeOperations()
          .filter((op) => op.namespace === namespace)
          .map((op) => ({
            label: op.name,
            type: "function",
            detail: `(${op.params
              .filter((p) => p.required)
              .map((p) => p.name)
              .join(", ")})`,
            info: op.variableCredits
              ? op.description
              : `${op.description} ${costText(op.credits)}`,
          }));
  if (!options.length) return null;
  return { from: context.pos - typed.length, options, validFor: word };
}

function globalNames(context: CompletionContext): CompletionResult | null {
  const typed = context.matchBefore(/[\w$]+$/);
  if (!typed && !context.explicit) return null;
  const from = typed?.from ?? context.pos;
  if (context.state.sliceDoc(from - 1, from) === ".") return null;
  return { from, options: globals(), validFor: word };
}

let globalOptions: Completion[] | undefined;
function globals() {
  return (globalOptions ??= [
    ...codeGlobals().map((g) => ({
      label: g.name,
      type: g.kind,
      detail: g.detail,
      info: g.description,
    })),
    ...[...new Set(codeOperations().map((op) => op.namespace))].map(
      (label) => ({ label, type: "namespace" })
    ),
  ]);
}
