---
sidebar_position: 46.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Command Cooldown

<EmbedFlowNode type="option_command_cooldown" />

The `Command Cooldown` block limits how often your command can be used. While the command is on cooldown, it replies with a cooldown message instead of running the rest of the flow.

If the flow fails, the cooldown it started is cleared again, so a failed run doesn't use it up.

:::warning
Cooldowns are kept in memory and **reset when Kite restarts**, so they're limited to 1 hour. For longer cooldowns, like daily rewards, use stored variables as shown in the [cooldowns example](/examples/cooldowns).
:::

### Options

> `Scope` Who the cooldown applies to:
>
> - `User` Each user has their own cooldown. This is the default.
> - `Server` Everyone in the server shares one cooldown. In DMs this falls back to a per-user cooldown.
> - `Global` Everyone everywhere shares one cooldown.
>
> `Duration` How many seconds the cooldown lasts for, as a whole number from `1` to `3600` (1 hour). You can also use a single placeholder, like `{{arg('seconds')}}`. An empty, invalid, fractional or out-of-range duration makes the command fail with an error rather than skipping the cooldown.
>
> `Message` The message shown when someone uses the command while it's on cooldown. Use `{{var('cooldown_remaining')}}` to show how many seconds are left. Leave it empty to use the default message: _You're on cooldown. Try again in X seconds._

For longer or more flexible cooldowns built from stored variables, see the [cooldowns example](/examples/cooldowns).

<NodeInfoExplorer type="option_command_cooldown" />
