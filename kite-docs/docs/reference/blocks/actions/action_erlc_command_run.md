---
sidebar_position: 37.8
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Run ER:LC command

<EmbedFlowNode type="action_erlc_command_run" />

The `Run ER:LC command` block runs a command in your [ER:LC](https://apidocs.erlc.gg) private server, like `:h Server restart in 5 minutes` or `:kick {{ arg('player') }}`. Enable ER:LC under [Integrations](../../integrations.md) first, then use _Authorize Kite_ there once, as ER:LC only runs commands from apps the server owner authorized.

ER:LC runs one command every 5 seconds per server. If another command just ran, the block waits for its turn, and fails if that's more than 10 seconds away. The server needs players in it to run commands.

ER:LC's rules forbid spamming and using `:pm` instead of the in-game chat. Breaking them can cost the server its API access.

<NodeInfoExplorer type="action_erlc_command_run" />
