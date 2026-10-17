// Helpers for JSON documents that contain {{ }} placeholders, like the body of
// the Send API Request block. They mirror eval.EvalJSONTemplate in the service:
// a placeholder inside a string is inserted as escaped text, a placeholder
// outside a string is inserted as a typed JSON value.

export interface JsonTemplatePlaceholder {
  // Offsets into the template, end is exclusive and includes the closing }}.
  start: number;
  end: number;
  expression: string;
  inString: boolean;
}

export interface JsonTemplateScan {
  placeholders: JsonTemplatePlaceholder[];
  // Offset of a {{ without a matching }}.
  unclosedAt?: number;
}

export function scanJsonTemplate(template: string): JsonTemplateScan {
  const placeholders: JsonTemplatePlaceholder[] = [];

  let inString = false;
  let escaped = false;

  for (let i = 0; i < template.length; ) {
    if (template.startsWith("{{", i)) {
      const end = template.indexOf("}}", i + 2);
      if (end === -1) {
        return { placeholders, unclosedAt: i };
      }

      placeholders.push({
        start: i,
        end: end + 2,
        expression: template.slice(i + 2, end),
        inString,
      });
      escaped = false;
      i = end + 2;
      continue;
    }

    const ch = template[i];
    if (inString) {
      if (escaped) escaped = false;
      else if (ch === "\\") escaped = true;
      else if (ch === '"') inString = false;
    } else if (ch === '"') {
      inString = true;
    }
    i++;
  }

  return { placeholders };
}

// Replaces every placeholder with something of the same length that is valid
// JSON in its position, so the document can be parsed and error offsets still
// point at the right place.
function maskPlaceholders(template: string, scan: JsonTemplateScan) {
  let res = "";
  let last = 0;
  for (const p of scan.placeholders) {
    const length = p.end - p.start;
    res += template.slice(last, p.start);
    res += p.inString ? "x".repeat(length) : "1" + "0".repeat(length - 1);
    last = p.end;
  }
  return res + template.slice(last);
}

export interface JsonTemplateError {
  message: string;
  from: number;
  to: number;
}

// Returns why the template won't produce valid JSON, or null if it will.
export function validateJsonTemplate(
  template: string
): JsonTemplateError | null {
  if (!template.trim()) return null;

  const scan = scanJsonTemplate(template);
  if (scan.unclosedAt !== undefined) {
    return {
      message: "Placeholder is missing its closing }}",
      from: scan.unclosedAt,
      to: Math.min(scan.unclosedAt + 2, template.length),
    };
  }

  try {
    JSON.parse(maskPlaceholders(template, scan));
    return null;
  } catch (e) {
    const message = e instanceof Error ? e.message : String(e);
    const pos = errorPosition(message, template);
    return {
      message: `Invalid JSON: ${message}`,
      from: pos,
      to: Math.min(pos + 1, template.length),
    };
  }
}

function errorPosition(message: string, template: string) {
  const match = /position (\d+)/.exec(message);
  if (match) return Math.min(parseInt(match[1]), template.length);

  const lineMatch = /line (\d+) column (\d+)/.exec(message);
  if (lineMatch) {
    const lines = template.split("\n");
    const line = parseInt(lineMatch[1]) - 1;
    let pos = 0;
    for (let i = 0; i < line && i < lines.length; i++) {
      pos += lines[i].length + 1;
    }
    return Math.min(pos + parseInt(lineMatch[2]) - 1, template.length);
  }

  return template.length;
}

// Pretty prints the template while keeping its placeholders as they are.
// Returns null if the template isn't valid.
export function formatJsonTemplate(template: string): string | null {
  if (!template.trim()) return "";

  const scan = scanJsonTemplate(template);
  if (scan.unclosedAt !== undefined) return null;

  // Unique tokens that can't clash with the rest of the document.
  const nonce = Math.random().toString(36).slice(2, 10);
  const token = (i: number) => `__kite_${nonce}_${i}__`;

  let masked = "";
  let last = 0;
  scan.placeholders.forEach((p, i) => {
    masked += template.slice(last, p.start);
    masked += p.inString ? token(i) : `"${token(i)}"`;
    last = p.end;
  });
  masked += template.slice(last);

  let formatted: string;
  try {
    formatted = JSON.stringify(JSON.parse(masked), null, 2);
  } catch {
    return null;
  }

  scan.placeholders.forEach((p, i) => {
    const original = template.slice(p.start, p.end);
    const search = p.inString ? token(i) : `"${token(i)}"`;
    // A function replacement, so $ in the placeholder isn't special.
    formatted = formatted.replace(search, () => original);
  });

  return formatted;
}

// Converts the object body of blocks created before body types existed into
// the text the editor works with.
export function legacyJsonBodyToTemplate(body: unknown): string {
  if (body === undefined || body === null) return "";
  return JSON.stringify(body, null, 2);
}
