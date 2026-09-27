---
sidebar_position: 26
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Bulk delete messages

<EmbedFlowNode type="action_message_bulk_delete" />

The `Bulk delete messages` block deletes the most recent messages in a channel in a single action.

> `Target Channel` The channel to delete messages in.
>
> `Amount` How many of the most recent messages to delete, up to 1000.
>
> `Ignore pinned messages` When enabled, pinned messages are skipped.

Messages older than two weeks cannot be bulk deleted, so they are skipped instead of failing the block. The block returns `deleted` (how many messages were actually deleted) and `failed` (how many could not be deleted), which you can store in a temporary variable.

Example: clear the last 50 messages in a channel when a `/purge` command runs.

<NodeInfoExplorer type="action_message_bulk_delete" />
