---
sidebar_position: 1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Command

<EmbedFlowNode type="entry_command" />

The `Command` block is the entry point for slash commands. This is where your command flow begins when a user invokes your slash command.

You can configure the command name and description directly in this block. The command will be automatically registered with Discord when you deploy your app.

## Argument Order

Once a command has two or more [Command Argument](../options/option_command_argument) blocks, the `Command` block lists them in the order Discord shows them. Use the arrows to move an argument up or down. Each argument block also shows its position in the corner.

Required arguments always come before optional ones, because Discord doesn't allow it the other way around. New arguments are added at the end.

:::note

The order is part of the command in Discord, so you need to deploy again after changing it.

:::

<NodeInfoExplorer type="entry_command" />
