---
sidebar_position: 3
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Button

<EmbedFlowNode type="entry_component_button" />

The `Button` block is the entry point for button and select menu interactions. This block gets triggered when a user clicks a button or picks an option in a select menu that was created by your bot. In the flow of a select menu the block is called `Select Menu`.

When a user picks an option in a select menu, the value of the option is available as `{{interaction.value}}`. If the select menu allows picking more than one option, all picked values are available as `{{interaction.values}}`.

This is typically used in conjunction with interactive components in your messages.

<NodeInfoExplorer type="entry_component_button" />
