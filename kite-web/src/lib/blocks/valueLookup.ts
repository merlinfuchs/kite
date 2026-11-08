import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { LookupRow } from "../types/flow.gen";
import { BlockDefinition } from "./types";

// Limits of the service, see LookupRow in pkg/flow/data.go.
export const lookupMaxRows = 500;
const lookupMaxKeyLength = 500;
const lookupMaxTextLength = 2000;

export const lookupModes = [
  { value: "equal", label: "Equal" },
  { value: "contains", label: "Contains" },
  { value: "starts_with", label: "Starts With" },
] as const;

type LookupMode = (typeof lookupModes)[number]["value"];

export const valueLookup: BlockDefinition = {
  type: "action_value_lookup",
  id_prefix: "lookup",
  title: "Lookup Value",
  description: "Find the value for a key in a table",
  icon: "table",
  category: "Utilities",
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "lookup_value",
      schema: templated(
        z.string().max(lookupMaxTextLength),
        "The value to look up, e.g. {{message.content}}."
      ),
    },
    {
      name: "lookup_mode",
      schema: z
        .enum(lookupModes.map((m) => m.value) as [LookupMode, ...LookupMode[]])
        .optional()
        .describe(
          "How a row's key matches the value: equal, contains (the value contains the key) or starts_with (the value starts with the key). Defaults to equal."
        ),
    },
    {
      name: "lookup_ignore_case",
      schema: z
        .boolean()
        .optional()
        .describe("Whether upper and lower case don't matter."),
    },
    {
      name: "lookup_rows",
      schema: z
        .array(
          z.object({
            key: templated(
              z.string().min(1, "Required").max(lookupMaxKeyLength),
              "Key that is compared with the value."
            ),
            value: templated(
              z.string().max(lookupMaxTextLength),
              "Result of the block when the key matches."
            ),
          })
        )
        .min(1, "Add at least one row")
        .max(lookupMaxRows)
        .describe(
          `The table, up to ${lookupMaxRows} rows. The first row whose key matches wins.`
        ),
    },
    {
      name: "lookup_default",
      schema: templated(
        z.string().max(lookupMaxTextLength).optional(),
        "Result when no row matches. Without it the result is empty."
      ),
    },
  ],
  result: {
    schema: z
      .string()
      .describe(
        "The value of the matching row, or the default value if none matches."
      ),
  },
  run: { kind: "custom" },
};

// Rows pasted as lines of "key = value", or of a key and value separated by a
// tab, like two columns copied from a spreadsheet. A line with neither is a key
// without a value.
export function parseLookupLines(text: string): LookupRow[] {
  return text
    .split(/\r?\n/)
    .filter((line) => line.trim())
    .map((line) => {
      if (line.includes("\t")) {
        const [key, value = ""] = line.split("\t");
        return { key: key.trim(), value: value.trim() };
      }
      const separator = /\s*=\s*/.exec(line);
      if (!separator) return { key: line.trim(), value: "" };
      return {
        key: line.slice(0, separator.index).trim(),
        value: line.slice(separator.index + separator[0].length).trim(),
      };
    });
}
