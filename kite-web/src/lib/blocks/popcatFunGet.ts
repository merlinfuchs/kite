import { z } from "zod";
import { popcatBlock, popcatOperation as op } from "./popcat";

export const popcatFunGet = popcatBlock({
  type: "action_popcat_fun_get",
  id_prefix: "popcat",
  title: "Get Popcat fun",
  description: "Get a random joke, fact, 8-ball answer and more from Popcat",
  icon: "party-popper",
  setting_label: "Get",
  operations: [
    op(
      "/v2/8ball",
      "Magic 8-ball",
      "A random magic 8-ball answer. Returns answer."
    ),
    op("/v2/joke", "Joke", "A random joke. Returns joke."),
    op("/v2/fact", "Fact", "A random fact. Returns fact."),
    op(
      "/v2/showerthoughts",
      "Shower thought",
      "A random shower thought from Reddit. Returns result, author and upvotes."
    ),
    op(
      "/v2/wyr",
      "Would you rather",
      "A random would you rather question. Returns its two options as ops1 and ops2."
    ),
    op(
      "/v2/meme",
      "Meme",
      "A random meme from Reddit. Returns title, link, subreddit, content.image, misc.upvotes and misc.nsfw, which is true for memes that aren't safe for work."
    ),
    op("/v2/car", "Car", "A random picture of a car. Returns title and image."),
    op(
      "/v2/randomcolor",
      "Random color",
      "A random color. Returns hex, name and image."
    ),
  ],
  result: {
    schema: z
      .record(z.unknown())
      .describe(
        "What Popcat returned, with the fields popcat_endpoint lists, like result('id').answer for the magic 8-ball"
      ),
  },
});
