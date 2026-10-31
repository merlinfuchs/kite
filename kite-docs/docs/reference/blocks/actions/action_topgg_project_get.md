---
sidebar_position: 37.91
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Get Top.gg listing

<EmbedFlowNode type="action_topgg_project_get" />

The `Get Top.gg listing` block gets your bot's or server's listing on [Top.gg](https://top.gg), like its votes this month, its votes of all time and its review score. Enable Top.gg under [Integrations](../../integrations.md) first, with the API token of your listing.

`{{ result('topgg').votes }}` is the number of votes this month and `{{ result('topgg').votes_total }}` the number of all time, for example for a `/vote` command that shows both next to the link to vote.

<NodeInfoExplorer type="action_topgg_project_get" />
