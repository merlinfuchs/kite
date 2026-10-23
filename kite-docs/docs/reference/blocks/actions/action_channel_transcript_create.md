---
sidebar_position: 26
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create channel transcript

<EmbedFlowNode type="action_channel_transcript_create" />

The `Create channel transcript` block saves the most recent messages of a channel or thread as an HTML file. It's made for ticket systems: create the transcript, send it to a staff log channel, then delete the ticket channel.

Download the file and open it in your browser to read it. It looks like the Discord client and includes authors, replies, embeds, attachments, reactions and Components V2 messages. Times are shown in the reader's own time zone.

## Options

- `Target Channel`: the channel or thread to make the transcript of.
- `Message Limit`: how many of the most recent messages to include, between 1 and 1000. Leave it empty for 1000.

## Sending the transcript

The block doesn't send the file itself. Its result is the file, and any block that sends a new message attaches it when you put the result in the message as its own placeholder:

```
Ticket closed by {{user.mention}}
{{result('transcript_block_id')}}
```

This works in `Create channel message`, `Send direct message` and `Create response message`. The placeholder is removed from the text and the file is attached instead. `{{result('transcript_block_id').name}}` gives the file name and `.size` its size in bytes, as text.

Blocks that edit a message can't attach files and fail if the message contains the transcript.

:::info
The bot needs the `View Channel` and `Read Message History` permissions in the channel it makes the transcript of, and `Attach Files` where the transcript is sent. Message text is only included if the app has the `Message Content` intent enabled in the Discord developer portal.
:::

:::caution
Links to images and files in the transcript point to Discord and stop working after about a day. Avatars and emojis keep working.
:::

<NodeInfoExplorer type="action_channel_transcript_create" />
