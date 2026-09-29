import { Integration } from "../types";
import { inviteCreate } from "./blocks/inviteCreate";
import { messageBulkDelete } from "./blocks/messageBulkDelete";
import { messageList } from "./blocks/messageList";
import { roleCreate } from "./blocks/roleCreate";

// The app's bot token is its credential, so Discord is always connected.
export const discord: Integration = {
  id: "discord",
  name: "Discord",
  blocks: [messageList, messageBulkDelete, inviteCreate, roleCreate],
};
