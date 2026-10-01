---
sidebar_position: 37.7
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Get ER:LC server

<EmbedFlowNode type="action_erlc_server_get" />

The `Get ER:LC server` block gets the status of your [ER:LC](https://apidocs.erlc.gg) private server, like its name, join code and player count. Enable ER:LC under [Integrations](../../integrations.md) first, with the server key from your private server's settings.

Turn on the parts you need, like players, staff, the queue or recent logs. ER:LC only returns those, and one block can get several at once. The result has them under their own names, so `{{ result('server').CurrentPlayers }}` is the number of players and `{{ result('server').Players }}` the list of them, if you included it.

<NodeInfoExplorer type="action_erlc_server_get" />
