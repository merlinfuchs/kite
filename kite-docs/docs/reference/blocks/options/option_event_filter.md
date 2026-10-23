---
sidebar_position: 48
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Event Filter

<EmbedFlowNode type="option_event_filter" />

The `Event Filter` block allows you to filter events based on their properties. You can specify conditions that must be met for the event to trigger your flow.

You can filter on the message content, user ID, server ID, channel ID or message ID. Filtering on the message ID is useful for reaction events, for example to build reaction roles on a single message.

For [Audit Log Entry](../../event.md#audit-log-entry) listeners, filter on the **Audit Log Action** and pick the actions the flow runs for, like Member Banned and Member Kicked.

**Is One Of** and **Is Not One Of** compare with several values separated by commas, for example a few channel IDs.

This is useful for creating more specific event handlers that only respond to certain types of events or events with specific characteristics.

<NodeInfoExplorer type="option_event_filter" />
