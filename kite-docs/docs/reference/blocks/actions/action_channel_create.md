---
sidebar_position: 26
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create channel

<EmbedFlowNode type="action_channel_create" />

The `Create channel` block is used to create a channel in a chosen category.

You can choose different types of channels including:

> `Text`
>
> `Voice`
>
> `Category`
>
> `Announcement`
>
> `Stage`
>
> `Forum`
>
> `Media`

### Options

> `Name` The name for the channel.
>
> `Topic` The topic for the channel.
>
> `Category` The category that the channel will be created in.
>
> `Position` The position for the channel.
>
> `Slowmode` The cooldown in seconds users must wait between sending messages, from `0` to `21600` (6 hours). `0` turns slowmode off. Leave it empty for no slowmode. Placeholders are supported. Not available for announcement and category channels.

### Permission Overwrites

Here you can choose which roles and users will have access to the channel. Leave this blank to allow all roles and users to access the channel.

<NodeInfoExplorer type="action_channel_create" />
