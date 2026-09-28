---
sidebar_position: 15.05
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Clear message reactions

<EmbedFlowNode type="action_message_reaction_clear" />

The `Clear message reactions` block is used to remove reactions from every user on a message. Set **Reactions** to `All reactions` to clear the whole message, or to `One emoji` and pick an emoji to only remove the reactions of that emoji.

Unlike [Delete message reaction](./action_message_reaction_delete.md), which only removes the bot's own reaction, this block removes the reactions of all users. The bot needs the `Manage Messages` permission in the channel.

<NodeInfoExplorer type="action_message_reaction_clear" />
