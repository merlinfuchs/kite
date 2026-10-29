// Tells search engines that support IndexNow, like Bing, which pages of the
// web app a deploy changed. web.sh runs it before the deploy, which prints a
// snapshot of the pages, and after it with `submit` and that snapshot on stdin.
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";

const sitemap = readFileSync(
  new URL("../kite-web/public/sitemap-pages.xml", import.meta.url),
  "utf8"
);
// The marketplace pages change with their install and star counts, not with
// deploys.
const urls = [...sitemap.matchAll(/<loc>([^<]+)<\/loc>/g)]
  .map((m) => m[1])
  .filter((url) => !new URL(url).pathname.startsWith("/marketplace"));

if (process.argv[2] !== "submit") {
  console.log(JSON.stringify(await snapshot()));
  process.exit(0);
}

const before = JSON.parse(readFileSync(0, "utf8"));
const after = await snapshot();
// A page that failed to load after the deploy isn't sent, one that failed
// before is, since it may be new.
const changed = urls.filter(
  (url) => after[url] !== null && after[url] !== before[url]
);
if (changed.length === 0) {
  console.log("IndexNow: no pages changed");
  process.exit(0);
}
const res = await fetch("https://api.indexnow.org/indexnow", {
  method: "POST",
  headers: { "Content-Type": "application/json; charset=utf-8" },
  body: JSON.stringify({
    host: process.env.WEB_DOMAIN,
    key: process.env.INDEXNOW_KEY,
    urlList: changed,
  }),
});
console.log(`IndexNow: sent ${changed.length} changed pages, ${res.status}`);
for (const url of changed) {
  console.log(`  ${url}`);
}

// snapshot hashes what search engines see of each page, a few at a time.
async function snapshot() {
  const hashes = {};
  const queue = [...urls];
  await Promise.all(
    Array.from({ length: 8 }, async () => {
      for (let url = queue.shift(); url; url = queue.shift()) {
        hashes[url] = await pageHash(url);
      }
    })
  );
  return hashes;
}

// pageHash hashes the text, title, description and structured data of a page,
// without other scripts and markup, so a new build alone doesn't count as a
// change. Live numbers are marked with data-indexnow-ignore. It's null if the
// page didn't load.
async function pageHash(url) {
  let html;
  try {
    const res = await fetch(url, { signal: AbortSignal.timeout(15000) });
    if (!res.ok) {
      return null;
    }
    html = await res.text();
  } catch {
    return null;
  }
  const meta = [
    ...html.matchAll(
      /<meta[^>]+name="description"[^>]*>|<link[^>]+rel="canonical"[^>]*>|<script[^>]+application\/ld\+json[^>]*>[\s\S]*?<\/script>/g
    ),
  ].join("\n");
  const text = html
    .replace(/<(\w+)[^>]*data-indexnow-ignore[^>]*>[^<]*<\/\1>/g, " ")
    .replace(/<(script|style)[^>]*>[\s\S]*?<\/\1>/g, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
  return createHash("sha256").update(meta).update(text).digest("hex");
}
