---
sidebar_position: 19.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Prune members

<EmbedFlowNode type="action_member_prune" />

The `Prune members` block is used to remove inactive members from a server. It removes every member that hasn't been active for the set number of days, which has to be between 1 and 30. The number of removed members is stored as the result of the block.

Members with any role are never removed, so only members without roles that have been inactive are affected.

:::warning
Pruning can't be undone. Make sure only trusted people can run the command, for example with the `Command Permissions` block.
:::

:::info
The bot needs the `Kick Members` permission in the server.
:::

<NodeInfoExplorer type="action_member_prune" />
