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

export interface ComponentEmbedOptions {
  title?: string;
  description?: string;
}

function linkButton(label: string, url: string) {
  return { type: 2, style: 5, label, url };
}

export function buildKiteComponentEmbed(opts: ComponentEmbedOptions = {}) {
  const site = env.NEXT_PUBLIC_SITE_URL.replace(/\/$/, "");

  const title = opts.title ?? "Kite";
  const description =
    opts.description ??
    "Kite is a free Discord bot maker. Build your own bot with slash commands, buttons, events and AI in a visual editor, without writing code. We host it for you 24/7.";

  const buttons = [
    linkButton("Open Kite", site),
    linkButton("Read Docs", env.NEXT_PUBLIC_DOCS_LINK),
  ];
  // Only show the support button when a real invite is configured
  if (
    env.NEXT_PUBLIC_DISCORD_LINK &&
    env.NEXT_PUBLIC_DISCORD_LINK.replace(/\/$/, "") !== "https://discord.gg"
  ) {
    buttons.push(
      linkButton("Join Support Server", env.NEXT_PUBLIC_DISCORD_LINK)
    );
  }

  return {
    component: {
      type: 17,
      accent_color: KITE_ACCENT_COLOR,
      components: [
        {
          type: 9,
          components: [
            { type: 10, content: `## ${title}` },
            { type: 10, content: description },
          ],
          accessory: {
            type: 11,
            media: { url: `${site}/logo.png` },
            description: "Kite logo",
          },
        },
        { type: 1, components: buttons },
      ],
    },
  };
}

// Serialized for an inline <script type="application/json">. "<" is escaped
// so nothing in the payload can close the script tag early.
export function serializeComponentEmbed(payload: object) {
  return JSON.stringify(payload).replace(/</g, "\\u003c");
}
