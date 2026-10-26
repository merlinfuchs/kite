// urlSitemap is a sitemap of paths of kite.onl, like the static ones in
// public/.
export function urlSitemap(paths: string[]) {
  const urls = paths.map(
    (path) => `  <url><loc>https://kite.onl${escapeXml(path)}</loc></url>`
  );
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    ...urls,
    "</urlset>",
    "",
  ].join("\n");
}

function escapeXml(text: string) {
  return text
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&apos;");
}
