import { Template } from "@/lib/flow/templates";
import type { NextApiRequest, NextApiResponse } from "next";
import { ogImagePath } from "./path";
import { renderTemplateImage } from "./templateImage";

type Kind = Parameters<typeof ogImagePath>[0];
// OgImage is what a link preview shows, with the version in its URL.
export type OgImage = { version: string; render: () => Promise<Response> };
type Load = () => Promise<OgImage | undefined>;
type Entry = {
  image: Buffer;
  version: string;
  // When the template is loaded again, to drop it when it's gone.
  expires: number;
  // When a request asked for another version last, which may be a new one.
  checked: number;
};

// Rendered images by kind and ID, so crawlers asking again don't render them
// again. Requests that come in while one is loaded wait for it.
const cache = new Map<string, Promise<Entry | null>>();
const cacheDuration = 15 * 60 * 1000;
const checkInterval = 60 * 1000;
const maxCached = 500;

// handleTemplateImage responds with the link preview of a template.
export function handleTemplateImage(
  req: NextApiRequest,
  res: NextApiResponse,
  kind: Kind,
  id: string,
  load: () => Template | undefined | Promise<Template | undefined>
) {
  return handleOgImage(req, res, kind, id, async () => {
    const template = await load();
    return (
      template && {
        version: String(template.version),
        render: () => renderTemplateImage(template),
      }
    );
  });
}

// Responds with the link preview, or the image of the site if there is none.
// Pages link the image with its version. Other versions redirect to the
// current one instead of being rendered, so made up URLs can't make the server
// render images.
export async function handleOgImage(
  req: NextApiRequest,
  res: NextApiResponse,
  kind: Kind,
  id: string,
  load: Load
) {
  const key = `${kind}/${id}`;
  const version = String(req.query.v);

  try {
    let entry = await cache.get(key);
    const now = Date.now();
    const otherVersion = entry && version !== entry.version;
    if (
      !entry ||
      entry.expires <= now ||
      (otherVersion && entry.checked + checkInterval <= now)
    ) {
      if (entry) {
        // Requests coming in meanwhile get the current image.
        entry.expires = now + cacheDuration;
        entry.checked = now;
      }
      entry = await refresh(key, load, entry ?? undefined);
    }

    if (!entry) {
      res.redirect(302, "/og.png");
    } else if (version !== entry.version) {
      res.redirect(302, ogImagePath(kind, id, entry.version) ?? "/og.png");
    } else {
      // Not longer, since the version doesn't change with the name and avatar
      // of users, or the name and description of official templates.
      res.setHeader("Cache-Control", "public, max-age=86400");
      res.setHeader("Content-Type", "image/png");
      res.send(entry.image);
    }
  } catch {
    res.redirect(302, "/og.png");
  }
}

// refresh loads the image and renders it, unless the previous entry is of the
// same version. Unknown IDs aren't kept, so made up ones don't push out real
// images. When loading fails, the previous entry is kept.
function refresh(key: string, load: Load, previous?: Entry) {
  const pending = (async (): Promise<Entry | null> => {
    const og = await load();
    if (!og) return null;
    const now = Date.now();
    if (og.version === previous?.version) {
      return { ...previous, expires: now + cacheDuration, checked: now };
    }
    const image = await og.render();
    return {
      image: Buffer.from(await image.arrayBuffer()),
      version: og.version,
      expires: now + cacheDuration,
      checked: now,
    };
  })();

  if (!cache.has(key) && cache.size >= maxCached) {
    cache.delete(cache.keys().next().value!);
  }
  cache.set(key, pending);
  pending.then(
    (entry) => {
      if (!entry && cache.get(key) === pending) cache.delete(key);
    },
    () => {
      if (cache.get(key) !== pending) return;
      if (previous) cache.set(key, Promise.resolve(previous));
      else cache.delete(key);
    }
  );
  return pending;
}
