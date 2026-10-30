---
sidebar_position: 42
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Play audio

<EmbedFlowNode type="action_voice_audio_play" />

The `Play audio` block makes your app join a voice channel and play an audio file in it. If the app is already in a voice channel in that server, it moves to the new one. If it already plays something in the server, that stops and the new audio plays instead.

The audio file is downloaded from the **Audio URL**, like an MP3, WAV, OGG or M4A file. To play a file someone uploads with a command, add an attachment [command argument](../options/option_command_argument.md) and use its URL, like `{{arg('sound').url}}`. Files can be up to 25 MB, and at most the first 15 minutes play.

The block finishes as soon as the audio starts playing, so the blocks after it run while it plays. The **Volume** is in percent, from 1 to 200, and defaults to the volume of the file. With **Leave When Done**, the app leaves the voice channel once the audio ended, otherwise it stays until a [Leave voice channel](./action_voice_channel_leave.md) block runs. Use [Stop audio](./action_voice_audio_stop.md) to stop the audio early.

Playing audio doesn't work in stage channels. Your app needs the Connect and Speak permissions in the voice channel. The audio stops when Kite restarts.

To protect your app's connection to Discord, this block shares a rate limit with the other blocks that change the status or voice state. Your app can run 5 of them at once, then one every 10 seconds. Runs over the limit fail.

<NodeInfoExplorer type="action_voice_audio_play" />
