---
sidebar_position: 41.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create soundboard sound

<EmbedFlowNode type="action_soundboard_sound_create" />

The `Create soundboard sound` block is used to add a sound to a server's soundboard. The app needs the `Create Expressions` permission in the server.

### Options

> `Sound Name` The name of the sound, 2 to 32 characters.
>
> `Sound File` The URL of an MP3 or OGG file. It can be at most 512 KB and 5.2 seconds long.
>
> `Volume` The volume of the sound from 0 to 1. Defaults to 1.
>
> `Emoji` An optional emoji shown next to the sound. Custom emojis must be from the same server.

### Using an attachment argument

To let users upload the sound themselves, add a command argument of type `Attachment`, for example named `sound`, and set `Sound File` to `{{arg('sound')}}`. The file is checked when the block runs, so anything that isn't an MP3 or OGG file fails with an error you can catch with an error handler.

The created sound is stored as the block's result, so later blocks can use its ID, for example to delete it again with the [Delete soundboard sound](./action_soundboard_sound_delete.md) block.

<NodeInfoExplorer type="action_soundboard_sound_create" />
