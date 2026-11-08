import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

// Match MaxCodeLength and MaxInputs in kite-service/pkg/js/js.go.
const maxCodeLength = 4000;
export const javascriptMaxInputs = 20;

// parameterKeyPattern, which blocks can't import as parameters.ts imports
// them, and ParameterKeyPattern in kite-service/pkg/flow/params.go.
const inputNamePattern = /^[A-Za-z_][A-Za-z0-9_]*$/;

export const javascriptRun: BlockDefinition = {
  type: "action_javascript_run",
  id_prefix: "js",
  title: "Run JavaScript",
  description: "Run a short piece of JavaScript and use what it returns later",
  icon: "code",
  category: "Utilities",
  credits: 10,
  strict_settings: true,
  fields: [
    {
      name: "javascript_data",
      schema: z
        .object({
          code: z
            .string()
            .min(1)
            .max(maxCodeLength)
            .refine((code) => !/\{\{[^{}]+\}\}/.test(code), {
              message:
                "Placeholders don't work in code. Add an input and read it as input.name instead.",
            })
            .describe(
              `JavaScript that runs as the body of a function and returns the result, like: return input.items.length. Inputs are properties of input. Placeholders in the code aren't filled in. There's no async, network, timers or Intl. formatDate(date, format, timeZone) formats a date with tokens like "dddd HH:mm" in a time zone like "Europe/Berlin". It can run for 25 ms and use 8 MB of memory. Up to ${maxCodeLength} characters.`
            ),
          inputs: z
            .array(
              z.object({
                name: z
                  .string()
                  .regex(inputNamePattern, {
                    message:
                      "Start with a letter or _ and use letters, digits and _",
                  })
                  .max(32)
                  .describe(
                    "Name of the input, read in the code as input.name."
                  ),
                value: templated(
                  z.string().optional(),
                  "Value of the input, like {{arg('amount')}}. A single placeholder keeps its type, like a number or a list. Users, members, channels, roles and messages become objects with the fields of their placeholders, like input.user.display_name, but without user.role_names, permissions, is_admin, is_owner, top_role and color, the category of a channel, and the counts and channels of a server. {{guild}} and {{channel}} are the ID of the server and channel the flow runs in, so pass the fields you need. A response of a Send API Request block has status_code, body and data."
                ),
              })
            )
            .max(javascriptMaxInputs)
            .superRefine((inputs, ctx) => {
              const names = new Set<string>();
              inputs.forEach((input, i) => {
                if (names.has(input.name)) {
                  ctx.addIssue({
                    code: "custom",
                    message: `Another input is named ${input.name} already`,
                    path: [i, "name"],
                  });
                }
                names.add(input.name);
              });
            })
            .optional()
            .describe("Values the code gets from the flow."),
        })
        .describe("The code to run and the values it gets."),
    },
  ],
  result: {
    schema: z
      .unknown()
      .describe("What the code returns, like a number, text, list or object."),
  },
  run: { kind: "custom" },
};
