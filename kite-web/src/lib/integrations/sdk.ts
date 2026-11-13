import { ApiParam, operationLabel } from "./api";
import { allOperations, Operation } from "./operations";

// What the code of the Run JavaScript block can call, as TypeScript, for the
// editor's completions, the MCP server and the kite.d.ts of the docs.

// cookie_api -> cookieApi, the reverse of snakeName in kite-service/pkg/js.
export function camelName(name: string) {
  return name.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
}

export interface CodeOperation {
  // The full name, like discord.getChannel.
  call: string;
  // The namespace code calls it in, like discord.
  namespace: string;
  // Its name in the namespace, like getChannel.
  name: string;
  description: string;
  credits: number;
  // The type of the argument and of what the promise resolves to.
  params: CodeParam[];
  result: string;
}

export interface CodeParam {
  name: string;
  type: string;
  required: boolean;
  description?: string;
}

let operations: CodeOperation[] | undefined;

// The operations code can call, see OperationPolicy.code.
export function codeOperations(): CodeOperation[] {
  return (operations ??= allOperations()
    .filter((op) => op.policy.code)
    .map(codeOperation));
}

function codeOperation(op: Operation): CodeOperation {
  const namespace = camelName(op.integration);
  const name = camelName(op.id);
  const base = {
    call: `${namespace}.${name}`,
    namespace,
    name,
    credits: op.credits,
  };
  if (op.kind === "handler") {
    const params = op.params as JSONSchema;
    return {
      ...base,
      description: op.description,
      params: Object.entries(params.properties ?? {}).map(([name, p]) => ({
        name,
        type: schemaType(p),
        required: !!params.required?.includes(name),
        description: p.description,
      })),
      result: schemaType(op.result as JSONSchema),
    };
  }
  return {
    ...base,
    description: `${operationLabel(op.id)}, ${op.method} ${op.path}.`,
    params: [...op.path_params, ...op.query_params].map((p) => ({
      name: p.name,
      type: paramType(p),
      required: p.required || op.path_params.includes(p),
    })),
    // Its JSON, which api.json doesn't describe.
    result: "any",
  };
}

function paramType(p: ApiParam) {
  if (p.enum) return p.enum.map((v) => JSON.stringify(v)).join(" | ");
  switch (p.type) {
    case "snowflake":
    case "string":
      return "string";
    case "integer":
    case "number":
      return "number";
    case "boolean":
      return "boolean";
    case "array":
      return "(string | number)[]";
  }
  return "any";
}

// The part of JSON Schema z.toJSONSchema gives for the operations.
interface JSONSchema {
  type?: string;
  description?: string;
  enum?: unknown[];
  properties?: Record<string, JSONSchema>;
  required?: string[];
  items?: JSONSchema;
  additionalProperties?: JSONSchema | boolean;
  anyOf?: JSONSchema[];
}

function schemaType(s: JSONSchema): string {
  if (s.enum) return s.enum.map((v) => JSON.stringify(v)).join(" | ");
  if (s.anyOf) return s.anyOf.map(schemaType).join(" | ");
  switch (s.type) {
    case "string":
    case "number":
    case "boolean":
    case "null":
      return s.type;
    case "integer":
      return "number";
    case "array":
      return `${wrap(schemaType(s.items ?? {}))}[]`;
    case "object": {
      if (s.properties) {
        const fields = Object.entries(s.properties).map(
          ([name, p]) =>
            `${name}${s.required?.includes(name) ? "" : "?"}: ${schemaType(p)}`
        );
        return `{ ${fields.join("; ")} }`;
      }
      if (typeof s.additionalProperties === "object") {
        return `Record<string, ${schemaType(s.additionalProperties)}>`;
      }
      return "Record<string, any>";
    }
  }
  // Any value, like that of a stored variable.
  return "any";
}

function wrap(type: string) {
  return type.includes("|") ? `(${type})` : type;
}

export function costText(credits: number) {
  if (!credits) return "Costs nothing.";
  return credits === 1 ? "Costs 1 credit." : `Costs ${credits} credits.`;
}

// What the prelude of kite-service/pkg/js defines besides the calls.
export interface CodeGlobal {
  name: string;
  kind: "variable" | "function";
  // Its parameters, for completions.
  detail?: string;
  description: string;
  // Its TypeScript declaration.
  declaration: string;
}

export function codeGlobals(): CodeGlobal[] {
  const request = codeOperations().find((op) => op.call === "http.request");
  return [
    {
      name: "input",
      kind: "variable",
      description: "The values of the block's inputs, by name.",
      declaration: "declare const input: Record<string, any>;",
    },
    {
      name: "fetch",
      kind: "function",
      detail: "(url, { method, headers, body })",
      description: `Sends a request to a URL with http.request. ${costText(
        request?.credits ?? 0
      )}`,
      declaration: [
        "declare function fetch(url: string, init?: {",
        // fetch takes them in any case, like in a browser.
        `  method?: ${fetchMethods(request)};`,
        "  headers?: Record<string, string> | [string, string][];",
        "  body?: string;",
        "}): Promise<{",
        "  ok: boolean;",
        "  status: number;",
        "  statusText: string;",
        "  url: string;",
        "  headers: { get(name: string): string | null; has(name: string): boolean };",
        "  text(): Promise<string>;",
        "  json(): Promise<any>;",
        "}>;",
      ].join("\n"),
    },
    {
      name: "formatDate",
      kind: "function",
      detail: "(date, format, timeZone)",
      description:
        'Formats a date with tokens like "dddd HH:mm" in a time zone like "Europe/Berlin".',
      declaration:
        "declare function formatDate(date?: Date | number | string, format?: string, timeZone?: string): string;",
    },
  ];
}

// The methods of http.request in upper and lower case, like "POST" | "post".
function fetchMethods(request: CodeOperation | undefined) {
  const type = request?.params.find((p) => p.name === "method")?.type;
  if (!type) return "string";
  const methods = type.split(" | ");
  return [...methods, ...methods.map((m) => m.toLowerCase())].join(" | ");
}

// The call as a TypeScript method, with what it does and costs.
export function codeSignature(op: CodeOperation, indent = "") {
  const params = op.params.map(
    (p) =>
      `${indent}  ${p.description ? `/** ${p.description} */ ` : ""}${p.name}${
        p.required ? "" : "?"
      }: ${p.type};`
  );
  const optional = op.params.every((p) => !p.required) ? "?" : "";
  return [
    `${indent}/** ${op.description} ${costText(op.credits)} */`,
    op.params.length
      ? `${indent}${op.name}(arg${optional}: {\n${params.join(
          "\n"
        )}\n${indent}}): Promise<${op.result}>;`
      : `${indent}${op.name}(): Promise<${op.result}>;`,
  ].join("\n");
}

// Declarations of everything code can use besides JavaScript itself, which
// the docs offer as kite.d.ts for editors.
export function codeDeclarations() {
  const namespaces = new Map<string, CodeOperation[]>();
  for (const op of codeOperations()) {
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
  return parts.join("\n") + "\n";
}
