import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const log: BlockDefinition = {
  type: "action_log",
  title: "Log Message",
  description: "Log some text which is only visible in the application logs",
  icon: "scroll-text",
  category: "Utilities",
  credits: 1,
  allow_unknown_settings: true,
  fields: [
    {
      name: "log_level",
      type: "string",
      schema: z
        .enum(["debug", "info", "warn", "error"])
        .describe("Severity of the log entry."),
    },
    {
      name: "log_message",
      type: "string",
      schema: templated(
        z.string().max(2000).min(1),
        "Text to write to the app's logs."
      ),
    },
  ],
  run: { kind: "custom" },
};
