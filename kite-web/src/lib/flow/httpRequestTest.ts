import { HTTPRequestData } from "../types/flow.gen";

const placeholderRe = /\{\{([\s\S]*?)\}\}/g;

// The expressions of the placeholders a request uses, which need a value when
// it's sent from the editor, as there's no event to fill them in. Secrets are
// filled in by Kite.
export function httpRequestPlaceholders(data: HTTPRequestData): string[] {
  const templates = [
    data.url || "",
    ...(data.headers || []).map((h) => h.value),
    ...(data.query || []).map((q) => q.value),
  ];
  collectStrings(data.body_json, templates);

  const res: string[] = [];
  for (const template of templates) {
    for (const match of template.matchAll(placeholderRe)) {
      const expression = match[1].trim();
      if (
        expression &&
        !/\bsecrets\./.test(expression) &&
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
