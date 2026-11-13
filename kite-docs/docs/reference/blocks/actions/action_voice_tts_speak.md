---
sidebar_position: 43
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Speak text

<EmbedFlowNode type="action_voice_tts_speak" />

The `Speak text` block makes your app join a voice channel and read a text out loud. If the app is already in a voice channel in that server, it moves to the new one. If it already plays something in the server, that stops and the text is read instead. This block needs Kite Plus or a higher plan.

Set the **Text** to what the app says, up to 500 characters. It can use placeholders, like `{{arg('text')}}` to read what a user typed into a command. The **Language** is a code like `en`, `en-GB`, `de`, `fr`, `es`, `it`, `pt-BR`, `ja` or `ko`, and defaults to English.

The speech is made with Google Translate's text to speech, which needs no API key but isn't an official service. It can be slow or refuse requests when it's used a lot. Kite converts the speech like an uploaded sound, so it can be at most 2 minutes long.

The block finishes as soon as the speech starts playing, so the blocks after it run while it plays. With **Leave When Done**, the app leaves the voice channel once it finished, otherwise it stays until a [Leave voice channel](./action_voice_channel_leave.md) block runs. Use [Stop audio](./action_voice_audio_stop.md) to stop it early.

Speaking doesn't work in stage channels. Your app needs the Connect and Speak permissions in the voice channel. The speech stops when Kite restarts.

To protect your app's connection to Discord, this block shares a rate limit with the other blocks that change the status or voice state. Your app can run 5 of them at once, then one every 10 seconds. Runs over the limit fail.

<NodeInfoExplorer type="action_voice_tts_speak" />
