import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const randomPick: BlockDefinition = {
  type: "action_random_pick",
  id_prefix: "pick",
  title: "Pick Random Item",
  description: "Pick a random item from a list",
  icon: "shuffle",
  category: "Utilities",
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "random_items",
      schema: templated(
        z.string().min(1).max(10000),
        "Items to pick from, one per line. Empty lines are ignored."
      ),
    },
  ],
  result: { schema: z.string().describe("The picked item.") },
  run: { kind: "custom" },
};
