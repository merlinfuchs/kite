import { z } from "zod";
import { numericRegex, placeholderRegex } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

const maxRandomIdLength = 256;

const lengthMessage = "Must be a whole number, or a single {{ }} placeholder";

export const randomId: BlockDefinition = {
  type: "action_random_id",
  title: "Generate Random ID",
  description: "Generate a random ID or code",
  icon: "fingerprint",
  category: "Utilities",
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "random_id_type",
      schema: z
        .enum(["uuid_v4", "nanoid", "alphanumeric", "numeric", "hex"])
        .optional()
        .describe(
          "Format of the ID: uuid_v4, nanoid (letters, digits, - and _), alphanumeric, numeric or hex. Defaults to uuid_v4."
        ),
    },
    {
      name: "random_id_length",
      schema: z
        .string()
        .regex(numericRegex, lengthMessage)
        .refine(
          (v) => Number(v) >= 1 && Number(v) <= maxRandomIdLength,
          `Must be between 1 and ${maxRandomIdLength}`
        )
        .or(z.string().regex(placeholderRegex, lengthMessage))
        .optional()
        .describe(
          `Number of characters, from 1 to ${maxRandomIdLength}. Not used for uuid_v4. Defaults to 21 for nanoid, 16 for alphanumeric, 6 for numeric and 32 for hex.`
        ),
    },
  ],
  result: { schema: z.string().describe("The generated ID.") },
  run: { kind: "custom" },
};
