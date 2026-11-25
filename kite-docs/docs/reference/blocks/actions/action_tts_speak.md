---
sidebar_position: 42.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Speak Text

<EmbedFlowNode type="action_tts_speak" />

The `Speak Text` block makes your app join a voice channel and speak text in it. Set the **Text** to what should be said, like `{{message.content}}`, pick a **Voice** and a **Speed**, and set the **Voice Channel**. If the app is already in a voice channel in that server, it moves to the new one. If it already plays something in the server, that stops and the speech plays instead. This block needs Kite Plus or a higher plan.

The block finishes as soon as the speech starts, so the blocks after it run while it plays. With **Leave When Done**, the app leaves the voice channel once it has spoken, otherwise it stays until a [Leave voice channel](./action_voice_channel_leave.md) block runs. Use [Stop audio](./action_voice_audio_stop.md) to stop the speech early.

The text can be up to 1,000 characters after placeholders are filled in. The block costs [2 credits](../../credit-system.md#cost-breakdown) per run. It fails when the text is empty or too long, so check long text from users first.

Speaking doesn't work in stage channels. Your app needs the Connect and Speak permissions in the voice channel. The speech stops when Kite restarts.

To protect your app's connection to Discord, this block shares a rate limit with the other blocks that change the status or voice state. Your app can run 5 of them at once, then one every 10 seconds. Runs over the limit fail.

Speaking text needs an OpenAI API key on the Kite service and ffmpeg, which converts the speech like it does uploaded sounds. On kite.onl both are set up for you. If you host Kite yourself, add your key as `openai.api_key` in the config, the same key the AI blocks use.

<NodeInfoExplorer type="action_tts_speak" />
