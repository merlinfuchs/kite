---
sidebar_position: 38
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Comparison Condition

<EmbedFlowNode type="control_condition_compare" />

The `Comparison Condition` block allows you to create conditional logic based on comparing two values. You can set up multiple comparison conditions that will execute different actions based on the results.

This block creates a branching structure where different paths can be taken depending on the comparison results.

A branch's value can be a [secret](../../secrets.md) like `{{secrets.TOKEN}}` when it compares with **Equal** or **Not Equal**, for example to check the token a webhook sends.

<NodeInfoExplorer type="control_condition_compare" />
