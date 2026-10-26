---
sidebar_position: 38
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Comparison Condition

<EmbedFlowNode type="control_condition_compare" />

The `Comparison Condition` block allows you to create conditional logic based on comparing two values. You can set up multiple comparison conditions that will execute different actions based on the results.

This block creates a branching structure where different paths can be taken depending on the comparison results.

## Comparing with a list

To match several values in one branch, pick one of the list modes and put one value per line:

- **Is One Of** matches if the base value equals one of the values.
- **Is Not One Of** matches if it equals none of them.
- **Contains Any Of** matches if the base value contains any of them, for example to check a message for a list of words or a file name for a list of extensions.

Spaces around each value and empty lines are ignored, and so are spaces around the base value for **Is One Of** and **Is Not One Of**. Values aren't separated by commas, so a value can contain one. A placeholder that holds a list, like `{{result('words')}}` for the result of a Get Variable block, works too, and each of its items is one value.

## Ignoring case

With **Ignore Case** enabled, upper and lower case don't matter, so `gtg` also matches `GTG` and `Gtg`. It works with every mode that compares text. **Greater Than**, **Less Than** and the other number modes compare numbers and ignore it.

A branch's value can be a [secret](../../secrets.md) like `{{secrets.TOKEN}}` when it compares with **Equal** or **Not Equal** and **Ignore Case** is off, for example to check the token a webhook sends.

<NodeInfoExplorer type="control_condition_compare" />
