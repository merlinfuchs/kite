import { fetchListedBundles } from "@/lib/apiData";
import { templatePaths } from "@/lib/flow/templateFiles";
import { creatorPath, templatePath } from "@/lib/flow/templates";
import { urlSitemap } from "@/lib/sitemap";
import type { NextApiRequest, NextApiResponse } from "next";

// The pages of the marketplace and of the users with templates in it, served
// as /sitemap-marketplace.xml. The official templates come from the files, so
// they're listed even when the API can't be reached.
export default async function handler(
  _req: NextApiRequest,
  res: NextApiResponse
) {
  const listed = await fetchListedBundles().catch(() => null);
  res.setHeader("Content-Type", "application/xml; charset=utf-8");
  res.send(
    urlSitemap([
      ...templatePaths(),
      ...(listed ?? []).map((b) => templatePath(b.id)),
      ...Array.from(
        new Set((listed ?? []).flatMap((b) => b.owner?.discord_id ?? []))
      ).map(creatorPath),
    ])
  );
}
