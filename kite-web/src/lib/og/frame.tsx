import { ImageResponse } from "next/og";
import { readFile } from "node:fs/promises";
import path from "node:path";
import { ReactNode } from "react";

export function svgDataUrl(svg: string) {
  return `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
}

// The files are included in the standalone build by outputFileTracingIncludes
// in next.config.mjs. Their paths are written out, since the build would copy
// everything a computed path could point to. Satori caches the parsed fonts by
// the array, so it's the same one for every image.
const fonts = Promise.all([
  readFile(path.join(process.cwd(), "src/lib/og/Inter-Regular.ttf")),
  readFile(path.join(process.cwd(), "src/lib/og/Inter-Bold.ttf")),
]).then(([regular, bold]) => [
  {
    name: "Inter",
    data: regular,
    weight: 400 as const,
    style: "normal" as const,
  },
  { name: "Inter", data: bold, weight: 700 as const, style: "normal" as const },
]);
const logo = readFile(path.join(process.cwd(), "public/logo.svg"), "utf8").then(
  svgDataUrl
);

export const orange = "#f97316";
export const muted = "#a8a29e";

// The avatar as a data URL, or null if Discord's CDN doesn't answer quickly.
// Satori would fetch it itself, but fail the whole image when that fails.
export async function fetchAvatar(url: string, size: number) {
  try {
    const res = await fetch(`${url}?size=${size}`, {
      signal: AbortSignal.timeout(3000),
    });
    if (!res.ok) return null;
    const data = Buffer.from(await res.arrayBuffer()).toString("base64");
    return `data:image/png;base64,${data}`;
  } catch {
    return null;
  }
}

export function plural(count: number, word: string) {
  return `${count} ${word}${count === 1 ? "" : "s"}`;
}

export function Pill({ children }: { children: ReactNode }) {
  return (
    <div
      style={{
        display: "flex",
        padding: "8px 18px",
        borderRadius: 999,
        border: "2px solid #292524",
      }}
    >
      {children}
    </div>
  );
}

// renderOgImage draws a link preview like og.png, with the Kite logo and the
// label of what it shows at the top, the body in the middle and the footer at
// the bottom. They're lists, since Satori lays out a fragment as one element.
export async function renderOgImage({
  label,
  body,
  footer,
}: {
  label: string;
  body: ReactNode[];
  footer: ReactNode[];
}) {
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
              color: muted,
              textTransform: "capitalize",
            }}
          >
            {label}
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
          {body}
        </div>

        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            fontSize: 26,
            color: muted,
          }}
        >
          {footer}
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
