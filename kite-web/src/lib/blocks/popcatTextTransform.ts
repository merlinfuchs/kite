import { z } from "zod";
import { popcatFields, popcatOperation as op, popcatQuery } from "./popcat";
import { BlockDefinition } from "./types";

const text = popcatQuery("popcat_text", "text");

export const popcatTextTransform: BlockDefinition = {
  type: "action_popcat_text_transform",
  title: "Transform text with Popcat",
  description: "Turn text into mocking text, Morse code, binary and more",
  icon: "type",
  category: "Popcat",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request_choice",
    integration: "popcat",
    setting: "popcat_endpoint",
    setting_label: "Transform",
    response: "json",
    result_key: "message",
    operations: [
      op(
        "/v2/mock",
        "Mocking text",
        "Alternates upper and lower case, like mOcKiNg TeXt. Returns text.",
        text
      ),
      op("/v2/reverse", "Reverse", "Reverses the text. Returns text.", text),
      op(
        "/v2/doublestruck",
        "Double-struck",
        "Writes the text in double-struck letters, like 𝕥𝕙𝕚𝕤. Returns text.",
        text
      ),
      op(
        "/v2/lulcat",
        "LOLcat",
        "Translates the text into LOLcat speak. Returns text.",
        text
      ),
      op(
        "/v2/texttomorse",
        "Text to Morse code",
        "Translates the text into Morse code. Returns morse.",
        text
      ),
      op(
        "/v2/encode",
        "Text to binary",
        "Encodes the text as binary. Returns text.",
        text
      ),
      op(
        "/v2/decode",
        "Binary to text",
        "Decodes binary back into text. Returns text.",
        popcatQuery(
          "popcat_text",
          "binary",
          "Binary",
          "The binary to decode, like 01101000 01101001."
        )
      ),
    ],
  },
  fields: popcatFields("popcat_text"),
  result: {
    schema: z
      .record(z.unknown())
      .describe(
        "What Popcat returned, with the fields popcat_endpoint lists, like result('id').text"
      ),
  },
};
