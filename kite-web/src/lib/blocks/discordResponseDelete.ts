import { responseTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordResponseDelete: BlockDefinition = {
  type: "action_response_delete",
  id_prefix: "response",
  title: "Delete response message",
  description: "Bot deletes an existing interaction response message",
  icon: "message-circle-x",
  category: "Responses",
  needs_interaction: true,
  requires: ["discord"],
  credits: 1,
  fields: [responseTargetSetting],
  run: { kind: "custom" },
};
