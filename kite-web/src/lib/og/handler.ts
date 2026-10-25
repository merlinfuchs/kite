import { Template } from "@/lib/flow/templates";
import type { NextApiRequest, NextApiResponse } from "next";
import { ogImagePath } from "./path";
import { renderTemplateImage } from "./templateImage";

type Kind = Parameters<typeof ogImagePath>[0];
type Entry = { image: Buffer | null; version?: number; expires: number };

// Rendered images by kind and ID, so crawlers asking again don't render them
// again. Unknown IDs are kept too, as null. Requests that come in while an
// image renders wait for it instead of rendering it too.
const cache = new Map<string, Promise<Entry>>();
const cacheDuration = 15 * 60 * 1000;
const maxCached = 500;

// Responds with the link preview of the template, or the image of the site if
// there is none. Pages link the image with the template's version. Other
// versions redirect to the current one instead of being rendered, so made up
// URLs can't make the server render images.
export async function handleTemplateImage(
  req: NextApiRequest,
  res: NextApiResponse,
  kind: Kind,
  id: string,
  load: () => Template | undefined | Promise<Template | undefined>
) {
  const key = `${kind}/${id}`;
  const version = String(req.query.v);

  let entry = await cache.get(key);
  if (!entry || entry.expires <= Date.now()) {
    entry = await store(key, load);
  } else if (entry.image && version !== String(entry.version)) {
    // A version the cache doesn't have may be a new one. Loading the template
    // tells, which is cheap compared to rendering it.
    const template = await load();
    if (template?.version !== entry.version) {
      entry = await store(key, () => template);
    }
  }

  if (!entry.image) {
    res.redirect(302, "/og.png");
  } else if (version !== String(entry.version)) {
    res.redirect(302, ogImagePath(kind, id, entry.version!) ?? "/og.png");
  } else {
    res.setHeader("Content-Type", "image/png");
    res.send(entry.image);
  }
}

function store(
  key: string,
  load: () => Template | undefined | Promise<Template | undefined>
) {
  const entry = (async (): Promise<Entry> => {
    const template = await load();
    const image = template
      ? Buffer.from(await (await renderTemplateImage(template)).arrayBuffer())
      : null;
    return {
      image,
      version: template?.version,
      expires: Date.now() + cacheDuration,
    };
  })();

  if (!cache.has(key) && cache.size >= maxCached) {
    cache.delete(cache.keys().next().value!);
  }
  cache.set(key, entry);
  // Failures, like the API being down, aren't kept.
  entry.catch(() => {
    if (cache.get(key) === entry) cache.delete(key);
  });
  return entry;
}
