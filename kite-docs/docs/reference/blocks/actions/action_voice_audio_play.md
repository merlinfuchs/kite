---
sidebar_position: 42
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Play audio

<EmbedFlowNode type="action_voice_audio_play" />

The `Play audio` block makes your app join a voice channel and play a sound in it. If the app is already in a voice channel in that server, it moves to the new one. If it already plays something in the server, that stops and the new sound plays instead.

Upload the **Sound** in the block's settings, like an MP3, WAV, OGG or M4A file of up to 2 minutes. Kite converts it once when you upload it and evens out its volume, so all sounds play about equally loud. Uploaded sounds count toward your app's file storage.

The block finishes as soon as the sound starts playing, so the blocks after it run while it plays. With **Leave When Done**, the app leaves the voice channel once the sound ended, otherwise it stays until a [Leave voice channel](./action_voice_channel_leave.md) block runs. Use [Stop audio](./action_voice_audio_stop.md) to stop the sound early.

Playing audio doesn't work in stage channels. Your app needs the Connect and Speak permissions in the voice channel. The sound stops when Kite restarts.

To protect your app's connection to Discord, this block shares a rate limit with the other blocks that change the status or voice state. Your app can run 5 of them at once, then one every 10 seconds. Runs over the limit fail.

<NodeInfoExplorer type="action_voice_audio_play" />
