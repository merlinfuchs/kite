import {
  RegExpMatcher,
  englishDataset,
  englishRecommendedTransformers,
} from "obscenity";
import { on, call } from "@merlingg/kite-sdk";

const matcher = new RegExpMatcher({
  ...englishDataset.build(),
  ...englishRecommendedTransformers,
});

on("DISCORD_MESSAGE_CREATE", (msg) => {
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
