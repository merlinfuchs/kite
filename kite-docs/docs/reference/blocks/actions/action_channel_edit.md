---
sidebar_position: 26
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit channel

<EmbedFlowNode type="action_channel_edit" />

The `Edit channel` block is used to edit a channel in a chosen category.

### Options

> `Name` The name for the channel.
>
> `Topic` The topic for the channel.
>
> `Category` The category that the channel will be edited in.
>
> `Position` The position for the channel.
>
> `Slowmode` The cooldown in seconds users must wait between sending messages, from `0` to `21600` (6 hours). `0` turns slowmode off. Leave it empty to keep the channel's current slowmode. Placeholders are supported. Not available for announcement and category channels.
>
> `Archive Thread` Only for threads. Archives or unarchives the thread. A new message unarchives it unless it's locked. `Keep current` leaves it as it is.
>
> `Lock Thread` Only for threads. Only moderators can unarchive a locked thread. Archive and lock a thread to close it, like a ticket, and keep its messages. Discord doesn't allow other changes to an archived thread, so unarchive it in the same block to change it.

### Permission Overwrites

Here you can choose which roles and users will have access to the channel. Leave this blank to allow all roles and users to access the channel.

<NodeInfoExplorer type="action_channel_edit" />
