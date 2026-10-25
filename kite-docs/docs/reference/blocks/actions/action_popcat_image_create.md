---
sidebar_position: 37.84
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create Popcat image

<EmbedFlowNode type="action_popcat_image_create" />

The `Create Popcat image` block makes an image, like a meme, a welcome card or an effect on someone's avatar. Enable Popcat under [Integrations](../../integrations.md) first. It needs no key, but it's off until you turn it on, as its blocks send what you enter to [Popcat](https://popcat.xyz/api), a free API that isn't run by Kite.

Pick the image and fill in its settings, which can be placeholders like `{{ user.avatar_url }}` or `{{ arg('text') }}`. The result is the image's URL, which you can show in a message, like as the image of an embed with `{{ result('image') }}`. Kite doesn't load the image itself: Popcat makes it when Discord loads it, so the block is instant, and a mistake in a setting only shows as a missing image.

| Image | Settings | What it makes |
| --- | --- | --- |
| Drake meme | Disliked Text, Liked Text | Drake turning away from the first text and pointing at the second. |
| Fancy Pooh meme | Plain Text, Fancy Text | Winnie the Pooh plain for the first text and fancy for the second. |
| Happy and sad meme | Happy Text, Sad Text | A happy and a sad panel, with a text each. |
| Sad cat | Text | A crying cat with the text. |
| Oogway quote | Text | Master Oogway saying the text. |
| Surprised Pikachu | Text | Surprised Pikachu below the text. |
| iPhone alert | Text | An iPhone alert with the text. |
| Caution sign | Text | A caution sign with the text. |
| Unforgivable | Text | The text as something unforgivable. |
| Supreme logo | Text | The text as a Supreme logo. |
| If those kids could read | Text | If those kids could read, they'd be very upset, about the text. |
| Facts book | Text | The text in the book of facts. |
| Opinion | Image URL, Opinion | Someone with an opinion nobody asked for. |
| Quote | Image URL, Quote, Name | A quote of someone, with their picture. |
| Fake Discord message | Content, Username, Avatar URL, Username Color (optional) | A picture of a Discord message. |
| Welcome card | Background URL, Avatar URL, First Line, Second Line, Third Line | A welcome card with an avatar and three lines of text, like for new members. |
| Ship | First Image URL, Second Image URL | Two avatars shipped together. |
| Pet | Image URL | A hand petting the image, as a GIF. |
| Jail | Image URL | The image behind bars. |
| Wanted | Image URL | The image on a wanted poster. |
| Drip | Image URL | The image wearing a puffer jacket. |
| Clown | Image URL | The image as a clown. |
| Uncover | Image URL | The image uncovered behind a wall. |
| Ad | Image URL | The image as an ad. |
| Joke over head | Image URL | A joke flying over the image's head. |
| M&M | Image URL | The image on an M&M. |
| Blur | Image URL | The image blurred. |
| Invert | Image URL | The image with its colors inverted. |
| Greyscale | Image URL | The image in greyscale. |
| Colorify | Image URL, Color | The image tinted in a color. |

<NodeInfoExplorer type="action_popcat_image_create" />
