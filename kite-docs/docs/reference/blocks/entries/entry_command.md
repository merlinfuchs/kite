---
sidebar_position: 1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Command

<EmbedFlowNode type="entry_command" />

The `Command` block is the entry point for commands. This is where your command flow begins when a user invokes it.

You can configure the command type, name, and description directly in this block. The command will be automatically registered with Discord when you deploy your app.

The command type can be a **Slash Command** (invoked with `/`), a **User Context Menu** command (right-click a user), or a **Message Context Menu** command (right-click a message). Context menu commands have only a name, and expose the right-clicked user or message as `command.target`.

<NodeInfoExplorer type="entry_command" />
