import { responseTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordResponseDelete: BlockDefinition = {
  type: "action_response_delete",
  title: "Delete response message",
  description: "Bot deletes an existing interaction response message",
  icon: "message-circle-x",
  category: "Responses",
  requires: ["discord"],
  credits: 1,
  fields: [responseTargetSetting],
  run: { kind: "custom" },
};
