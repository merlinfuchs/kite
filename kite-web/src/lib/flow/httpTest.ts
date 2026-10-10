import { HTTPRequestData } from "../types/flow.gen";

const PLACEHOLDER_REGEX = /{{([\s\S]*?)}}/g;
const CALL_REGEX = /\b(arg|input|var|result)\(\s*(['"])([^'"]+)\2\s*\)/g;
const NODES_REGEX = /(^|[^\w.])nodes\.(\w+)/g;
// A dotted path that isn't part of a longer expression like
// result('x').data.id or a string literal like 'a.b'.
const PATH_REGEX = /(^|[^\w.'")\]])([A-Za-z_]\w*(?:\.\w+)+)(\s*\()?/g;

// The request fields that can hold placeholders.
function requestTemplates(request?: HTTPRequestData): string[] {
  if (!request) return [];

  return [
    request.url,
    ...(request.query || []).map((q) => q.value),
    ...(request.headers || []).map((h) => h.value),
    request.body,
  ].filter((v): v is string => !!v);
}

/**
 * Finds the values a test of the request needs, keyed the way the service
 * expects them: user.id, arg('name'), input('id'), var('name') or
 * result('node'). Keys are returned in the order they first appear.
 */
export function getHttpTestValueKeys(request?: HTTPRequestData): string[] {
  const keys: string[] = [];
  const add = (key: string) => {
    if (!keys.includes(key)) keys.push(key);
  };

  for (const template of requestTemplates(request)) {
    for (const [, expression] of template.matchAll(PLACEHOLDER_REGEX)) {
      for (const [, fn, , name] of expression.matchAll(CALL_REGEX)) {
        add(`${fn}('${name}')`);
      }

      for (const [, , id] of expression.matchAll(NODES_REGEX)) {
        add(`result('${id}')`);
      }

      for (const [, , path, call] of expression.matchAll(PATH_REGEX)) {
        const parts = path.split(".");
        // Nodes are covered above and secrets come from the app.
        if (parts[0] === "nodes" || parts[0] === "secrets") continue;
        // user.name.toUpper() only needs user.name.
        if (call) parts.pop();
        if (parts.length < 2) continue;
        add(parts.join("."));
      }
    }
  }

  return keys;
}

const SECRET_REGEX = /\bsecrets\.([A-Za-z_]\w*)/g;

/** The app secrets the request uses, which a test sends but never shows. */
export function getHttpTestSecretNames(request?: HTTPRequestData): string[] {
  const names: string[] = [];
  for (const template of requestTemplates(request)) {
    for (const [, expression] of template.matchAll(PLACEHOLDER_REGEX)) {
      for (const [, name] of expression.matchAll(SECRET_REGEX)) {
        if (!names.includes(name)) names.push(name);
      }
    }
  }
  return names;
}

// Test values only live for the session, so they survive closing and
// reopening the dialog without ending up in the saved flow.
const sessionTestValues = new Map<string, Record<string, string>>();

export function getSessionHttpTestValues(
  nodeId: string
): Record<string, string> {
  return sessionTestValues.get(nodeId) || {};
}

export function setSessionHttpTestValues(
  nodeId: string,
  values: Record<string, string>
) {
  sessionTestValues.set(nodeId, values);
}

/** Pretty-prints a body when it's JSON, otherwise returns it unchanged. */
export function formatHttpTestBody(body: string): {
  text: string;
  json: boolean;
} {
  const trimmed = body.trim();
  if (!trimmed.startsWith("{") && !trimmed.startsWith("[")) {
    return { text: body, json: false };
  }

  try {
    return { text: JSON.stringify(JSON.parse(trimmed), null, 2), json: true };
  } catch {
    return { text: body, json: false };
  }
}

export function formatHttpTestSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}
