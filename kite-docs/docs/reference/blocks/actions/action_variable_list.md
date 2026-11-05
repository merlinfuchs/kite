---
sidebar_position: 31
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# List stored variable values

<EmbedFlowNode type="action_variable_list" />

The `List stored variable values` block gets the values of a stored variable in order. Use it for a leaderboard of the members with the most points, or to list the latest entries. It costs no credits.

## Order

- `Highest value first` and `Lowest value first` sort numbers, and text of a number like `42`. Values that aren't numbers come last.
- `Last updated first` and `First updated first` sort by when each value last changed.

## Scope starts with

For a scoped variable, the block only lists values whose scope starts with this text. Leave it empty to list all values.

Say you store points per member and server with the scope `{{guild.id}}:{{user.id}}`. Set `Scope starts with` to `{{guild.id}}:` and the block lists the members of the server the flow runs in.

## Result

The result is a list of up to 100 values, 10 if you don't set `Number of values`. Each value has a `scope`, the `value`, and `updated_at`, when it last changed as a Unix timestamp in seconds.

This placeholder turns the leaderboard from the example above into one line per member, like `1. @Alex 1240`:

```
{{join(map(result('values'), { string(#index + 1) + '. <@' + split(#.scope, ':')[1] + '> ' + string(#.value) }), '\n')}}
```

<NodeInfoExplorer type="action_variable_list" />
