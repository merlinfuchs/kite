Kite.handle = function (event) {
  if (event.type != "DISCORD_MESSAGE_CREATE") return { success: true };

  const data = event.data;

  if (data.content == "!ping") {
    const resp = Kite.call({
      type: "DISCORD_MESSAGE_CREATE",
      data: {
        channel_id: data.channel_id,
        content: "Pong!",
      },
    });

    console.log(JSON.stringify(resp));
  }

  return { success: true };
};
