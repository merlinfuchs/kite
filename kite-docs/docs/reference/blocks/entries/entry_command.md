---
sidebar_position: 1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Command

<EmbedFlowNode type="entry_command" />

The `Command` block is the entry point for commands. This is where your command flow begins when someone uses your command.

You can configure the type, name, and description of the command directly in this block. Kite adds the command to Discord a few seconds after you save it.

A User or Message command shows in the Apps menu of a user or message. It has no description, and `interaction.command.target_user` or `interaction.command.target_message` is what it was used on. See [User and Message commands](../../command.md#user-and-message-commands).

## Argument order

Once a command has two or more [Command Argument](../options/option_command_argument) blocks, the `Command` block lists them in the order Discord shows them. Use the arrows to move an argument up or down. Each argument block also shows its position in the corner.

Required arguments always come before optional ones, because Discord doesn't allow it the other way around. New arguments are added at the end.

:::note

The order is part of the command in Discord, so Discord shows the new order a few seconds after you save.

:::

<NodeInfoExplorer type="entry_command" />
