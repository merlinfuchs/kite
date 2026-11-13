---
sidebar_position: 34.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Lookup value

<EmbedFlowNode type="action_value_lookup" />

The `Lookup value` block finds a value in a table of keys and values. It compares the value you give it, like `{{message.content}}`, with the key of each row and returns the value of the first row that matches. Use it instead of a condition with a branch for every key when each branch would do the same thing with a different text, then use `{{result('lookup_id')}}` in a single block after it.

Rows match when the value equals the key, contains it, or starts with it, depending on the match mode. Turn on ignore case to match `Hello` with `hello`. Keys and values can use placeholders, and you can paste many rows at once as `key = value` lines or two columns copied from a spreadsheet.

When no row matches, the result is the default value, or empty if there is none. Check for an empty result with a [comparison condition](../controls/control_condition_compare.md) to handle that case.

<NodeInfoExplorer type="action_value_lookup" />
