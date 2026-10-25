import { Template, templateNoun } from "@/lib/flow/templates";
import { ImageResponse } from "next/og";
import { readFile } from "node:fs/promises";
import path from "node:path";
import { renderToStaticMarkup } from "react-dom/server";

function svgDataUrl(svg: string) {
  return `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
}

// The files are included in the standalone build by outputFileTracingIncludes
// in next.config.mjs. Satori caches the parsed fonts by the array, so it's
// the same one for every image.
const file = (name: string) => readFile(path.join(process.cwd(), name));
const fonts = Promise.all([
  file("src/lib/og/Inter-Regular.ttf"),
  file("src/lib/og/Inter-Bold.ttf"),
]).then(([regular, bold]) => [
  {
    name: "Inter",
    data: regular,
    weight: 400 as const,
    style: "normal" as const,
  },
  { name: "Inter", data: bold, weight: 700 as const, style: "normal" as const },
]);
const logo = file("public/logo.svg").then((svg) => svgDataUrl(svg.toString()));

const orange = "#f97316";

// The avatar as a data URL, or null if Discord's CDN doesn't answer quickly.
// Satori would fetch it itself, but fail the whole image when that fails.
async function fetchAvatar(url: string) {
  try {
    const res = await fetch(`${url}?size=64`, {
      signal: AbortSignal.timeout(3000),
    });
    if (!res.ok) return null;
    const data = Buffer.from(await res.arrayBuffer()).toString("base64");
    return `data:image/png;base64,${data}`;
  } catch {
    return null;
  }
}

function plural(count: number, word: string) {
  return `${count} ${word}${count === 1 ? "" : "s"}`;
}

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
    author.kind === "user" ? await fetchAvatar(author.avatarUrl) : null;

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          padding: "64px 80px",
          backgroundColor: "#0c0a09",
          backgroundImage:
            "radial-gradient(circle at 85% 40%, rgba(249, 115, 22, 0.28), rgba(12, 10, 9, 0) 55%)",
          color: "#fafaf9",
          fontFamily: "Inter",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <img src={await logo} width={48} height={48} alt="" />
          <div style={{ fontSize: 36, fontWeight: 700 }}>Kite</div>
          <div
            style={{
              marginLeft: "auto",
              fontSize: 26,
              color: "#a8a29e",
              textTransform: "capitalize",
            }}
          >
            {templateNoun(template)}
          </div>
        </div>

        <div
          style={{
            display: "flex",
            flexDirection: "column",
            marginTop: "auto",
            marginBottom: "auto",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 32 }}>
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
          </div>
          <div
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
          </div>
        </div>

        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            fontSize: 26,
            color: "#a8a29e",
          }}
        >
          {contents.map((c) => (
            <div
              key={c}
              style={{
                display: "flex",
                padding: "8px 18px",
                borderRadius: 999,
                border: "2px solid #292524",
              }}
            >
              {c}
            </div>
          ))}
          {author.kind === "user" && (
            <div
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
          )}
          {author.kind === "kite" && (
            <div style={{ display: "flex", marginLeft: "auto" }}>
              Made by Kite
            </div>
          )}
        </div>
      </div>
    ),
    {
      width: 1200,
      height: 630,
      fonts: await fonts,
    }
  );
}
