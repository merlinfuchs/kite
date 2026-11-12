---
sidebar_position: 43.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Play soundboard sound

<EmbedFlowNode type="action_soundboard_sound_play" />

The `Play soundboard sound` block plays a soundboard sound in a voice channel, the same way members play one from the soundboard. Discord plays it for everyone in the channel, so it needs no uploaded audio and no Premium, unlike [Play audio](./action_voice_audio_play.md).

Your app has to be in the voice channel already, so use [Join voice channel](./action_voice_channel_join.md) first. It can't play sounds while it's muted or deafened, so leave **Mute Self** and **Deafen Self** off when joining. Leave the **Voice Channel** empty to use the channel the flow runs in, like a command used in a voice channel's chat.

The **Sound** is the ID of a soundboard sound of the server, or of one of Discord's default sounds. To play a sound of another server your app is in, also set **Sound From Server** to that server's ID. Your app then also needs the Use External Sounds permission. To find the IDs of a server's sounds, send a `GET /guilds/{guild_id}/soundboard-sounds` request with the [Discord API Request](./action_discord_api_request.md) block, and `GET /soundboard-default-sounds` for the default sounds.

Your app needs the Speak and Use Soundboard permissions in the voice channel.

<NodeInfoExplorer type="action_soundboard_sound_play" />
