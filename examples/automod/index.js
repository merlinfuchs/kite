const forbiddenPhrases = ["Kite is bad", "Kite is a bad bot"];

Kite.handle = function (event) {
  if (event.type != "DISCORD_MESSAGE_CREATE") return { success: true };

  const data = event.data;

  for (const phrase of forbiddenPhrases) {
    if (data.content.includes(phrase)) {
      Kite.call({
        type: "DISCORD_MESSAGE_DELETE",
        data: {
          channel_id: data.channel_id,
          message_id: data.id,
        },
      });

      Kite.call({
        type: "DISCORD_MESSAGE_CREATE",
        data: {
          channel_id: data.channel_id,
          content: `Hey, you can't say that!`,
        },
      });

      break;
    }
  }

  return { success: true };
};
