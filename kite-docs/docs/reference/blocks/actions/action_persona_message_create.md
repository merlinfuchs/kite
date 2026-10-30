---
sidebar_position: 9.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Send message as

<EmbedFlowNode type="action_persona_message_create" />

The `Send message as` block sends a message to a channel or thread with another name and avatar than your bot's, like a character in a roleplay or a message on behalf of a user. Use `{{user.display_name}}` and `{{user.avatar_url}}` to send it as the user who ran the command. The upload button next to the avatar URL uploads an image to Kite and fills in its URL. This block requires Premium.

The message is sent through a webhook your bot creates in the channel the first time, so your bot needs the `Manage Webhooks` permission there. Buttons and select menus on the message work like on other messages.

Discord doesn't allow names that contain `discord` or `clyde`. The [Edit channel message](./action_message_edit.md) block can't edit messages sent with this block, as only the webhook can.

<NodeInfoExplorer type="action_persona_message_create" />
