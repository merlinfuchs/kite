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

export type EmbedButton = { label: string; url: string };

// Discord doesn't render a masked link whose text has characters it thinks
// could hide where it goes, or looks like a URL. Which ones isn't documented,
// so only text of letters, numbers and plain punctuation is linked. The Hangul
// fillers are letters that show as nothing.
function linkableText(text: string) {
  return (
    /^[\p{L}\p{M}\p{N} .,:'!?&+\-]+$/u.test(text) &&
    !/[\u115f\u1160\u3164\uffa0]/.test(text) &&
    !/\.\S/.test(text)
  );
}

function escapeMarkdown(text: string) {
  return text.replace(/[\\*_~`|<>[\]()#]/g, "\\$&");
}

export function buildKiteComponentEmbed({
  title,
  description,
  url,
  imageUrl,
  buttons,
}: {
  title: string;
  description: string;
  url: string;
  imageUrl: string;
  // The buttons of the page, like for adding a template, instead of the ones
  // for getting started.
  buttons?: EmbedButton[];
}) {
  const site = env.NEXT_PUBLIC_SITE_URL.replace(/\/$/, "");
  const linked = linkableText(title);
  const headline = linked ? `[${title}](${url})` : escapeMarkdown(title);

  return {
    component: {
      type: 17,
      accent_color: KITE_ACCENT_COLOR,
      components: [
        { type: 10, content: `-# ${KITE_EMOJI}  Kite` },
        {
          type: 10,
          content: `## ${headline}\n${escapeMarkdown(description)}`,
        },
        {
          type: 12,
          items: [
            {
              media: { url: imageUrl },
              description: "Make Discord bots with Kite",
            },
          ],
        },
        { type: 14, divider: false, spacing: 1 },
        {
          type: 1,
          // A linked headline already links to the page itself
          components: (
            buttons ?? [
              { label: "Get started", url: `${site}/apps` },
              {
                label: "How to make a bot",
                url: `${site}/guides/how-to-make-a-discord-bot`,
              },
              { label: "Docs", url: env.NEXT_PUBLIC_DOCS_LINK },
            ]
          )
            .filter((b) => !linked || b.url !== url)
            .map((b) => linkButton(b.label, b.url)),
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
