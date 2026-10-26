import { fetchListedBundles } from "@/lib/apiData";
import { creatorBundles, templateFromBundle } from "@/lib/flow/templates";
import { renderCreatorImage } from "@/lib/og/creatorImage";
import { handleOgImage } from "@/lib/og/handler";
import { creatorImageVersion } from "@/lib/og/path";
import type { NextApiRequest, NextApiResponse } from "next";

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  const creatorId = String(req.query.creatorId);
  return handleOgImage(req, res, "creators", creatorId, async () => {
    const listed = creatorBundles(
      (await fetchListedBundles()) ?? [],
      creatorId
    );
    const owner = listed[0]?.owner;
    if (!owner) return undefined;
    const templates = listed.flatMap((b) => templateFromBundle(b) ?? []);
    return {
      version: creatorImageVersion(templates),
      render: () => renderCreatorImage(owner, templates),
    };
  });
}
