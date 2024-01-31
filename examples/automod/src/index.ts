import {
  RegExpMatcher,
  englishDataset,
  englishRecommendedTransformers,
} from "obscenity";
import { call, event } from "@merlingg/kite-sdk";

const matcher = new RegExpMatcher({
  ...englishDataset.build(),
  ...englishRecommendedTransformers,
});

event.on("DISCORD_MESSAGE_CREATE", (msg) => {
  if (matcher.hasMatch(msg.content)) {
    call("DISCORD_MESSAGE_DELETE", {
      channel_id: msg.channel_id,
      message_id: msg.id,
    });

    call("DISCORD_MESSAGE_CREATE", {
      channel_id: msg.channel_id,
      content: `Hey, you can't say that!`,
    });
  }
});
