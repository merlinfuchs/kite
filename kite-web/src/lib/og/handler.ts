import { Template } from "@/lib/flow/templates";
import type { NextApiRequest, NextApiResponse } from "next";
import { renderTemplateImage } from "./templateImage";

// Rendered images by URL, so crawlers asking again don't render them again.
// Unknown IDs are kept too, as null.
const cache = new Map<string, { image: Buffer | null; expires: number }>();
const cacheDuration = 15 * 60 * 1000;
const maxCached = 500;

// Responds with the link preview of the template, or the image of the site if
// there is none.
export async function handleTemplateImage(
  req: NextApiRequest,
  res: NextApiResponse,
  load: () => Promise<Template | undefined>
) {
  const key = req.url ?? "";
  let entry = cache.get(key);
  if (!entry || entry.expires < Date.now()) {
    const template = await load();
    const image = template
      ? Buffer.from(await (await renderTemplateImage(template)).arrayBuffer())
      : null;
    entry = { image, expires: Date.now() + cacheDuration };
    if (cache.size >= maxCached) {
      cache.delete(cache.keys().next().value!);
    }
    cache.set(key, entry);
  }

  if (!entry.image) {
    res.redirect(302, "/og.png");
    return;
  }
  res.setHeader("Content-Type", "image/png");
  res.send(entry.image);
}
