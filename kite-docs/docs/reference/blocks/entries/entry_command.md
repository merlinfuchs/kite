---
sidebar_position: 1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Command

<EmbedFlowNode type="entry_command" />

The `Command` block is the entry point for slash and context menu commands. This is where your command flow begins when someone invokes the command.

You can configure the command type and name directly in this block. Slash commands also need a description. Kite adds the command to Discord a few seconds after you save it.

For a User context menu command, use `interaction.command.target_user` to access the selected user. For a Message context menu command, use `interaction.command.target_message` to access the selected message and `interaction.command.target_message.author` for its author. After a button or modal resumes the flow, use `origin.interaction.command.target_user` or `origin.interaction.command.target_message`.

## Argument order

Once a command has two or more [Command Argument](../options/option_command_argument) blocks, the `Command` block lists them in the order Discord shows them. Use the arrows to move an argument up or down. Each argument block also shows its position in the corner.

Required arguments always come before optional ones, because Discord doesn't allow it the other way around. New arguments are added at the end.

:::note

The order is part of the command in Discord, so Discord shows the new order a few seconds after you save.

:::

<NodeInfoExplorer type="entry_command" />
