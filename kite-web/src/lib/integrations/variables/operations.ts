import { z } from "zod";
import { variableListSorts } from "../../blocks/variableList";
import { HandlerOperationDefinition } from "../handlerOperations";

// The stored variables the Run JavaScript block lets the code use, by the
// names under its Stored variables.
const name = z
  .string()
  .describe("Name of the variable under Stored variables of the block.");
const scope = z
  .string()
  .optional()
  .describe(
    "Scope of the value, like a user ID for one value per user. Required for scoped variables, ignored for others."
  );
const value = z
  .unknown()
  .describe("The value: text, a number, a list, an object, or null.");

export const variablesOperations: HandlerOperationDefinition[] = [
  {
    id: "get",
    description: "Get the value of a stored variable, or null if it has none.",
    credits: 0,
    params: z.object({ name, scope }).strict(),
    result: value,
  },
  {
    id: "set",
    description: "Set the value of a stored variable, and get it back.",
    credits: 1,
    params: z.object({ name, scope, value }).strict(),
    result: value,
  },
  {
    id: "append",
    description:
      "Add an item to the end of the list in a stored variable, or text to the end of other values, and get the new value. Set the variable to an empty list first to start a list. Calls at the same time don't overwrite each other.",
    credits: 1,
    params: z.object({ name, scope, value }).strict(),
    result: value,
  },
  {
    id: "prepend",
    description: "Like append, at the start of the list or the text.",
    credits: 1,
    params: z.object({ name, scope, value }).strict(),
    result: value,
  },
  {
    id: "increment",
    description:
      "Add to the number in a stored variable, which starts at 0, and get the new value. Calls at the same time don't overwrite each other, unlike get and set.",
    credits: 1,
    params: z
      .object({
        name,
        scope,
        by: z
          .number()
          .optional()
          .describe("How much to add, negative to subtract. Defaults to 1."),
      })
      .strict(),
    result: z.number(),
  },
  {
    id: "delete",
    description: "Delete the value of a stored variable.",
    credits: 1,
    params: z.object({ name, scope }).strict(),
    result: z.null(),
    destructive: true,
  },
  {
    id: "list",
    description:
      "List the values of a stored variable, like a leaderboard of the highest ones.",
    credits: 1,
    params: z
      .object({
        name,
        scope_prefix: z
          .string()
          .optional()
          .describe(
            "Only values whose scope starts with it, like the server ID for scopes like <server ID>:<user ID>."
          ),
        sort: z
          .enum(variableListSorts.map((s) => s.value))
          .optional()
          .describe("Defaults to value_desc, the highest values first."),
        limit: z
          .number()
          .int()
          .min(1)
          .max(100)
          .optional()
          .describe("How many values, from 1 to 100. Defaults to 10."),
        skip: z
          .number()
          .int()
          .min(0)
          .max(10000)
          .optional()
          .describe("How many values to skip first, for the next page."),
      })
      .strict(),
    result: z.array(
      z.object({
        scope: z.string().nullable(),
        value,
        updated_at: z.number().describe("Unix timestamp in seconds."),
      })
    ),
  },
];
