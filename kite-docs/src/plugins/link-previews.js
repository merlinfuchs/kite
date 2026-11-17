const fs = require("fs/promises");
const path = require("path");
const { Resvg } = require("@resvg/resvg-js");

// The fonts and logo of the link previews of kite.onl, see
// kite-web/src/lib/og/frame.tsx, which this draws them like.
const webOg = path.join(__dirname, "../../../kite-web/src/lib/og");
const logoPath = path.join(__dirname, "../../static/img/logo.svg");

const orange = "#f97316";
const muted = "#a8a29e";
const discordInvite = "https://discord.gg/rNd9jWHnXh";
const kiteEmoji = "<:kite:1266749020676821053>";

function labelOf(pathname) {
  if (pathname.startsWith("/reference/blocks/")) return "block";
  const labels = {
    reference: "reference",
    guides: "guide",
    examples: "example",
    blog: "blog",
  };
  return labels[pathname.split("/")[1]] ?? "docs";
}

// Satori takes elements as plain objects, which saves a JSX build step.
function h(type, style, ...children) {
  return { type, props: { style, children } };
}

function decode(text) {
  return text
    .replace(/&quot;/g, '"')
    .replace(/&#x27;|&#39;/g, "'")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&amp;/g, "&");
}

function metaContent(html, property) {
  const match = html.match(
    new RegExp(`<meta[^>]*property="${property}"[^>]*content="([^"]*)"`)
  );
  return match && decode(match[1]);
}

async function htmlFiles(dir) {
  const entries = await fs.readdir(dir, {
    withFileTypes: true,
    recursive: true,
  });
  return entries
    .filter((e) => e.isFile() && e.name.endsWith(".html"))
    .map((e) => path.join(e.parentPath, e.name));
}

function image({ logo, label, title, description, url }) {
  return h(
    "div",
    {
      width: 1200,
      height: 630,
      boxSizing: "border-box",
      display: "flex",
      flexDirection: "column",
      padding: "64px 80px",
      backgroundColor: "#0c0a09",
      backgroundImage:
        "radial-gradient(circle at 85% 40%, rgba(249, 115, 22, 0.28), rgba(12, 10, 9, 0) 55%)",
      color: "#fafaf9",
      fontFamily: "Inter",
    },
    h(
      "div",
      { display: "flex", alignItems: "center", gap: 16 },
      { type: "img", props: { src: logo, width: 48, height: 48 } },
      h("div", { fontSize: 36, fontWeight: 700 }, "Kite"),
      h(
        "div",
        {
          marginLeft: "auto",
          fontSize: 26,
          color: muted,
          textTransform: "capitalize",
        },
        label
      )
    ),
    h(
      "div",
      {
        display: "flex",
        flexDirection: "column",
        marginTop: "auto",
        marginBottom: "auto",
      },
      h(
        "div",
        {
          fontSize: title.length > 28 ? 64 : 80,
          fontWeight: 700,
          lineHeight: 1.1,
          lineClamp: 2,
          display: "block",
        },
        title
      ),
      description &&
        h(
          "div",
          {
            marginTop: 32,
            fontSize: 34,
            lineHeight: 1.4,
            color: "#d6d3d1",
            lineClamp: 2,
            display: "block",
          },
          description
        )
    ),
    h(
      "div",
      { display: "flex", fontSize: 26, color: muted },
      h("div", { color: orange }, url.host),
      url.pathname === "/" ? "" : url.pathname
    )
  );
}

// The Discord component embed of a page, like the ones of kite.onl, see
// kite-web/src/lib/discordComponentEmbed.ts for the rules Discord has for it.
function componentEmbed({ title, description, url, imageUrl, appUrl }) {
  const buttons = [
    { label: "Read page", url },
    { label: "Get started", url: `${appUrl}/apps` },
    { label: "Support server", url: discordInvite },
  ];
  return {
    component: {
      type: 17,
      accent_color: 0xf97316,
      components: [
        { type: 10, content: `-# ${kiteEmoji}  Kite docs` },
        {
          type: 10,
          content: `## [${title}](${url})${
            description ? `\n${description}` : ""
          }`,
        },
        { type: 12, items: [{ media: { url: imageUrl }, description: title }] },
        { type: 14, divider: false, spacing: 1 },
        {
          type: 1,
          components: buttons.map((b) => ({
            type: 2,
            style: 5,
            label: b.label,
            url: b.url,
          })),
        },
      ],
    },
  };
}

// Draws a link preview for every page at build time, in the style of the ones
// of kite.onl, and links it instead of the default image of the site. Discord
// shows the component embed instead, with the same image.
module.exports = function linkPreviewsPlugin() {
  return {
    name: "link-previews",
    async postBuild({ outDir, siteConfig }) {
      const { default: satori } = await import("satori");
      const fonts = [
        {
          name: "Inter",
          data: await fs.readFile(path.join(webOg, "Inter-Regular.ttf")),
          weight: 400,
          style: "normal",
        },
        {
          name: "Inter",
          data: await fs.readFile(path.join(webOg, "Inter-Bold.ttf")),
          weight: 700,
          style: "normal",
        },
      ];
      const logo = `data:image/svg+xml;base64,${(
        await fs.readFile(logoPath)
      ).toString("base64")}`;
      const defaultImage = new URL(
        siteConfig.themeConfig.image,
        siteConfig.url + "/"
      ).href;

      for (const file of await htmlFiles(outDir)) {
        let html = await fs.readFile(file, "utf8");
        const title = metaContent(html, "og:title");
        const pageUrl = metaContent(html, "og:url");
        // Pages with an image of their own keep it.
        if (!title || !pageUrl || !html.includes(`content="${defaultImage}"`)) {
          continue;
        }
        const url = new URL(pageUrl);
        // 404.html, which isn't shared.
        if (url.pathname.endsWith(".html")) continue;

        const pageTitle = title.replace(/ \| Kite$/, "");
        const description = metaContent(html, "og:description");
        const svg = await satori(
          image({
            logo,
            label: labelOf(url.pathname),
            title: pageTitle,
            description,
            url,
          }),
          { width: 1200, height: 630, fonts }
        );
        const png = new Resvg(svg).render().asPng();

        const name = url.pathname === "/" ? "index" : url.pathname.slice(1);
        const imagePath = `/og/${name}.png`;
        await fs.mkdir(path.dirname(path.join(outDir, imagePath)), {
          recursive: true,
        });
        await fs.writeFile(path.join(outDir, imagePath), png);

        const imageUrl = siteConfig.url + imagePath;
        const embed = componentEmbed({
          title: pageTitle,
          description,
          url: url.href,
          imageUrl,
          appUrl: siteConfig.customFields.appBaseUrl,
        });
        // "<" is escaped so nothing in it can close the script tag early.
        const script = `<script id="discord:component-embed" type="application/json">${JSON.stringify(
          embed
        ).replace(/</g, "\\u003c")}</script>`;
        html = html
          .replaceAll(`content="${defaultImage}"`, `content="${imageUrl}"`)
          .replace("</head>", `${script}</head>`);
        await fs.writeFile(file, html);
      }
    },
  };
};
