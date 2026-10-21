import env from "@/lib/env/client";

// Discord Component Embeds: when a Kite link is posted in Discord, the crawler
// reads this payload from the page <head> and renders it as a Components V2
// container instead of the plain Open Graph preview.
// Docs: https://docs.discord.com/developers/link-previews/component-embeds
//
// Rules (enforced by Discord, the whole payload is dropped if broken):
// - top level is exactly one Container (type 17)
// - max 40 components, max 3000 bytes serialized
// - buttons must be link buttons (style 5) with only
//   id/type/url/style/label/emoji/disabled keys
// - media URLs must be absolute http(s)

// Kite orange (#F97316, same as --primary in globals.css)
export const KITE_ACCENT_COLOR = 0xf97316;

const KITE_EMOJI = "<:kite:1266749020676821053>";

function linkButton(label: string, url: string) {
  return { type: 2, style: 5, label, url };
}

export function buildKiteComponentEmbed({
  title,
  description,
  url,
}: {
  title: string;
  description: string;
  url: string;
}) {
  const site = env.NEXT_PUBLIC_SITE_URL.replace(/\/$/, "");

  return {
    component: {
      type: 17,
      accent_color: KITE_ACCENT_COLOR,
      components: [
        { type: 10, content: `-# ${KITE_EMOJI}  Kite` },
        { type: 10, content: `## [${title}](${url})\n${description}` },
        {
          type: 12,
          items: [
            {
              media: { url: `${site}/og.png` },
              description: "Make Discord bots with Kite",
            },
          ],
        },
        { type: 14, divider: false, spacing: 1 },
        {
          type: 1,
          // The headline already links to the page itself
          components: [
            linkButton("Get started", `${site}/apps`),
            linkButton(
              "How to make a bot",
              `${site}/guides/how-to-make-a-discord-bot`
            ),
            linkButton("Docs", env.NEXT_PUBLIC_DOCS_LINK),
          ].filter((b) => b.url !== url),
        },
      ],
    },
  };
}

// Serialized for an inline <script type="application/json">. "<" is escaped
// so nothing in the payload can close the script tag early.
export function serializeComponentEmbed(payload: object) {
  return JSON.stringify(payload).replace(/</g, "\\u003c");
}
