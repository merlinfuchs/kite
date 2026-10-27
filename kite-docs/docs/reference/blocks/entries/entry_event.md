---
sidebar_position: 2
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Listen for Event

<EmbedFlowNode type="entry_event" />

The `Listen for Event` block is the entry point for event-triggered flows. This block listens for specific Discord events and triggers your flow when they occur.

You can configure which event type to listen for and add filters to control when the flow should be triggered.

You can also give the event listener a **Name**, like `Honeypot Softban`. It's shown instead of the event in your list of event listeners, in bundles and on the Marketplace, so several listeners for the same event are easy to tell apart. Without a name, the event is shown, like `Message Create`.

<NodeInfoExplorer type="entry_event" />
