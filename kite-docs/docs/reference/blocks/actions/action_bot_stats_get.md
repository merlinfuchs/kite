---
sidebar_position: 43
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Get bot stats

<EmbedFlowNode type="action_bot_stats_get" />

The `Get bot stats` block returns numbers about your app as a whole, for commands like `/botinfo` or `/ping`. It has no settings.

| Value          | Description                                                                 |
| -------------- | --------------------------------------------------------------------------- |
| `guild_count`  | The number of servers your app is in.                                       |
| `member_count` | The members of all those servers added up.                                  |
| `uptime`       | The number of seconds since your app last connected to Discord.             |
| `connected_at` | When your app last connected to Discord, as a Unix timestamp in seconds.    |
| `latency`      | The number of milliseconds Discord took to answer your app's last heartbeat. |

Later blocks read these values from the block's result:

```
In {{result('stats').guild_count}} servers with {{result('stats').member_count}} members
Online since <t:{{result('stats').connected_at}}:R>, ping {{result('stats').latency}}ms
```

The member count uses Discord's approximate count of each server, and a user who is in several of your app's servers is counted once for each.

The uptime starts over whenever Kite reconnects your app to Discord, for example after you change its token. The latency is `0` for up to a minute after connecting, until the first heartbeat was answered.

<NodeInfoExplorer type="action_bot_stats_get" />
