// Tells search engines that support IndexNow, like Bing, which pages of the
// web app a deploy changed. web.sh runs `snapshot` before and after the deploy
// and then `submit` with both snapshots.
//
//   node indexnow.mjs snapshot <sitemap.xml> <out.json>
//   node indexnow.mjs submit <before.json> <after.json>
import { createHash } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";

const [command, ...args] = process.argv.slice(2);
const domain = process.env.WEB_DOMAIN;
const key = process.env.INDEXNOW_KEY;

if (command === "snapshot") {
  const [sitemapPath, outPath] = args;
  // The marketplace pages change with their install and star counts, not with
  // deploys.
  const urls = [
    ...readFileSync(sitemapPath, "utf8").matchAll(/<loc>([^<]+)<\/loc>/g),
  ]
    .map((m) => m[1])
    .filter((url) => !new URL(url).pathname.startsWith("/marketplace"));
  const hashes = {};
  for (const url of urls) {
    hashes[url] = await textHash(url);
  }
  writeFileSync(outPath, JSON.stringify(hashes));
} else if (command === "submit") {
  const [before, after] = args.map((p) => JSON.parse(readFileSync(p, "utf8")));
  // A page that failed to load after the deploy isn't sent, one that failed
  // before is, since it may be new.
  const changed = Object.keys(after).filter(
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
      host: domain,
      key,
      keyLocation: `https://${domain}/${key}.txt`,
      urlList: changed,
    }),
  });
  console.log(`IndexNow: sent ${changed.length} changed pages, ${res.status}`);
  for (const url of changed) {
    console.log(`  ${url}`);
  }
} else {
  console.error("Usage: indexnow.mjs snapshot|submit ...");
  process.exit(1);
}

// textHash hashes the text of a page without its scripts and markup, so a new
// build alone doesn't count as a change. It's null if the page didn't load.
async function textHash(url) {
  try {
    const res = await fetch(url, { signal: AbortSignal.timeout(15000) });
    if (!res.ok) {
      return null;
    }
    const text = (await res.text())
      .replace(/<(script|style)[^>]*>[\s\S]*?<\/\1>/g, " ")
      .replace(/<[^>]+>/g, " ")
      .replace(/\s+/g, " ")
      .trim();
    return createHash("sha256").update(text).digest("hex");
  } catch {
    return null;
  }
}
