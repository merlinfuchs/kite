import { Template } from "@/lib/flow/templates";
import { renderToStaticMarkup } from "react-dom/server";
import {
  fetchAvatar,
  orange,
  Pill,
  plural,
  renderOgImage,
  svgDataUrl,
} from "./frame";

// The link preview of a template or bundle, like og.png with its icon and name.
export async function renderTemplateImage(template: Template) {
  // Satori only draws plain elements, so the Lucide icon is drawn as an image.
  const icon = svgDataUrl(
    renderToStaticMarkup(
      <template.icon color={orange} size={72} strokeWidth={2} />
    )
  );

  const contents = [
    template.commands.length && plural(template.commands.length, "command"),
    template.eventListeners.length &&
      plural(template.eventListeners.length, "event listener"),
    template.messages.length && plural(template.messages.length, "message"),
  ].filter(Boolean);

  const author = template.author;
  const avatar =
    author.kind === "user" ? await fetchAvatar(author.avatarUrl, 64) : null;

  return renderOgImage({
    label: "template",
    body: [
      <div
        key="name"
        style={{ display: "flex", alignItems: "center", gap: 32 }}
      >
        <div
          style={{
            width: 128,
            height: 128,
            flexShrink: 0,
            borderRadius: 28,
            backgroundColor: "rgba(249, 115, 22, 0.15)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <img src={icon} width={72} height={72} alt="" />
        </div>
        <div
          style={{
            // Long names take two lines, which fit at a smaller size.
            fontSize: template.name.length > 24 ? 60 : 76,
            fontWeight: 700,
            lineHeight: 1.1,
            lineClamp: 2,
            display: "block",
          }}
        >
          {template.name}
        </div>
      </div>,
      <div
        key="description"
        style={{
          marginTop: 36,
          fontSize: 34,
          lineHeight: 1.4,
          color: "#d6d3d1",
          lineClamp: 2,
          display: "block",
        }}
      >
        {template.description}
      </div>,
    ],
    footer: [
      ...contents.map((c) => <Pill key={c}>{c}</Pill>),
      author.kind === "user" && (
        <div
          key="author"
          style={{
            display: "flex",
            alignItems: "center",
            gap: 14,
            marginLeft: "auto",
          }}
        >
          <span>Made by</span>
          {avatar && (
            <img
              src={avatar}
              width={44}
              height={44}
              style={{ borderRadius: 999 }}
              alt=""
            />
          )}
          <span style={{ color: "#fafaf9" }}>{author.name}</span>
        </div>
      ),
      author.kind === "kite" && (
        <div key="author" style={{ display: "flex", marginLeft: "auto" }}>
          Made by Kite
        </div>
      ),
    ],
  });
}
