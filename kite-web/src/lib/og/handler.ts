import { Template } from "@/lib/flow/templates";
import type { NextApiRequest, NextApiResponse } from "next";
import { renderTemplateImage } from "./templateImage";

type Entry = { image: Buffer | null; version?: number; expires: number };

// Rendered images by kind and ID, so crawlers asking again don't render them again.
// Unknown IDs are kept too, as null.
const cache = new Map<string, Entry>();
const cacheDuration = 15 * 60 * 1000;
const maxCached = 500;

// Responds with the link preview of the template, or the image of the site if
// there is none. Pages link the image with the template's version, so a new
// version gets a new URL. Other versions redirect to the current one instead
// of being rendered, so made up URLs can't make the server render images.
export async function handleTemplateImage(
  req: NextApiRequest,
  res: NextApiResponse,
  kind: "templates" | "bundles",
  id: string,
  load: () => Promise<Template | undefined>
) {
  const path = `/api/og/${kind}/${encodeURIComponent(id)}`;
  const version = String(req.query.v);

  let entry = cache.get(path);
  if (!entry || entry.expires <= Date.now()) {
    entry = await store(path, await load());
  } else if (entry.image && version !== String(entry.version)) {
    // A version the cache doesn't have may be a new one. Loading the template
    // tells, which is cheap compared to rendering it.
    const template = await load();
    if (template?.version !== entry.version) {
      entry = await store(path, template);
    }
  }

  if (!entry.image) {
    res.redirect(302, "/og.png");
    return;
  }
  if (version !== String(entry.version)) {
    res.redirect(302, `${path}?v=${entry.version}`);
    return;
  }
  res.setHeader("Content-Type", "image/png");
  res.send(entry.image);
}

async function store(path: string, template: Template | undefined) {
  const entry: Entry = {
    image: template
      ? Buffer.from(await (await renderTemplateImage(template)).arrayBuffer())
      : null,
    version: template?.version,
    expires: Date.now() + cacheDuration,
  };
  if (!cache.has(path) && cache.size >= maxCached) {
    cache.delete(cache.keys().next().value!);
  }
  cache.set(path, entry);
  return entry;
}
