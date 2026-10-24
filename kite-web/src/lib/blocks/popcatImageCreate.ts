import { z } from "zod";
import { popcatBlock, popcatOperation, popcatQuery } from "./popcat";

const text = popcatQuery("popcat_text", "text");
const image = popcatQuery("popcat_image", "image");

// A meme with one text on it.
const textMeme = (path: string, label: string, description: string) =>
  popcatOperation(path, label, description, text);

// An image with an effect or overlay.
const imageEffect = (path: string, label: string, description: string) =>
  popcatOperation(path, label, description, image);

// Endpoints are picked rather than all exposed, as some put text on images
// of real politicians. See design/integrations.md.
const operations = [
  popcatOperation(
    "/v2/drake",
    "Drake meme",
    "Drake turning away from the first text and pointing at the second.",
    popcatQuery("popcat_text", "text1", "Disliked Text"),
    popcatQuery("popcat_text_2", "text2", "Liked Text")
  ),
  popcatOperation(
    "/v2/pooh",
    "Fancy Pooh meme",
    "Winnie the Pooh plain for the first text and fancy for the second.",
    popcatQuery("popcat_text", "text1", "Plain Text"),
    popcatQuery("popcat_text_2", "text2", "Fancy Text")
  ),
  popcatOperation(
    "/v2/happysad",
    "Happy and sad meme",
    "A happy and a sad panel, with a text each.",
    popcatQuery("popcat_text", "text1", "Happy Text"),
    popcatQuery("popcat_text_2", "text2", "Sad Text")
  ),
  textMeme("/v2/sadcat", "Sad cat", "A crying cat with the text."),
  textMeme("/v2/oogway", "Oogway quote", "Master Oogway saying the text."),
  textMeme(
    "/v2/pikachu",
    "Surprised Pikachu",
    "Surprised Pikachu below the text."
  ),
  textMeme("/v2/alert", "iPhone alert", "An iPhone alert with the text."),
  textMeme("/v2/caution", "Caution sign", "A caution sign with the text."),
  textMeme(
    "/v2/unforgivable",
    "Unforgivable",
    "The text as something unforgivable."
  ),
  textMeme("/v2/supreme", "Supreme logo", "The text as a Supreme logo."),
  textMeme(
    "/v2/couldread",
    "If those kids could read",
    "If those kids could read, they'd be very upset, about the text."
  ),
  textMeme("/v2/facts", "Facts book", "The text in the book of facts."),
  popcatOperation(
    "/v2/opinion",
    "Opinion",
    "Someone with an opinion nobody asked for.",
    image,
    popcatQuery("popcat_text", "text", "Opinion")
  ),
  popcatOperation(
    "/v2/quote",
    "Quote",
    "A quote of someone, with their picture.",
    image,
    popcatQuery(
      "popcat_text",
      "text",
      "Quote",
      "The quote, at most 125 characters."
    ),
    popcatQuery("popcat_text_2", "name", "Name", "Who said it.")
  ),
  popcatOperation(
    "/v2/discord-message",
    "Fake Discord message",
    "A picture of a Discord message.",
    popcatQuery("popcat_text", "content", "Content"),
    popcatQuery("popcat_text_2", "username", "Username"),
    popcatQuery("popcat_image", "avatar", "Avatar URL"),
    popcatQuery(
      "popcat_color",
      "color",
      "Username Color",
      "A hex color without #, like ff0000. Leave empty for white.",
      false
    )
  ),
  popcatOperation(
    "/v2/welcomecard",
    "Welcome card",
    "A welcome card with an avatar and three lines of text, like for new members.",
    popcatQuery(
      "popcat_image_2",
      "background",
      "Background URL",
      "URL of the background image, which has to be a PNG."
    ),
    popcatQuery("popcat_image", "avatar", "Avatar URL"),
    popcatQuery("popcat_text", "text1", "First Line"),
    popcatQuery("popcat_text_2", "text2", "Second Line"),
    popcatQuery("popcat_text_3", "text3", "Third Line")
  ),
  popcatOperation(
    "/v2/ship",
    "Ship",
    "Two avatars shipped together.",
    popcatQuery("popcat_image", "user1", "First Image URL"),
    popcatQuery("popcat_image_2", "user2", "Second Image URL")
  ),
  imageEffect("/v2/pet", "Pet", "A hand petting the image, as a GIF."),
  imageEffect("/v2/jail", "Jail", "The image behind bars."),
  imageEffect("/v2/wanted", "Wanted", "The image on a wanted poster."),
  imageEffect("/v2/drip", "Drip", "The image wearing a puffer jacket."),
  imageEffect("/v2/clown", "Clown", "The image as a clown."),
  imageEffect("/v2/uncover", "Uncover", "The image uncovered behind a wall."),
  imageEffect("/v2/ad", "Ad", "The image as an ad."),
  imageEffect(
    "/v2/jokeoverhead",
    "Joke over head",
    "A joke flying over the image's head."
  ),
  imageEffect("/v2/mnm", "M&M", "The image on an M&M."),
  imageEffect("/v2/blur", "Blur", "The image blurred."),
  imageEffect("/v2/invert", "Invert", "The image with its colors inverted."),
  imageEffect("/v2/greyscale", "Greyscale", "The image in greyscale."),
  popcatOperation(
    "/v2/colorify",
    "Colorify",
    "The image tinted in a color.",
    image,
    popcatQuery("popcat_color", "color")
  ),
];

export const popcatImageCreate = popcatBlock({
  type: "action_popcat_image_create",
  id_prefix: "popcat",
  title: "Create Popcat image",
  description: "Make a meme, welcome card or image effect with Popcat",
  icon: "image",
  setting_label: "Image",
  response: "url",
  operations,
  result: {
    schema: z
      .string()
      .describe(
        "URL of the image, which Discord loads when a message shows it, like as an embed's image or as a link. Popcat makes the image then, so a mistake in a setting only shows as a broken image."
      ),
  },
});
