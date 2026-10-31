---
sidebar_position: 37.92
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Post Top.gg server count

<EmbedFlowNode type="action_topgg_metrics_update" />

The `Post Top.gg server count` block shows how many servers your bot is in on its [Top.gg](https://top.gg) listing. Enable Top.gg under [Integrations](../../integrations.md) first, with the API token of your listing.

Set the server count to `{{app.guild_count}}` and put the block in a [scheduled event listener](../../event.md#scheduled-event-listeners) that runs every hour or so. Leave the shard count empty, Top.gg only shows it when it's set.

<NodeInfoExplorer type="action_topgg_metrics_update" />
