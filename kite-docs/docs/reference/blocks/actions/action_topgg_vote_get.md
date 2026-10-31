---
sidebar_position: 37.9
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Check Top.gg vote

<EmbedFlowNode type="action_topgg_vote_get" />

The `Check Top.gg vote` block checks whether a user voted for your bot or server on [Top.gg](https://top.gg). Enable Top.gg under [Integrations](../../integrations.md) first, with the API token of your listing.

`{{ result('vote').voted }}` is `true` while the user has an active vote and `false` if they never voted or can vote again, so a comparison condition can lock a command to voters. `{{ result('vote').expires_at }}` is when they can vote again, and `{{ result('vote').weight }}` is `2` for votes on weekends, when Top.gg counts them twice.

Leave the user as `{{user.id}}` to check the user who ran the command. To react to votes as they happen instead, use a webhook event listener, see the [Top.gg guide](../../../guides/topgg.md).

<NodeInfoExplorer type="action_topgg_vote_get" />
