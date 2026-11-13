import { HTTPRequestData } from "../types/flow.gen";
import { placeholderRe, plainSecretRe } from "./templateSecrets";

// The expressions of the placeholders a request uses, which need a value when
// it's sent from the editor, as there's no event to fill them in. Kite fills in
// the ones that are just a secret.
export function httpRequestPlaceholders(data: HTTPRequestData): string[] {
  const templates = [
    data.url || "",
    ...(data.headers || []).map((h) => h.value),
    ...(data.query || []).map((q) => q.value),
  ];
  collectStrings(data.body_json, templates);

  const res: string[] = [];
  for (const template of templates) {
    for (const [, placeholder] of template.matchAll(placeholderRe)) {
      const expression = placeholder.trim();
      if (
        expression &&
        !plainSecretRe.test(expression) &&
        !res.includes(expression)
      ) {
        res.push(expression);
      }
    }
  }
  return res;
}

function collectStrings(value: unknown, res: string[]) {
  if (typeof value === "string") {
    res.push(value);
  } else if (Array.isArray(value)) {
    value.forEach((v) => collectStrings(v, res));
  } else if (value && typeof value === "object") {
    for (const [key, v] of Object.entries(value)) {
      res.push(key);
      collectStrings(v, res);
    }
  }
}

// A response body indented if it's JSON, so it's easier to read.
export function formatResponseBody(body: string): string {
  try {
    return JSON.stringify(JSON.parse(body), null, 2);
  } catch {
    return body;
  }
}
