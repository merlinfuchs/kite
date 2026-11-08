import { fetchBundles } from "@/lib/apiData";
import { creatorPath, templatePath } from "@/lib/flow/templates";
import { urlSitemap } from "@/lib/sitemap";
import type { NextApiRequest, NextApiResponse } from "next";

// The pages of the templates in the marketplace and of the users with templates
// in it, served as /sitemap-marketplace.xml.
export default async function handler(
  _req: NextApiRequest,
  res: NextApiResponse
) {
  const bundles = await fetchBundles().catch(() => null);
  res.setHeader("Content-Type", "application/xml; charset=utf-8");
  res.send(
    urlSitemap([
      ...(bundles ?? []).map((b) => templatePath(b.id)),
      ...Array.from(
        new Set((bundles ?? []).flatMap((b) => b.owner?.discord_id ?? []))
      ).map(creatorPath),
    ])
  );
}
